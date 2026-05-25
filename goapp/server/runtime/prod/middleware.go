package prod

import (
	"net/http"

	"github.com/zetaoss/zengine/goapp/server/runtime/common"
)

func AccessLogMiddleware(next http.Handler) http.Handler {
	return common.AccessLogMiddleware(next, func(sw *common.StatusWriter) http.ResponseWriter {
		return sw
	})
}
