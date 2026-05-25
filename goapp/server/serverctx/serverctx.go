package serverctx

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/zetaoss/zengine/goapp/app/appctx"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"

	"gorm.io/gorm"
)

type contextKey string

const MWUserKey contextKey = "mw_user"

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
	c.JSONErrorStatus(c.W, status, message)
}

func (c *Context) JSONErrorStatus(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func (c *Context) InternalError() {
	c.JSONError(http.StatusInternalServerError, "Internal Server Error")
}

func (c *Context) NotFound() {
	c.JSONError(http.StatusNotFound, "Not Found")
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

func (c *Context) User() (models.MWUser, bool) {
	u, ok := c.R.Context().Value(MWUserKey).(models.MWUser)
	return u, ok
}

func (c *Context) SetUser(user models.MWUser) {
	c.R = c.R.WithContext(context.WithValue(c.R.Context(), MWUserKey, user))
}
