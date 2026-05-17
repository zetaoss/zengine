package facebook

import (
	"encoding/json"
	"net/url"
	"strings"
)

type Provider struct{}

func (Provider) Name() string { return "facebook" }

func (Provider) BuildAuthorizeURL(clientID, redirectURL, state string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("state", state)
	q.Set("scope", "email")
	return "https://www.facebook.com/v19.0/dialog/oauth?" + q.Encode()
}

func (Provider) BuildTokenForm(clientID, clientSecret, redirectURL, code string) url.Values {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", redirectURL)
	form.Set("code", code)
	return form
}

func (Provider) TokenURL() string { return "https://graph.facebook.com/v19.0/oauth/access_token" }

func (Provider) BuildUserRequest(accessToken string) (string, string) {
	u := "https://graph.facebook.com/me?fields=id&access_token=" + url.QueryEscape(accessToken)
	return u, ""
}

func (Provider) ExtractSocialID(body []byte) (string, error) {
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.ID), nil
}

func (Provider) SupportsDeauthorize() bool { return true }
