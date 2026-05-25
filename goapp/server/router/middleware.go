package router

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/auth"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"github.com/zetaoss/zengine/goapp/server/throttle"

	"gorm.io/gorm"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, m ...Middleware) http.Handler {
	wrapped := h
	for i := len(m) - 1; i >= 0; i-- {
		wrapped = m[i](wrapped)
	}
	return wrapped
}

// Middleware Factories

func (r *Router) WithUser() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user, ok := auth.FetchUser(req, r.serverCtx.Cfg)
			if ok {
				ctx := context.WithValue(req.Context(), serverctx.MWUserKey, user)
				req = req.WithContext(ctx)
			}
			next.ServeHTTP(w, req)
		})
	}
}

func (r *Router) User() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user, ok := auth.FetchUser(req, r.serverCtx.Cfg)
			if !ok {
				r.serverCtx.JSONErrorStatus(w, http.StatusUnauthorized, "Unauthenticated")
				return
			}
			ctx := context.WithValue(req.Context(), serverctx.MWUserKey, user)
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

func (r *Router) Unblocked() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user, ok := req.Context().Value(serverctx.MWUserKey).(models.MWUser)
			if !ok {
				fetched, authOK := auth.FetchUser(req, r.serverCtx.Cfg)
				if !authOK {
					r.serverCtx.JSONErrorStatus(w, http.StatusUnauthorized, "Unauthenticated")
					return
				}
				user = fetched
				req = req.WithContext(context.WithValue(req.Context(), serverctx.MWUserKey, user))
			}
			if user.IsBlocked() {
				r.serverCtx.JSONErrorStatus(w, http.StatusForbidden, "Forbidden")
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

func (r *Router) Sysop() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user, ok := req.Context().Value(serverctx.MWUserKey).(models.MWUser)
			if !ok {
				fetched, authOK := auth.FetchUser(req, r.serverCtx.Cfg)
				if !authOK {
					r.serverCtx.JSONErrorStatus(w, http.StatusUnauthorized, "Unauthenticated")
					return
				}
				user = fetched
				req = req.WithContext(context.WithValue(req.Context(), serverctx.MWUserKey, user))
			}
			if !user.IsSysop() {
				r.serverCtx.JSONErrorStatus(w, http.StatusForbidden, "Forbidden")
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

func (r *Router) Owner(model models.Model, idKey ...string) Middleware {
	key := "id"
	if len(idKey) > 0 {
		key = idKey[0]
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user, ok := req.Context().Value(serverctx.MWUserKey).(models.MWUser)
			if !ok {
				fetched, authOK := auth.FetchUser(req, r.serverCtx.Cfg)
				if !authOK {
					r.serverCtx.JSONErrorStatus(w, http.StatusUnauthorized, "Unauthenticated")
					return
				}
				user = fetched
				req = req.WithContext(context.WithValue(req.Context(), serverctx.MWUserKey, user))
			}

			idStr := req.PathValue(key)
			if idStr == "" {
				r.serverCtx.JSONErrorStatus(w, http.StatusNotFound, "Resource ID missing")
				return
			}

			var ownerID int
			if err := r.serverCtx.DB.WithContext(req.Context()).Table(model.TableName()).Select("user_id").Where("id = ?", idStr).Take(&ownerID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					r.serverCtx.JSONErrorStatus(w, http.StatusNotFound, "Resource not found")
				} else {
					r.serverCtx.JSONErrorStatus(w, http.StatusInternalServerError, "Internal server error")
				}
				return
			}

			if user.ID != ownerID {
				r.serverCtx.JSONErrorStatus(w, http.StatusForbidden, "Forbidden")
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

func (r *Router) OwnerOrSysop(model models.Model, idKey ...string) Middleware {
	key := "id"
	if len(idKey) > 0 {
		key = idKey[0]
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			user, ok := req.Context().Value(serverctx.MWUserKey).(models.MWUser)
			if !ok {
				fetched, authOK := auth.FetchUser(req, r.serverCtx.Cfg)
				if !authOK {
					r.serverCtx.JSONErrorStatus(w, http.StatusUnauthorized, "Unauthenticated")
					return
				}
				user = fetched
				req = req.WithContext(context.WithValue(req.Context(), serverctx.MWUserKey, user))
			}

			if user.IsSysop() {
				next.ServeHTTP(w, req)
				return
			}

			idStr := req.PathValue(key)
			if idStr == "" {
				r.serverCtx.JSONErrorStatus(w, http.StatusNotFound, "Resource ID missing")
				return
			}

			var ownerID int
			if err := r.serverCtx.DB.WithContext(req.Context()).Table(model.TableName()).Select("user_id").Where("id = ?", idStr).Take(&ownerID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					r.serverCtx.JSONErrorStatus(w, http.StatusNotFound, "Resource not found")
				} else {
					r.serverCtx.JSONErrorStatus(w, http.StatusInternalServerError, "Internal server error")
				}
				return
			}

			if user.ID != ownerID {
				r.serverCtx.JSONErrorStatus(w, http.StatusForbidden, "Forbidden")
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

func (r *Router) Internal() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			secret := r.serverCtx.Cfg.App.InternalSecretKey
			timestamp := req.Header.Get("X-Api-Timestamp")
			signature := req.Header.Get("X-Api-Signature")

			if secret == "" || timestamp == "" || signature == "" {
				r.serverCtx.JSONErrorStatus(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			ts, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				r.serverCtx.JSONErrorStatus(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			diff := time.Now().Unix() - ts
			if diff < -300 || diff > 300 {
				r.serverCtx.JSONErrorStatus(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			msg := req.Method + "\n" + req.URL.Path + "\n" + req.URL.RawQuery + "\n" + timestamp
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write([]byte(msg))
			expected := hex.EncodeToString(mac.Sum(nil))

			if !hmac.Equal([]byte(expected), []byte(signature)) {
				r.serverCtx.JSONErrorStatus(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

func (r *Router) Throttle(limit int, window time.Duration, namespace string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			allowed, retryAfter := throttle.Allow(req, r.serverCtx.Cfg, limit, window, namespace)
			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				r.serverCtx.JSONErrorStatus(w, http.StatusTooManyRequests, "Too Many Requests")
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}
