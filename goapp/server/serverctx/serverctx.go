package serverctx

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zetaoss/zengine/goapp/app/appctx"
	"github.com/zetaoss/zengine/goapp/app/config"

	"gorm.io/gorm"
)

type Context struct {
	*appctx.AppContext
	DB *gorm.DB
	W  http.ResponseWriter
	R  *http.Request
}

func New(cfg *config.Config) (*Context, error) {
	base, err := appctx.NewAppContext(cfg)
	if err != nil {
		return nil, err
	}
	db, err := base.GetDB()
	if err != nil {
		return nil, err
	}
	return &Context{AppContext: base, DB: db}, nil
}

func (c *Context) JSON(v any) {
	c.JSONStatus(http.StatusOK, v)
}

func (c *Context) JSONStatus(status int, v any) {
	c.W.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.W.WriteHeader(status)
	_ = json.NewEncoder(c.W).Encode(v)
}

func (c *Context) JSONError(status int, message string) {
	c.JSONStatus(status, map[string]string{"message": message})
}

func (c *Context) InternalError() {
	http.Error(c.W, "internal server error", http.StatusInternalServerError)
}

func (c *Context) NotFound() {
	http.NotFound(c.W, c.R)
}

func (c *Context) PathInt(key string) (int, bool) {
	id, err := strconv.Atoi(c.R.PathValue(key))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func (c *Context) Decode(v any) bool {
	if err := json.NewDecoder(c.R.Body).Decode(v); err != nil {
		c.JSONError(http.StatusUnprocessableEntity, "Invalid payload")
		return false
	}
	return true
}
