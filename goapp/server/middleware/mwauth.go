package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"net/http"
	"strings"
	"sync"
	"time"
)

type contextKey string

const mwUserKey contextKey = "mw_user"

const (
	mwUserCachePrefix = "zengine:mwauth:user:v1:"
	mwUserCacheTTL    = 10 * time.Second
)

var (
	mwAuthRedisOnce sync.Once
	mwAuthRedis     *goredis.Client
)

type MWUser struct {
	ID          int
	Name        string
	Groups      []string
	BlockID     *int
	BlockedByID *int
	BlockedBy   *string
	BlockExpiry *string
}

func (u MWUser) IsSysop() bool {
	for _, g := range u.Groups {
		if g == "sysop" {
			return true
		}
	}
	return false
}

func (u MWUser) IsBlocked() bool {
	return u.BlockID != nil || u.BlockedByID != nil || u.BlockedBy != nil || u.BlockExpiry != nil
}

func UserFromRequest(r *http.Request) (MWUser, bool) {
	u, ok := r.Context().Value(mwUserKey).(MWUser)
	return u, ok
}

func withMWUser(ctx context.Context, user MWUser) context.Context {
	return context.WithValue(ctx, mwUserKey, user)
}

func RequireLoggedIn(cfg *config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := fetchMWUser(r, cfg)
			if !ok {
				writeErrorJSON(w, http.StatusUnauthorized, "Unauthenticated")
				return
			}
			ctx := withMWUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func MaybeLoggedIn(cfg *config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := fetchMWUser(r, cfg)
			if ok {
				ctx := withMWUser(r.Context(), user)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireUnblocked(cfg *config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromRequest(r)
			if !ok {
				fetched, authOK := fetchMWUser(r, cfg)
				if !authOK {
					writeErrorJSON(w, http.StatusUnauthorized, "Unauthenticated")
					return
				}
				user = fetched
				ctx := withMWUser(r.Context(), user)
				r = r.WithContext(ctx)
			}
			if user.IsBlocked() {
				writeErrorJSON(w, http.StatusForbidden, "Forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func fetchMWUser(r *http.Request, cfg *config.Config) (MWUser, bool) {
	cookie := r.Header.Get("Cookie")
	if cookie == "" {
		return MWUser{}, false
	}
	if cached, ok := loadMWUserCache(cfg, cookie); ok {
		return cached, true
	}

	apiServer := strings.TrimRight(cfg.App.APIServer, "/")
	if apiServer == "" {
		return MWUser{}, false
	}

	req, err := http.NewRequest(http.MethodGet, apiServer+"/w/api.php?action=query&meta=userinfo&uiprop=groups|blockinfo&format=json", nil)
	if err != nil {
		return MWUser{}, false
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return MWUser{}, false
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusOK {
		return MWUser{}, false
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
		return MWUser{}, false
	}

	ui := payload.Query.UserInfo
	if ui.Anon || ui.ID < 1 {
		return MWUser{}, false
	}
	user := MWUser{
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

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func loadMWUserCache(cfg *config.Config, cookie string) (MWUser, bool) {
	client := mwAuthRedisClient(cfg)
	if client == nil {
		return MWUser{}, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	val, err := client.Get(ctx, mwUserCacheKey(cookie)).Result()
	if err != nil || strings.TrimSpace(val) == "" {
		return MWUser{}, false
	}
	var user MWUser
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return MWUser{}, false
	}
	if user.ID < 1 {
		return MWUser{}, false
	}
	return user, true
}

func saveMWUserCache(cfg *config.Config, cookie string, user MWUser) {
	client := mwAuthRedisClient(cfg)
	if client == nil || user.ID < 1 {
		return
	}
	raw, err := json.Marshal(user)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	_ = client.Set(ctx, mwUserCacheKey(cookie), string(raw), mwUserCacheTTL).Err()
}

func mwAuthRedisClient(cfg *config.Config) *goredis.Client {
	mwAuthRedisOnce.Do(func() {
		if cfg == nil {
			return
		}
		client, err := appredis.Open(cfg)
		if err != nil {
			return
		}
		mwAuthRedis = client
	})
	return mwAuthRedis
}

func mwUserCacheKey(cookie string) string {
	sum := sha256.Sum256([]byte(cookie))
	return fmt.Sprintf("%s%x", mwUserCachePrefix, sum[:])
}
