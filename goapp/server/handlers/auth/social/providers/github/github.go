package github

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Provider struct{}

func (Provider) Name() string { return "github" }

func (Provider) BuildAuthorizeURL(clientID, redirectURL, state string) string {
	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("state", state)
	q.Set("scope", "read:user user:email")
	return "https://github.com/login/oauth/authorize?" + q.Encode()
}

func (Provider) BuildTokenForm(clientID, clientSecret, redirectURL, code string) url.Values {
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", redirectURL)
	form.Set("code", code)
	return form
}

func (Provider) TokenURL() string { return "https://github.com/login/oauth/access_token" }

func (Provider) BuildUserRequest(accessToken string) (string, string) {
	return "https://api.github.com/user", "Bearer " + accessToken
}

func (Provider) ExtractSocialID(body []byte) (string, error) {
	var out struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if out.ID < 1 {
		return "", nil
	}
	return fmt.Sprintf("%d", out.ID), nil
}

func (Provider) SupportsDeauthorize() bool { return false }
