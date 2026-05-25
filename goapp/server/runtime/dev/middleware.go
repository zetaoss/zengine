package dev

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/zetaoss/zengine/goapp/server/runtime/common"
)

func AccessLogMiddleware(next http.Handler) http.Handler {
	return common.AccessLogMiddleware(next, func(sw *common.StatusWriter) http.ResponseWriter {
		return &statusWriterDev{StatusWriter: sw}
	})
}

type statusWriterDev struct {
	*common.StatusWriter
}

func (w *statusWriterDev) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("http.Hijacker not implemented by underlying ResponseWriter")
	}
	return hj.Hijack()
}
