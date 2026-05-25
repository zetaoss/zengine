package auth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/models"
)

const (
	mwUserCachePrefix = "zengine:user:"
	mwUserCacheTTL    = 1 * time.Minute
)

var (
	mwAuthRedisMu sync.RWMutex
	mwAuthRedis   *goredis.Client
)

func FetchUser(r *http.Request, cfg *config.Config) (models.MWUser, bool) {
	cookie := r.Header.Get("Cookie")
	if cookie == "" {
		return models.MWUser{}, false
	}
	if cached, ok := loadMWUserCache(cfg, cookie); ok {
		return cached, true
	}

	apiServer := strings.TrimRight(cfg.App.APIServer, "/")
	if apiServer == "" {
		return models.MWUser{}, false
	}

	req, _ := http.NewRequest(http.MethodGet, apiServer+"/w/api.php?action=query&meta=userinfo&uiprop=groups|blockinfo&format=json", nil)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return models.MWUser{}, false
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return models.MWUser{}, false
	}

	var payload struct {
		Query struct {
			UserInfo struct {
				ID    int      `json:"id"`
				Name  string   `json:"name"`
				Anon  bool     `json:"anon"`
				Group []string `json:"groups"`

				BlockID     *int    `json:"blockid"`
				BlockedByID *int    `json:"blockedbyid"`
				BlockedBy   *string `json:"blockedby"`
				BlockExpiry *string `json:"blockexpiry"`
			} `json:"userinfo"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return models.MWUser{}, false
	}

	ui := payload.Query.UserInfo
	if ui.Anon || ui.ID < 1 {
		return models.MWUser{}, false
	}
	user := models.MWUser{
		ID:          ui.ID,
		Name:        ui.Name,
		Groups:      ui.Group,
		BlockID:     ui.BlockID,
		BlockedByID: ui.BlockedByID,
		BlockedBy:   ui.BlockedBy,
		BlockExpiry: ui.BlockExpiry,
	}
	saveMWUserCache(cfg, cookie, user)
	return user, true
}

func loadMWUserCache(cfg *config.Config, cookie string) (models.MWUser, bool) {
	client := mwAuthRedisClient(cfg)
	if client == nil {
		return models.MWUser{}, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	sum := sha256.Sum256([]byte(cookie))
	key := fmt.Sprintf("%s%x", mwUserCachePrefix, sum[:])
	val, err := client.Get(ctx, key).Result()
	if err != nil || strings.TrimSpace(val) == "" {
		return models.MWUser{}, false
	}
	var user models.MWUser
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return models.MWUser{}, false
	}
	return user, true
}

func saveMWUserCache(cfg *config.Config, cookie string, user models.MWUser) {
	client := mwAuthRedisClient(cfg)
	if client == nil || user.ID < 1 {
		return
	}
	raw, _ := json.Marshal(user)
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	sum := sha256.Sum256([]byte(cookie))
	key := fmt.Sprintf("%s%x", mwUserCachePrefix, sum[:])
	_ = client.Set(ctx, key, string(raw), mwUserCacheTTL).Err()
}

func mwAuthRedisClient(cfg *config.Config) *goredis.Client {
	mwAuthRedisMu.RLock()
	if mwAuthRedis != nil {
		client := mwAuthRedis
		mwAuthRedisMu.RUnlock()
		return client
	}
	mwAuthRedisMu.RUnlock()

	mwAuthRedisMu.Lock()
	defer mwAuthRedisMu.Unlock()
	if mwAuthRedis != nil {
		return mwAuthRedis
	}
	client, err := appredis.Open(cfg)
	if err == nil {
		mwAuthRedis = client
	}
	return mwAuthRedis
}
