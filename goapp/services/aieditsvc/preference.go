package aieditsvc

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zetaoss/zengine/goapp/app/config"
)

func SetAiEditAsMine(cfg *config.Config, userID int, enabled bool) error {
	apiServer := strings.TrimRight(cfg.App.APIServer, "/")
	secret := cfg.App.InternalSecretKey
	if apiServer == "" || secret == "" {
		return fmt.Errorf("missing API_SERVER or INTERNAL_SECRET_KEY")
	}

	restURL := apiServer + "/w/rest.php/ai-edit/preference"
	payload := map[string]any{
		"secret":         secret,
		"user_id":        userID,
		"enable_ai_edit": enabled,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := httpClient.Post(restURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var result struct {
		Status  string `json:"status"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w (status: %d)", err, resp.StatusCode)
	}

	if result.Status != "success" {
		code := result.Code
		if code == "" {
			code = result.Status
		}
		return fmt.Errorf("ai-edit preference update failed: %s (%s)", result.Message, code)
	}

	return nil
}
