package aiprompts

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/prompts/jinja"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"github.com/zetaoss/zengine/goapp/tasks/aiedit"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// promptView is a template with the run form it runs with (the user inputs of its chain it picks), or the error
// that stops it from running.
type promptView struct {
	models.AIEditPrompt
	UserInputs []prompts.Input `json:"user_inputs"`
	Error      string          `json:"error,omitempty"`
}

func newPromptView(p models.AIEditPrompt) promptView {
	v := promptView{AIEditPrompt: p, UserInputs: []prompts.Input{}}
	if spec, err := aiedit.TemplatePrompt(p); err == nil {
		v.Spec = spec.YAML()
	}
	if chain, err := aiedit.TemplateChain(p); err != nil {
		v.Error = err.Error()
	} else {
		v.UserInputs = chain.UserInputs
	}
	return v
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

// readSpec reads a template's spec for its request type: a prompts.PromptTemplate with a name that runs on the
// chain (prompts.Chain.WithTemplate).
func readSpec(requestType, spec string) (prompts.PromptTemplate, prompts.Chain, error) {
	t, err := prompts.ParsePromptTemplate(spec)
	if err != nil {
		return t, prompts.Chain{}, err
	}
	t.Name = strings.TrimSpace(t.Name)
	if t.Name == "" {
		return t, prompts.Chain{}, errors.New("name은 필수입니다")
	}
	chain, err := aiedit.ChainOf(requestType).WithTemplate(t)
	return t, chain, err
}

// Store creates a template, or updates the caller's own, from {"id", "request_type", "spec", "model", "bump"}. The
// spec's name and description are kept as columns too; model, outside the spec, is the model the template runs on
// ("" for its chain's). A change to its template or input_variables bumps the version: the minor, or with bump
// "major" the major; a change of model does not.
func Store(c *serverctx.Context) {
	var body struct {
		ID          int    `json:"id"`
		RequestType string `json:"request_type"`
		Spec        string `json:"spec"`
		Model       string `json:"model"`
		Bump        string `json:"bump"`
	}
	if !c.Decode(&body) {
		return
	}
	requestType, ok := models.NormalizeAIEditRequestType(body.RequestType)
	if !ok {
		c.JSONError(http.StatusUnprocessableEntity, "유효하지 않은 요청 유형입니다.")
		return
	}
	spec, _, err := readSpec(requestType, body.Spec)
	if err != nil {
		c.JSONError(http.StatusUnprocessableEntity, err.Error())
		return
	}
	model := strings.TrimSpace(body.Model)
	if model != "" {
		if model, err = config.ValidateModel(model); err != nil {
			c.JSONError(http.StatusUnprocessableEntity, err.Error())
			return
		}
	}

	user, _ := c.User()
	now := time.Now().Format("2006-01-02 15:04:05")
	row := models.AIEditPrompt{
		ID:          body.ID,
		Title:       spec.Name,
		Description: strings.TrimSpace(spec.Description),
		RequestType: requestType,
		Spec:        spec.YAML(),
		Model:       model,
		UpdatedAt:   now,
	}

	var count int64
	c.DB.Model(&models.AIEditPrompt{}).Where("title = ? AND id != ?", row.Title, row.ID).Count(&count)
	if count > 0 {
		c.JSONError(http.StatusConflict, "이미 사용 중인 이름입니다.")
		return
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
		before, err := aiedit.TemplatePrompt(existing)
		if err != nil || before.Template != spec.Template || !slices.Equal(before.InputVariables, spec.InputVariables) || existing.RequestType != requestType {
			row.Version = prompts.NextVersion(existing.Version, body.Bump == "major")
		}
		columns := []string{"title", "description", "request_type", "spec", "model", "version", "updated_at"}
		if err := c.DB.Model(&models.AIEditPrompt{}).Where("id = ?", row.ID).Select(columns).Updates(&row).Error; err != nil {
			c.InternalError()
			return
		}
		row.UserID, row.UserName, row.CreatedAt, row.UseCount, row.UsedAt = existing.UserID, existing.UserName, existing.CreatedAt, existing.UseCount, existing.UsedAt
	} else {
		row.UserID = user.ID
		row.UserName = user.Name
		row.CreatedAt = now
		row.Version = "v1.0"
		if err := c.DB.Create(&row).Error; err != nil {
			c.InternalError()
			return
		}
	}
	c.JSON(newPromptView(row))
}

// span is template syntax in the spec's YAML, for the editor to mark: a line from 1 and a column range from 1, in
// characters.
type span struct {
	Line  int            `json:"line"`
	Start int            `json:"start"`
	End   int            `json:"end"`
	Kind  jinja.SpanKind `json:"kind"`
}

// Check reads {"request_type", "spec", "preview_values"} as Store would and returns the template's run form, the
// template syntax it reads (spans), and its messages rendered with preview_values (the run form's values, and
// title and body; a value left out takes a sample), or the error that would stop it from saving. It stores nothing.
func Check(c *serverctx.Context) {
	var body struct {
		RequestType   string         `json:"request_type"`
		Spec          string         `json:"spec"`
		PreviewValues map[string]any `json:"preview_values"`
	}
	if !c.Decode(&body) {
		return
	}
	requestType, ok := models.NormalizeAIEditRequestType(body.RequestType)
	if !ok {
		c.JSONError(http.StatusUnprocessableEntity, "유효하지 않은 요청 유형입니다.")
		return
	}
	spec, chain, err := readSpec(requestType, body.Spec)
	if err != nil {
		c.JSON(app.H{"ok": false, "error": err.Error()})
		return
	}
	response := app.H{"ok": true, "user_inputs": chain.UserInputs, "spans": templateSpans(body.Spec, spec, chain)}
	values := map[string]any{}
	for _, in := range chain.UserInputs {
		if v, ok := body.PreviewValues[in.Name]; ok {
			values[in.Name] = v
		}
	}
	vars, err := chain.UserVars(values)
	if err != nil {
		response["preview_error"] = err.Error()
		c.JSON(response)
		return
	}
	samples := map[string]string{"title": "Hello World", "body": "문서 시작했다\n문서 끝났다"}
	for _, in := range chain.Inputs {
		vars[in.Name] = samples[in.Name]
		if v, ok := body.PreviewValues[in.Name].(string); ok {
			vars[in.Name] = v
		}
	}
	response["rendered_messages"] = chain.PromptStep().Render(vars)
	c.JSON(response)
}

// templateSpans places the template syntax of the spec's template on the lines of its YAML: the template is a
// block whose lines follow its key, each indented alike.
func templateSpans(text string, spec prompts.PromptTemplate, chain prompts.Chain) []span {
	ct, err := prompts.ParseChat(spec.Template, chain.PromptStep().Inputs)
	if err != nil {
		return nil
	}
	var root yaml.Node
	if yaml.Unmarshal([]byte(text), &root) != nil || len(root.Content) == 0 {
		return nil
	}
	m := root.Content[0]
	var node *yaml.Node
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == "template" {
			node = m.Content[i+1]
		}
	}
	if node == nil || (node.Style != yaml.LiteralStyle && node.Style != yaml.FoldedStyle) {
		return nil
	}
	lines := strings.Split(text, "\n")
	first := node.Line
	for first < len(lines) && strings.TrimSpace(lines[first]) == "" {
		first++
	}
	if first >= len(lines) {
		return nil
	}
	indent := len(lines[first]) - len(strings.TrimLeft(lines[first], " "))
	tpl := spec.Template
	var out []span
	for _, s := range ct.Spans() {
		line := strings.Count(tpl[:s.Start], "\n")
		lineStart := strings.LastIndex(tpl[:s.Start], "\n") + 1
		if strings.Contains(tpl[s.Start:s.End], "\n") {
			continue
		}
		start := utf8.RuneCountInString(tpl[lineStart:s.Start]) + indent + 1
		out = append(out, span{Line: node.Line + 1 + line, Start: start, End: start + utf8.RuneCountInString(tpl[s.Start:s.End]), Kind: s.Kind})
	}
	return out
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
