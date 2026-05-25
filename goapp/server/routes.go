package server

import (
	"net/http"
	"time"

	"github.com/zetaoss/zengine/goapp/models"
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
	"github.com/zetaoss/zengine/goapp/server/router"
	"github.com/zetaoss/zengine/goapp/server/runtime/common"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

func RegisterRoutes(mux *http.ServeMux, serverCtx *serverctx.Context, components *common.Components) (*router.Router, error) {
	r := router.New(mux, serverCtx)

	r.GET("/api/me", me.Me, r.WithUser())
	r.GET("/api/me/avatar", me.GetAvatar, r.Unblocked())
	r.POST("/api/me/avatar", me.UpdateAvatar, r.Unblocked())
	r.GET("/api/me/gravatar/verify", me.VerifyGravatar, r.Unblocked())

	r.GET("/api/comments/recent", comments.Recent)
	r.GET("/api/comments/{pageID}", comments.List)
	r.POST("/api/comments", comments.Store, r.Unblocked())
	r.PUT("/api/comments/{id}", comments.Update, r.Owner(models.PageComment{}))
	r.DELETE("/api/comments/{id}", comments.Destroy, r.OwnerOrSysop(models.PageComment{}))

	r.GET("/api/binders", binder.Index)
	r.PUT("/api/binders/{binder}", binder.Update, r.Unblocked())

	r.GET("/api/common-report", commonreport.Index)
	r.GET("/api/common-report/{id}", commonreport.Show)
	r.POST("/api/common-report", commonreport.Store, r.Unblocked())
	r.POST("/api/common-report/{id}/clone", commonreport.Clone, r.Unblocked())
	r.POST("/api/common-report/{id}/rerun", commonreport.Rerun, r.OwnerOrSysop(models.CommonReport{}))
	r.DELETE("/api/common-report/{id}", commonreport.Destroy, r.OwnerOrSysop(models.CommonReport{}))

	r.GET("/api/editbot", editbot.Index)
	r.GET("/api/editbot/{id}", editbot.Show)
	r.POST("/api/editbot/from-page", editbot.StoreFromPage, r.User())
	r.POST("/api/editbot/from-write-request/id/{writeRequest}", editbot.StoreFromWriteRequest, r.User())
	r.DELETE("/api/editbot/{id}", editbot.Destroy, r.Sysop())

	r.GET("/api/editbot/prompts", editbot.PromptIndex, r.WithUser())
	r.GET("/api/editbot/prompts/exists", editbot.PromptExists)
	r.GET("/api/editbot/prompts/{id}", editbot.PromptShow, r.WithUser())
	r.POST("/api/editbot/prompts", editbot.PromptStore, r.Unblocked())
	r.POST("/api/editbot/prompts/{id}/favorite", editbot.PromptToggleFavorite, r.User())
	r.DELETE("/api/editbot/prompts/{id}", editbot.PromptDestroy, r.OwnerOrSysop(models.EditbotPrompt{}))

	r.GET("/api/internal/profiles/{id}", internalprofile.Show, r.Internal())

	r.GET("/api/reactions/page/{page}", pagereaction.Show)
	r.POST("/api/reactions/page", pagereaction.Store, r.User())

	r.GET("/api/runbox/{hash}", runbox.Show)
	r.POST("/api/runbox", runbox.Store)
	r.POST("/api/runbox/{hash}/rerun", runbox.Rerun, r.User())

	r.GET("/api/posts", post.Index)
	r.GET("/api/posts/recent", post.Recent)
	r.GET("/api/posts/{id}", post.Show)
	r.POST("/api/posts", post.Store, r.User())
	r.PUT("/api/posts/{id}", post.Update, r.Owner(models.ForumPost{}))
	r.DELETE("/api/posts/{id}", post.Destroy, r.OwnerOrSysop(models.ForumPost{}))

	r.GET("/api/posts/{postId}/replies", reply.Index)
	r.POST("/api/posts/{postId}/replies", reply.Store, r.User())
	r.PUT("/api/posts/{postId}/replies/{id}", reply.Update, r.Owner(models.ForumReply{}))
	r.DELETE("/api/posts/{postId}/replies/{id}", reply.Destroy, r.OwnerOrSysop(models.ForumReply{}))

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
	r.POST("/api/onelines", oneline.Store, r.Unblocked())
	r.DELETE("/api/onelines/{id}", oneline.Destroy, r.User())

	r.GET("/api/write-request/count", writerequest.Count)
	r.GET("/api/write-request/todo", writerequest.IndexTodo)
	r.GET("/api/write-request/todo-top", writerequest.IndexTodoTop)
	r.GET("/api/write-request/done", writerequest.IndexDone)
	r.POST("/api/write-request", writerequest.Store, r.Unblocked())
	r.POST("/api/write-request/{id}/recommend", writerequest.Recommend, r.Unblocked())
	r.DELETE("/api/write-request/{id}", writerequest.Destroy, r.User())

	authThrottle := r.Throttle(30, time.Minute, "auth-social")
	r.GET("/auth/redirect/{provider}", social.Redirect, authThrottle)
	r.GET("/auth/callback/{provider}", social.Callback, authThrottle)
	r.POST("/auth/deauthorize/{provider}", social.Deauthorize, authThrottle)
	r.POST("/auth/deletion/{provider}", social.Deletion, authThrottle)
	r.GET("/auth/deletion/{provider}/status/{code}", social.DeletionStatus, authThrottle)

	mux.Handle("/", components.RootHandler)
	return r, nil
}
