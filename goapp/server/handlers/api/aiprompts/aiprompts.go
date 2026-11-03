package aiprompts

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"github.com/zetaoss/zengine/goapp/tasks/aiedit"
	"gorm.io/gorm"
)

// promptView is a template with the run form and messages it runs with (aiedit.TemplateUserInputs, Messages),
// and the two as the YAML its editor shows (spec, user_inputs_yaml, messages_yaml).
type promptView struct {
	models.AIEditPrompt
	UserInputs     []prompts.Input   `json:"user_inputs"`
	Messages       []prompts.Message `json:"messages"`
	UserInputsYAML string            `json:"user_inputs_yaml"`
	MessagesYAML   string            `json:"messages_yaml"`
	Spec           string            `json:"spec"`
}

func newPromptView(p models.AIEditPrompt) promptView {
	t := prompts.Template{UserInputs: aiedit.TemplateUserInputs(p), Messages: aiedit.TemplateMessages(p)}
	return promptView{
		AIEditPrompt:   p,
		UserInputs:     t.UserInputs,
		Messages:       t.Messages,
		UserInputsYAML: t.UserInputsYAML(),
		MessagesYAML:   t.MessagesYAML(),
		Spec:           t.YAML(),
	}
}

// validateEditorRoles keeps newly saved editor blocks in system-then-user order.
func validateEditorRoles(messages []prompts.Message) error {
	if len(messages) == 0 || messages[0].Role != "system" {
		return errors.New("첫 번째 블록은 system이어야 합니다")
	}
	if messages[0].If != "" || messages[0].Each != "" {
		return errors.New("system 블록에는 조건이나 반복을 설정할 수 없습니다")
	}
	for _, m := range messages[1:] {
		if m.Role != "user" {
			return errors.New("두 번째 블록부터는 user여야 합니다")
		}
	}
	return nil
}

// editorMessages derives user messages from the chain inputs and the template's run form.
func editorMessages(requestType string, inputs []prompts.Input, system string) []prompts.Message {
	messages := []prompts.Message{{Role: "system", Content: system}, {Role: "user", Content: "-----title-----\n{title}"}}
	if requestType == models.AIEditRequestUpdate {
		messages = append(messages, prompts.Message{Role: "user", Content: "-----body-----\n{body}"})
	}
	for _, in := range inputs {
		if in.Name == "useBacklinks" {
			messages = append(messages, prompts.Message{Role: "user", If: "backlinks", Content: "-----backlinks-----\n{backlinks}"})
			continue
		}
		heading := in.Name

		m := prompts.Message{Role: "user"}
		if in.Type == prompts.TypeTextList {
			m.Each = in.Name
			heading += " {index}"
		} else {
			m.If = in.Name
		}
		m.Content = "-----" + heading + "-----\n{" + in.Name + "}"
		messages = append(messages, m)
	}
	return messages
}

func Index(c *serverctx.Context) {
	var rows []models.AIEditPrompt

	if err := c.DB.Order("id DESC").Find(&rows).Error; err != nil {
		c.InternalError()
		return
	}
	views := make([]promptView, len(rows))
	for i, r := range rows {
		views[i] = newPromptView(r)
	}
	c.JSON(views)
}

func Show(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	var row models.AIEditPrompt

	if err := c.DB.Where("id = ?", id).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.NotFound()
			return
		}
		c.InternalError()
		return
	}

	c.JSON(newPromptView(row))
}

func Exists(c *serverctx.Context) {
	title := strings.TrimSpace(c.R.URL.Query().Get("title"))
	excludeID, _ := strconv.Atoi(c.R.URL.Query().Get("exclude_id"))

	if title == "" {
		c.JSON(app.H{"exists": false})
		return
	}

	var count int64
	q := c.DB.Model(&models.AIEditPrompt{}).Where("title = ?", title)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	q.Count(&count)

	c.JSON(app.H{"exists": count > 0})
}

func Store(c *serverctx.Context) {
	var body struct {
		ID           int    `json:"id"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		RequestType  string `json:"request_type"`
		Content      string `json:"content"`
		AutoMessages bool   `json:"auto_messages"`
		SystemPrompt string `json:"system_prompt"`
		// Model is left as it is when omitted; "" runs the template on its chain's model.
		Model *string `json:"model"`
		// Messages, when given, are the template's own messages in place of the chain step's; they are checked
		// against the chain (prompts.Chain.ValidateMessages) and content becomes their text.
		Messages *[]prompts.Message `json:"messages"`
		// UserInputs, when given, are the template's own run form (aiedit.TemplateUserInputs).
		UserInputs *[]prompts.Input `json:"user_inputs"`
		// UserInputsYAML and MessagesYAML, when given, are the run form and messages as separate YAML lists without root keys.
		UserInputsYAML *string `json:"user_inputs_yaml"`
		MessagesYAML   *string `json:"messages_yaml"`
		// Spec, when given, is the run form and messages as YAML (prompts.Template), in place of both.
		Spec *string `json:"spec"`
		// Bump is how an edit that changes the messages moves the version: "minor" (the default) or "major".
		Bump string `json:"bump"`
	}
	if !c.Decode(&body) {
		return
	}
	if body.Model != nil {
		m := strings.TrimSpace(*body.Model)
		if m != "" {
			var err error
			if m, err = config.ValidateModel(m); err != nil {
				c.JSONError(http.StatusUnprocessableEntity, err.Error())
				return
			}
		}
		body.Model = &m
	}
	body.Title = strings.TrimSpace(body.Title)
	if body.Title == "" {
		c.JSONError(http.StatusUnprocessableEntity, "제목은 필수입니다.")
		return
	}
	requestType, ok := models.NormalizeAIEditRequestType(body.RequestType)
	if !ok {
		c.JSONError(http.StatusUnprocessableEntity, "유효하지 않은 요청 유형입니다.")
		return
	}
	body.RequestType = requestType
	if body.UserInputsYAML != nil || body.MessagesYAML != nil {
		if body.UserInputsYAML != nil {
			u, err := prompts.ParseUserInputsYAML(*body.UserInputsYAML)
			if err != nil {
				c.JSONError(http.StatusUnprocessableEntity, err.Error())
				return
			}
			body.UserInputs = &u
		}
		if body.MessagesYAML != nil {
			m, err := prompts.ParseMessagesYAML(*body.MessagesYAML)
			if err != nil {
				c.JSONError(http.StatusUnprocessableEntity, err.Error())
				return
			}
			body.Messages = &m
		}
	} else if body.Spec != nil {
		t, err := prompts.ParseTemplate(*body.Spec)
		if err != nil {
			c.JSONError(http.StatusUnprocessableEntity, err.Error())
			return
		}
		body.UserInputs, body.Messages = &t.UserInputs, &t.Messages
	}
	if body.UserInputs != nil {
		picked, err := aiedit.PickUserInputs(requestType, *body.UserInputs)
		if err != nil {
			c.JSONError(http.StatusUnprocessableEntity, err.Error())
			return
		}
		body.UserInputs = &picked
	}
	if body.AutoMessages {
		if body.UserInputs == nil {
			c.JSONError(http.StatusUnprocessableEntity, "사용자 입력 스펙이 필요합니다.")
			return
		}
		messages := editorMessages(requestType, *body.UserInputs, body.SystemPrompt)
		body.Messages = &messages
	}
	var messagesJSON, userInputsJSON string
	if body.Messages != nil || body.UserInputs != nil {
		draft := models.AIEditPrompt{RequestType: requestType, Content: body.Content}
		if body.ID > 0 {
			_ = c.DB.Where("id = ?", body.ID).Take(&draft).Error
			draft.RequestType = requestType
		}
		if body.UserInputs != nil {
			b, _ := json.Marshal(aiedit.UserInputNames(*body.UserInputs))
			userInputsJSON = string(b)
			draft.UserInputs = userInputsJSON
		}
		if body.Messages != nil {
			b, _ := json.Marshal(*body.Messages)
			messagesJSON = string(b)
			draft.Messages = messagesJSON
			body.Content = aiedit.MessagesText(*body.Messages)
		}
		// Messages given are checked as they are: the fallback of a template without messages (its content) must
		// not stand in for an empty list that would be saved.
		messages := aiedit.TemplateMessages(draft)
		if body.Messages != nil {
			messages = *body.Messages
		}
		chain, err := aiedit.ChainOf(requestType).WithUserInputs(aiedit.TemplateUserInputs(draft))
		// The editor's system-then-user order holds for the messages it builds; messages given as they are (a
		// skin's clone of a template saved before that order) run as they did.
		if err == nil && body.AutoMessages {
			err = validateEditorRoles(messages)
		}
		if err == nil {
			err = chain.ValidateMessages(messages)
		}
		if err != nil {
			c.JSONError(http.StatusUnprocessableEntity, err.Error())
			return
		}
	}

	user, _ := c.User()

	now := time.Now().Format("2006-01-02 15:04:05")
	row := models.AIEditPrompt{
		ID:          body.ID,
		Title:       body.Title,
		Description: strings.TrimSpace(body.Description),
		RequestType: body.RequestType,
		Content:     body.Content,
		UpdatedAt:   now,
	}
	columns := []string{"title", "description", "request_type", "content", "version", "updated_at"}
	if body.Messages != nil {
		row.Messages = messagesJSON
		columns = append(columns, "messages")
	}
	if body.UserInputs != nil {
		row.UserInputs = userInputsJSON
		columns = append(columns, "user_inputs")
	}
	if body.Model != nil {
		row.Model = *body.Model
		columns = append(columns, "model")
	}

	if row.ID > 0 {
		var existing models.AIEditPrompt
		if err := c.DB.Where("id = ?", row.ID).First(&existing).Error; err != nil {
			c.InternalError()
			return
		}
		if existing.UserID != user.ID {
			c.JSONError(http.StatusForbidden, "본인만 편집할 수 있습니다.")
			return
		}
		row.Version = existing.Version
		changed := row.Content != existing.Content
		if body.Messages != nil {
			before, _ := json.Marshal(aiedit.TemplateMessages(existing))
			changed = messagesJSON != string(before)
		}
		if body.UserInputs != nil {
			before, _ := json.Marshal(aiedit.UserInputNames(aiedit.TemplateUserInputs(existing)))
			changed = changed || userInputsJSON != string(before)
		}
		if changed {
			row.Version = prompts.NextVersion(existing.Version, body.Bump == "major")
		}

		var count int64
		c.DB.Model(&models.AIEditPrompt{}).Where("title = ? AND id != ?", row.Title, row.ID).Count(&count)
		if count > 0 {
			c.JSONError(http.StatusConflict, "이미 사용 중인 제목입니다.")
			return
		}

		if err := c.DB.Model(&models.AIEditPrompt{}).Where("id = ?", row.ID).Select(columns).Updates(&row).Error; err != nil {
			c.InternalError()
			return
		}
	} else {
		row.UserID = user.ID
		row.UserName = user.Name
		row.CreatedAt = now
		row.Version = "v1.0"

		var count int64
		c.DB.Model(&models.AIEditPrompt{}).Where("title = ?", row.Title).Count(&count)
		if count > 0 {
			c.JSONError(http.StatusConflict, "이미 사용 중인 제목입니다.")
			return
		}

		if err := c.DB.Create(&row).Error; err != nil {
			c.InternalError()
			return
		}
	}
	c.JSON(newPromptView(row))
}

// Check reads {"request_type", "user_inputs_yaml", "messages_yaml"} or {"request_type", "spec"} as Store would
// and returns the run form and messages it holds, or the error that would stop it from saving, without storing anything.
func Check(c *serverctx.Context) {
	var body struct {
		RequestType      string             `json:"request_type"`
		UserInputsYAML   *string            `json:"user_inputs_yaml"`
		MessagesYAML     *string            `json:"messages_yaml"`
		UserInputs       *[]prompts.Input   `json:"user_inputs"`
		Messages         *[]prompts.Message `json:"messages"`
		Spec             string             `json:"spec"`
		PreviewValues    map[string]any     `json:"preview_values"`
		PreviewTitle     *string            `json:"preview_title"`
		PreviewBody      *string            `json:"preview_body"`
		PreviewBacklinks *[]string          `json:"preview_backlinks"`
		AutoMessages     bool               `json:"auto_messages"`
		SystemPrompt     string             `json:"system_prompt"`
		// Stored, with messages, previews them as a template runs them (aiedit.TemplateMessages): they are rendered
		// as they are, not checked as an editor's would be (system first, every variable declared).
		Stored bool `json:"stored"`
	}
	if !c.Decode(&body) {
		return
	}
	requestType, ok := models.NormalizeAIEditRequestType(body.RequestType)
	if !ok {
		c.JSONError(http.StatusUnprocessableEntity, "유효하지 않은 요청 유형입니다.")
		return
	}

	var userInputs []prompts.Input
	var messages []prompts.Message
	var err error

	if body.UserInputs != nil {
		userInputs = *body.UserInputs
	} else if body.UserInputsYAML != nil {
		userInputs, err = prompts.ParseUserInputsYAML(*body.UserInputsYAML)
	}

	if err == nil {
		if body.Messages != nil {
			messages = *body.Messages
		} else if body.MessagesYAML != nil {
			messages, err = prompts.ParseMessagesYAML(*body.MessagesYAML)
		} else if body.UserInputs == nil && body.UserInputsYAML == nil {
			var t prompts.Template
			t, err = prompts.ParseTemplate(body.Spec)
			userInputs = t.UserInputs
			messages = t.Messages
		}
	}

	if err == nil {
		userInputs, err = aiedit.PickUserInputs(requestType, userInputs)
	}
	if err == nil && body.AutoMessages {
		messages = editorMessages(requestType, userInputs, body.SystemPrompt)
	}
	var chain prompts.Chain
	if err == nil && !body.Stored {
		err = validateEditorRoles(messages)
	}
	if err == nil {
		if chain, err = aiedit.ChainOf(requestType).WithUserInputs(userInputs); err == nil && !body.Stored {
			err = chain.ValidateMessages(messages)
		}
	}
	if err != nil {
		c.JSON(app.H{"ok": false, "error": err.Error()})
		return
	}
	response := app.H{"ok": true, "user_inputs": userInputs, "messages": messages}
	if body.PreviewValues != nil {
		values := make(map[string]any)
		for _, in := range userInputs {
			if value, ok := body.PreviewValues[in.Name]; ok {
				values[in.Name] = value
			}
		}
		vars, previewErr := chain.UserVars(values)
		if previewErr == nil {
			vars["title"] = "Hello World"
			vars["body"] = "문서 시작했다\n문서 끝났다"
			if body.PreviewTitle != nil {
				vars["title"] = *body.PreviewTitle
			}
			if body.PreviewBody != nil {
				vars["body"] = *body.PreviewBody
			}
			backlinks := []string{"C language", "Programming language"}
			if body.PreviewBacklinks != nil {
				backlinks = nil
				for _, title := range *body.PreviewBacklinks {
					if title = strings.TrimSpace(title); title != "" {
						backlinks = append(backlinks, title)
					}
					if len(backlinks) == 10 {
						break
					}
				}
			}
			vars, previewErr = chain.RunFunctions(c.R.Context(), vars, map[string]prompts.Func{
				"wiki/backlinkTitles": func(_ context.Context, _ prompts.Vars) (any, error) {
					return backlinks, nil
				},
			})
		}
		if previewErr != nil {
			response["preview_error"] = previewErr.Error()
		} else {
			response["rendered_messages"] = prompts.MergeUserMessages((prompts.Prompt{Messages: messages}).Render(chain.Fill(vars)))
		}
	}
	c.JSON(response)
}

func Destroy(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok {
		c.NotFound()
		return
	}
	if err := c.DB.Where("id = ?", id).Delete(&models.AIEditPrompt{}).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(app.H{"ok": true})
}
