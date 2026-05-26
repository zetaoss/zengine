package editbotsvc

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
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

func PublishContent(cfg *config.Config, userID int, title string, requestType string, content string, taskID int) (*PublishResult, error) {
	apiServer := strings.TrimRight(cfg.App.APIServer, "/")
	secret := cfg.App.InternalSecretKey
	if apiServer == "" || secret == "" {
		return nil, fmt.Errorf("missing API_SERVER or INTERNAL_SECRET_KEY")
	}

	restURL := apiServer + "/w/rest.php/editbot/publish"
	payload := map[string]interface{}{
		"secret":       secret,
		"user_id":      userID,
		"title":        title,
		"text":         content,
		"summary":      fmt.Sprintf("Editbot task #%d", taskID),
		"request_type": requestType,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(restURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var result struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Revid   int    `json:"revid"`
		Title   string `json:"title"`
		User    string `json:"user"`
		Errors  []any  `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w (status: %d)", err, resp.StatusCode)
	}

	if result.Status != "success" {
		code := result.Code
		if code == "" {
			code = result.Status
		}
		return nil, &PublishError{Code: code, Info: result.Message}
	}

	slog.Info("[editbot] publish success via REST", "task_id", taskID, "revid", result.Revid, "user", result.User)

	return &PublishResult{
		OK:     true,
		Title:  result.Title,
		Revid:  result.Revid,
		Result: "Success",
	}, nil
}
