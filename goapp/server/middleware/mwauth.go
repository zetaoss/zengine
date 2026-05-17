package middleware

import (
	"context"
	"encoding/json"
	"github.com/zetaoss/zengine/goapp/app/config"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const mwUserKey contextKey = "mw_user"

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
	return MWUser{
		ID:          ui.ID,
		Name:        ui.Name,
		Groups:      ui.Group,
		BlockID:     ui.BlockID,
		BlockedByID: ui.BlockedByID,
		BlockedBy:   ui.BlockedBy,
		BlockExpiry: ui.BlockExpiry,
	}, true
}

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
