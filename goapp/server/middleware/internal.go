package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
)

func RequireInternal(cfg *config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			secret := cfg.App.InternalSecretKey
			timestamp := r.Header.Get("X-Api-Timestamp")
			signature := r.Header.Get("X-Api-Signature")

			if secret == "" || timestamp == "" || signature == "" {
				writeErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			ts, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				writeErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			diff := time.Now().Unix() - ts
			if diff < -300 || diff > 300 {
				writeErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			msg := r.Method + "\n" + r.URL.Path + "\n" + r.URL.RawQuery + "\n" + timestamp
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write([]byte(msg))
			expected := hex.EncodeToString(mac.Sum(nil))

			if !hmac.Equal([]byte(expected), []byte(signature)) {
				writeErrorJSON(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
