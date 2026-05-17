package middleware

import (
	"github.com/zetaoss/zengine/goapp/app/config"
	"net/http"
)

func RequireSysop(cfg *config.Config) Middleware {
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
				r = r.WithContext(withMWUser(r.Context(), user))
			}
			if !user.IsSysop() {
				writeErrorJSON(w, http.StatusForbidden, "Forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
