package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/zetaoss/zengine/goapp/server/handlers/api/binder"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/comments"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/commonreport"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/editbot"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/internalprofile"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/me"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/oneline"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/pagereaction"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/post"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/reply"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/runbox"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/stat"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/user"
	"github.com/zetaoss/zengine/goapp/server/handlers/api/writerequest"
	"github.com/zetaoss/zengine/goapp/server/handlers/auth/social"
	"github.com/zetaoss/zengine/goapp/server/handlers/root"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/router"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

func registerRoutes(mux *http.ServeMux, serverCtx *serverctx.Context) error {
	cfg := serverCtx.Cfg

	r := router.New(mux, serverCtx)

	r.GET("/api/me", me.Me, middleware.MaybeLoggedIn(cfg))
	r.GET("/api/me/avatar", me.GetAvatar, middleware.RequireUnblocked(cfg))
	r.POST("/api/me/avatar", me.UpdateAvatar, middleware.RequireUnblocked(cfg))
	r.GET("/api/me/gravatar/verify", me.VerifyGravatar, middleware.RequireUnblocked(cfg))

	r.GET("/api/comments/recent", comments.Recent)
	r.GET("/api/comments/{pageID}", comments.List)
	r.POST("/api/comments", comments.Store, middleware.RequireUnblocked(cfg))
	r.PUT("/api/comments/{id}", comments.Update, middleware.RequireUnblocked(cfg))
	r.DELETE("/api/comments/{id}", comments.Destroy, middleware.RequireUnblocked(cfg))

	r.GET("/api/binders", binder.Index)
	r.PUT("/api/binders/{binder}", binder.Update, middleware.RequireUnblocked(cfg))

	r.GET("/api/common-report", commonreport.Index)
	r.GET("/api/common-report/{id}", commonreport.Show)
	r.POST("/api/common-report", commonreport.Store, middleware.RequireUnblocked(cfg))
	r.POST("/api/common-report/{id}/clone", commonreport.Clone, middleware.RequireUnblocked(cfg))
	r.POST("/api/common-report/{id}/rerun", commonreport.Rerun, middleware.RequireUnblocked(cfg))
	r.DELETE("/api/common-report/{id}", commonreport.Destroy, middleware.RequireUnblocked(cfg))

	r.GET("/api/editbot", editbot.Index)
	r.GET("/api/editbot/{task}", editbot.Show)
	r.POST("/api/editbot/from-page", editbot.StoreFromPage, middleware.RequireLoggedIn(cfg))
	r.POST("/api/editbot/from-write-request/id/{writeRequest}", editbot.StoreFromWriteRequest, middleware.RequireLoggedIn(cfg))
	r.DELETE("/api/editbot/{task}", editbot.Destroy, middleware.RequireSysop(cfg))

	r.GET("/api/internal/profiles/{userId}", internalprofile.Show, middleware.RequireInternal(cfg))

	r.GET("/api/reactions/page/{page}", pagereaction.Show)
	r.POST("/api/reactions/page", pagereaction.Store, middleware.RequireLoggedIn(cfg))

	r.GET("/api/runbox/{hash}", runbox.Show)
	r.POST("/api/runbox", runbox.Store)
	r.POST("/api/runbox/{hash}/rerun", runbox.Rerun, middleware.RequireLoggedIn(cfg))

	r.GET("/api/posts", post.Index)
	r.GET("/api/posts/recent", post.Recent)
	r.GET("/api/posts/{post}", post.Show)
	r.POST("/api/posts", post.Store, middleware.RequireLoggedIn(cfg))
	r.PUT("/api/posts/{post}", post.Update, middleware.RequireLoggedIn(cfg))
	r.DELETE("/api/posts/{post}", post.Destroy, middleware.RequireLoggedIn(cfg))

	r.GET("/api/posts/{post}/replies", reply.Index)
	r.POST("/api/posts/{post}/replies", reply.Store, middleware.RequireLoggedIn(cfg))
	r.PUT("/api/posts/{post}/replies/{reply}", reply.Update, middleware.RequireLoggedIn(cfg))
	r.DELETE("/api/posts/{post}/replies/{reply}", reply.Destroy, middleware.RequireLoggedIn(cfg))

	r.GET("/api/user/{userId}/stats", user.Stats)
	r.GET("/api/user/{userName}", user.Show)

	r.GET("/api/stat/cf-analytics/hourly", stat.CFHourly)
	r.GET("/api/stat/cf-analytics/daily/{days}", stat.CFDaily)
	r.GET("/api/stat/ga/hourly", stat.GAHourly)
	r.GET("/api/stat/ga/daily/{days}", stat.GADaily)
	r.GET("/api/stat/gsc/hourly", stat.GSCHourly)
	r.GET("/api/stat/gsc/daily/{days}", stat.GSCDaily)
	r.GET("/api/stat/mw-statistics/hourly", stat.MWHourly)
	r.GET("/api/stat/mw-statistics/daily/{days}", stat.MWDaily)

	r.GET("/api/onelines/recent", oneline.Recent)
	r.GET("/api/onelines", oneline.Index)
	r.POST("/api/onelines", oneline.Store, middleware.RequireUnblocked(cfg))
	r.DELETE("/api/onelines/{id}", oneline.Destroy, middleware.RequireLoggedIn(cfg))

	r.GET("/api/write-request/count", writerequest.Count)
	r.GET("/api/write-request/todo", writerequest.IndexTodo)
	r.GET("/api/write-request/todo-top", writerequest.IndexTodoTop)
	r.GET("/api/write-request/done", writerequest.IndexDone)
	r.POST("/api/write-request", writerequest.Store, middleware.RequireUnblocked(cfg))
	r.POST("/api/write-request/{id}/recommend", writerequest.Recommend, middleware.RequireUnblocked(cfg))
	r.DELETE("/api/write-request/{id}", writerequest.Destroy, middleware.RequireLoggedIn(cfg))

	authThrottle := middleware.Throttle(30, time.Minute, "auth-social")
	r.GET("/auth/redirect/{provider}", social.Redirect, authThrottle)
	r.GET("/auth/callback/{provider}", social.Callback, authThrottle)
	r.POST("/auth/deauthorize/{provider}", social.Deauthorize, authThrottle)
	r.POST("/auth/deletion/{provider}", social.Deletion, authThrottle)
	r.GET("/auth/deletion/{provider}/status/{code}", social.DeletionStatus, authThrottle)

	rootHandler, err := root.New(cfg)
	if err != nil {
		return fmt.Errorf("build root handler: %w", err)
	}
	mux.Handle("/", rootHandler)
	return nil
}
