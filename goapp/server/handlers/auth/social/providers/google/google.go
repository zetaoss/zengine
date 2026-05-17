package google

import (
	"encoding/json"
	"net/url"
	"strings"
)

type Provider struct{}

func (Provider) Name() string { return "google" }

func (Provider) BuildAuthorizeURL(clientID, redirectURL, state string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("state", state)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("access_type", "online")
	return "https://accounts.google.com/o/oauth2/v2/auth?" + q.Encode()
}

func (Provider) BuildTokenForm(clientID, clientSecret, redirectURL, code string) url.Values {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", redirectURL)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	return form
}

func (Provider) TokenURL() string { return "https://oauth2.googleapis.com/token" }

func (Provider) BuildUserRequest(accessToken string) (string, string) {
	return "https://openidconnect.googleapis.com/v1/userinfo", "Bearer " + accessToken
}

func (Provider) ExtractSocialID(body []byte) (string, error) {
	var out struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.Sub), nil
}

func (Provider) SupportsDeauthorize() bool { return false }
