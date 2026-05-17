package social

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/server/handlers/auth/social/providers"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

const (
	socialJoinTTL = 600
	mwBridgeTTL   = 60
)

var returntoRE = regexp.MustCompile(`^[\pL\pN _:\-/().%]+$`)

type oauthCookie struct {
	Provider string `json:"provider"`
	State    string `json:"state"`
	Returnto string `json:"returnto"`
}

func Redirect(c *serverctx.Context) {
	provider := strings.TrimSpace(c.R.PathValue("provider"))
	p, ok := providers.Get(provider)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	cfg, ok := providerConfig(c.Cfg, provider)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}

	returnto := sanitizeReturnto(c.R.URL.Query().Get("returnto"))
	state, err := randomHex(16)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	payload := oauthCookie{Provider: provider, State: state, Returnto: returnto}
	raw, err := json.Marshal(payload)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	cookieVal := base64.RawURLEncoding.EncodeToString(raw)
	http.SetCookie(c.W, &http.Cookie{
		Name:     "social_oauth",
		Value:    cookieVal,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   600,
		SameSite: http.SameSiteLaxMode,
	})

	authURL := p.BuildAuthorizeURL(cfg.ClientID, cfg.RedirectURL, state)
	http.Redirect(c.W, c.R, authURL, http.StatusFound)
}

func Callback(c *serverctx.Context) {
	provider := strings.TrimSpace(c.R.PathValue("provider"))
	p, ok := providers.Get(provider)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	cfg, ok := providerConfig(c.Cfg, provider)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}

	cookie, err := c.R.Cookie("social_oauth")
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		redirectLoginError(c, "social_auth_failed")
		return
	}
	clearCookie(c.W)

	raw, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		redirectLoginError(c, "social_auth_failed")
		return
	}
	var st oauthCookie
	if err := json.Unmarshal(raw, &st); err != nil {
		redirectLoginError(c, "social_auth_failed")
		return
	}
	if st.Provider != provider || st.State == "" || c.R.URL.Query().Get("state") != st.State {
		redirectLoginError(c, "social_auth_failed")
		return
	}

	code := strings.TrimSpace(c.R.URL.Query().Get("code"))
	if code == "" {
		redirectLoginError(c, "social_auth_failed")
		return
	}

	accessToken, err := exchangeToken(c.R.Context(), p, cfg, code)
	if err != nil || accessToken == "" {
		redirectLoginError(c, "social_auth_failed")
		return
	}

	socialID, err := fetchSocialID(c.R.Context(), p, accessToken)
	if err != nil {
		redirectLoginError(c, "social_auth_failed")
		return
	}
	if socialID == "" {
		redirectLoginError(c, "invalid_social_id")
		return
	}

	row, err := findOrCreateUserSocial(c.DB, provider, socialID)
	if err != nil || row.ID < 1 {
		redirectLoginError(c, "social_link_failed")
		return
	}

	if row.UserID.Valid && row.UserID.Int64 > 0 {
		token, err := putToken(c.Cfg, "mwbridge", app.H{
			"user_id":  row.UserID.Int64,
			"returnto": st.Returnto,
		}, mwBridgeTTL)
		if err != nil {
			http.Error(c.W, "internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(c.W, c.R, "/w/rest.php/social/bridge?token="+url.QueryEscape(token), http.StatusFound)
		return
	}

	token, err := putToken(c.Cfg, "socialjoin", app.H{
		"user_social_id": row.ID,
		"provider":       provider,
		"social_id":      socialID,
		"returnto":       st.Returnto,
	}, socialJoinTTL)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(c.W, c.R, "/social-join/"+token, http.StatusFound)
}

func Deauthorize(c *serverctx.Context) {
	provider := strings.TrimSpace(c.R.PathValue("provider"))
	if provider != "facebook" {
		http.NotFound(c.W, c.R)
		return
	}

	socialID, ok := extractFacebookSocialID(c)
	if !ok {
		return
	}

	err := c.DB.Transaction(func(tx *gorm.DB) error {
		row, err := lockUserSocial(tx, provider, socialID)
		if err != nil {
			return err
		}
		if row == nil {
			return nil
		}
		res := tx.Table("zetawiki.user_social").Where("id = ?", row.ID).Update("deauthorized_at", time.Now())
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			return rotateUserToken(tx, row.UserID)
		}
		return nil
	})
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	c.JSON(app.H{"status": "ok"})
}

func Deletion(c *serverctx.Context) {
	provider := strings.TrimSpace(c.R.PathValue("provider"))
	if provider != "facebook" {
		http.NotFound(c.W, c.R)
		return
	}

	socialID, ok := extractFacebookSocialID(c)
	if !ok {
		return
	}
	confirmationCode, err := randomHex(16)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	err = c.DB.Transaction(func(tx *gorm.DB) error {
		row, err := lockUserSocial(tx, provider, socialID)
		if err != nil {
			return err
		}
		if row == nil {
			return nil
		}
		res := tx.Table("zetawiki.user_social").Where("id = ?", row.ID).Updates(app.H{
			"social_id":     nil,
			"deletion_code": confirmationCode,
			"deleted_at":    time.Now(),
		})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected > 0 {
			return rotateUserToken(tx, row.UserID)
		}
		return nil
	})
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	c.JSON(app.H{
		"url":               c.Cfg.App.AppURL + "/auth/deletion/" + provider + "/status/" + confirmationCode,
		"confirmation_code": confirmationCode,
	})
}

func DeletionStatus(c *serverctx.Context) {
	provider := strings.TrimSpace(c.R.PathValue("provider"))
	if provider != "facebook" {
		http.NotFound(c.W, c.R)
		return
	}
	code := strings.TrimSpace(c.R.PathValue("code"))
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(code) {
		http.NotFound(c.W, c.R)
		return
	}

	var row struct {
		ID int `gorm:"column:id"`
	}
	err := c.DB.Table("zetawiki.user_social").Select("id").Where("provider = ? AND deletion_code = ?", "facebook", code).Take(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			http.NotFound(c.W, c.R)
			return
		}
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}

	c.JSON(app.H{
		"status":            "completed",
		"confirmation_code": code,
		"deleted_links":     1,
	})
}

type providerCred struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func providerConfig(cfg *config.Config, provider string) (providerCred, bool) {
	appURL := strings.TrimRight(strings.TrimSpace(cfg.App.AppURL), "/")
	if appURL == "" {
		appURL = "http://localhost"
	}
	switch provider {
	case "facebook":
		return providerCred{cfg.OAuth.FacebookClientID, cfg.OAuth.FacebookClientSecret, appURL + "/auth/callback/facebook"}, true
	case "github":
		return providerCred{cfg.OAuth.GithubClientID, cfg.OAuth.GithubClientSecret, appURL + "/auth/callback/github"}, true
	case "google":
		return providerCred{cfg.OAuth.GoogleClientID, cfg.OAuth.GoogleClientSecret, appURL + "/auth/callback/google"}, true
	default:
		return providerCred{}, false
	}
}

func exchangeToken(ctx any, provider providers.Provider, cfg providerCred, code string) (string, error) {
	contextValue, ok := ctx.(interface{ Done() <-chan struct{} })
	if !ok {
		return "", fmt.Errorf("invalid context")
	}
	_ = contextValue

	form := provider.BuildTokenForm(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURL, code)

	req, err := http.NewRequest(http.MethodPost, provider.TokenURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("oauth token status %d", resp.StatusCode)
	}

	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.AccessToken), nil
}

func fetchSocialID(ctx any, provider providers.Provider, accessToken string) (string, error) {
	_ = ctx
	endpoint, authHeader := provider.BuildUserRequest(accessToken)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("social user status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return provider.ExtractSocialID(body)
}

type userSocialRow struct {
	ID     int           `gorm:"column:id"`
	UserID sql.NullInt64 `gorm:"column:user_id"`
}

func findOrCreateUserSocial(db *gorm.DB, provider, socialID string) (userSocialRow, error) {
	var row userSocialRow
	err := db.Table("zetawiki.user_social").
		Select("id", "user_id").
		Where("provider = ? AND social_id = ?", provider, socialID).
		Take(&row).Error
	if err == nil {
		return row, nil
	}
	if err != gorm.ErrRecordNotFound {
		return userSocialRow{}, err
	}

	_ = db.Table("zetawiki.user_social").Create(app.H{
		"provider":  provider,
		"social_id": socialID,
		"user_id":   nil,
	}).Error

	err = db.Table("zetawiki.user_social").
		Select("id", "user_id").
		Where("provider = ? AND social_id = ?", provider, socialID).
		Take(&row).Error
	return row, err
}

type lockedUserSocial struct {
	ID     int  `gorm:"column:id"`
	UserID *int `gorm:"column:user_id"`
}

func lockUserSocial(db *gorm.DB, provider, socialID string) (*lockedUserSocial, error) {
	var row lockedUserSocial
	err := db.Raw(`SELECT id, user_id FROM zetawiki.user_social WHERE provider = ? AND social_id = ? LIMIT 1 FOR UPDATE`, provider, socialID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID < 1 {
		return nil, nil
	}
	return &row, nil
}

func rotateUserToken(db *gorm.DB, userID *int) error {
	if userID == nil || *userID < 1 {
		return nil
	}
	token, err := randomHex(16)
	if err != nil {
		return err
	}
	return db.Table("zetawiki.user").Where("user_id = ?", *userID).Update("user_token", token).Error
}

func extractFacebookSocialID(c *serverctx.Context) (string, bool) {
	if err := c.R.ParseForm(); err != nil {
		c.JSONStatus(http.StatusBadRequest, app.H{"error": "invalid_signed_request"})
		return "", false
	}
	signedRequest := strings.TrimSpace(c.R.FormValue("signed_request"))
	payload := parseFacebookSignedRequest(signedRequest, c.Cfg.OAuth.FacebookClientSecret)
	if payload == nil {
		c.JSONStatus(http.StatusBadRequest, app.H{"error": "invalid_signed_request"})
		return "", false
	}
	socialID := strings.TrimSpace(fmt.Sprint(payload["user_id"]))
	if socialID == "" || socialID == "<nil>" {
		c.JSONStatus(http.StatusBadRequest, app.H{"error": "invalid_social_id"})
		return "", false
	}
	return socialID, true
}

func parseFacebookSignedRequest(signedRequest, appSecret string) app.H {
	if signedRequest == "" || appSecret == "" || !strings.Contains(signedRequest, ".") {
		return nil
	}
	parts := strings.SplitN(signedRequest, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	sig, err := base64URLDecode(parts[0])
	if err != nil || len(sig) == 0 {
		return nil
	}
	payloadRaw, err := base64URLDecode(parts[1])
	if err != nil || len(payloadRaw) == 0 {
		return nil
	}

	var payload app.H
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return nil
	}
	alg := strings.ToUpper(strings.TrimSpace(fmt.Sprint(payload["algorithm"])))
	if alg != "HMAC-SHA256" {
		return nil
	}
	expected := hmac.New(sha256.New, []byte(appSecret))
	_, _ = expected.Write([]byte(parts[1]))
	if !hmac.Equal(expected.Sum(nil), sig) {
		return nil
	}
	return payload
}

func base64URLDecode(v string) ([]byte, error) {
	if m := len(v) % 4; m > 0 {
		v += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(v)
}

func sanitizeReturnto(returnto string) string {
	returnto = strings.TrimSpace(returnto)
	if returnto == "" || len(returnto) > 200 {
		return ""
	}
	if strings.Contains(returnto, "://") || strings.Contains(returnto, `\\`) || strings.Contains(returnto, "..") {
		return ""
	}
	if !returntoRE.MatchString(returnto) {
		return ""
	}
	return returnto
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func putToken(cfg *config.Config, prefix string, payload app.H, ttlSeconds int) (string, error) {
	token, err := randomHex(32)
	if err != nil {
		return "", err
	}
	key := prefix + ":" + token
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	client, err := appredis.Open(cfg)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Set(ctx, key, string(raw), time.Duration(ttlSeconds)*time.Second).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func redirectLoginError(c *serverctx.Context, code string) {
	http.Redirect(c.W, c.R, "/login?error="+url.QueryEscape(code), http.StatusFound)
}

func clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "social_oauth",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}
