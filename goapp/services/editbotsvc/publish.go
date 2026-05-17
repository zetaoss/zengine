package editbotsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
)

type PublishResult struct {
	OK     bool   `json:"ok"`
	Title  string `json:"title"`
	Revid  int    `json:"revid"`
	Result string `json:"result"`
}

type PublishError struct {
	Code string
	Info string
}

func (e *PublishError) Error() string {
	return fmt.Sprintf("MediaWiki edit failed: %s (%s)", e.Info, e.Code)
}

type mwResponse struct {
	Edit *struct {
		Result   string `json:"result"`
		NewRevid int    `json:"newrevid"`
		PageID   int    `json:"pageid"`
		Title    string `json:"title"`
		NoChange bool   `json:"nochange"`
	} `json:"edit"`
	Error *struct {
		Code string `json:"code"`
		Info string `json:"info"`
	} `json:"error"`
}

func PublishContent(cfg *config.Config, title string, requestType string, content string, taskID int) (*PublishResult, error) {
	apiServer := strings.TrimRight(cfg.App.APIServer, "/")
	user := strings.TrimSpace(cfg.EditBot.Username)
	pass := strings.TrimSpace(cfg.EditBot.Password)
	if apiServer == "" || user == "" || pass == "" {
		return nil, fmt.Errorf("missing EDITBOT credentials or apiServer")
	}
	apiURL := apiServer + "/w/api.php"
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Timeout: 30 * time.Second, Jar: jar}

	loginToken, err := fetchMWToken(client, apiURL, "login")
	if err != nil {
		return nil, err
	}
	form := url.Values{}
	form.Set("action", "login")
	form.Set("format", "json")
	form.Set("formatversion", "2")
	form.Set("lgname", user)
	form.Set("lgpassword", pass)
	form.Set("lgtoken", loginToken)
	if _, err := postForm(client, apiURL, form); err != nil {
		return nil, fmt.Errorf("MediaWiki bot login failed: %w", err)
	}

	csrf, err := fetchMWToken(client, apiURL, "csrf")
	if err != nil {
		return nil, err
	}

	edit := url.Values{}
	edit.Set("action", "edit")
	edit.Set("format", "json")
	edit.Set("formatversion", "2")
	edit.Set("title", title)
	edit.Set("text", content)
	edit.Set("summary", fmt.Sprintf("Editbot task #%d", taskID))
	edit.Set("token", csrf)
	edit.Set("bot", "1")
	edit.Set("minor", "1")
	edit.Set("assert", "user")
	edit.Set("assertuser", user)
	if requestType == "create" {
		edit.Set("createonly", "1")
	} else {
		edit.Set("nocreate", "1")
	}

	resp, err := postForm(client, apiURL, edit)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, &PublishError{Code: resp.Error.Code, Info: resp.Error.Info}
	}

	if resp.Edit == nil {
		return nil, fmt.Errorf("MediaWiki edit failed: empty edit node")
	}

	if resp.Edit.Result != "Success" {
		return nil, fmt.Errorf("MediaWiki edit failed: result=%s", resp.Edit.Result)
	}

	slog.Info("[editbot] publish success", "task_id", taskID, "newrevid", resp.Edit.NewRevid, "nochange", resp.Edit.NoChange)

	return &PublishResult{
		OK:     true,
		Title:  resp.Edit.Title,
		Revid:  resp.Edit.NewRevid,
		Result: resp.Edit.Result,
	}, nil
}

func fetchMWToken(client *http.Client, apiURL string, tokenType string) (string, error) {
	u := fmt.Sprintf("%s?action=query&format=json&formatversion=2&meta=tokens&type=%s", apiURL, tokenType)
	resp, err := client.Get(u)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var payload struct {
		Query struct {
			Tokens map[string]string `json:"tokens"`
		} `json:"query"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	tok := payload.Query.Tokens[tokenType+"token"]
	if tok == "" {
		return "", fmt.Errorf("failed to fetch MediaWiki %s token", tokenType)
	}
	return tok, nil
}

func postForm(client *http.Client, apiURL string, form url.Values) (*mwResponse, error) {
	resp, err := client.PostForm(apiURL, form)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var mr mwResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return nil, err
	}
	return &mr, nil
}
