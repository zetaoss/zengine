package providers

import (
	"context"
	"net/url"
)

type Provider interface {
	Name() string
	BuildAuthorizeURL(clientID, redirectURL, state string) string
	BuildTokenForm(clientID, clientSecret, redirectURL, code string) url.Values
	TokenURL() string
	BuildUserRequest(accessToken string) (endpoint string, authHeader string)
	ExtractSocialID(body []byte) (string, error)
	SupportsDeauthorize() bool
}

type ctxKey struct{}

func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, true)
}
