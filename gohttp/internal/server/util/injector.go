package util

import (
	"bytes"
	"net/http"
	"strconv"
)

type Injector struct {
	zConfStatic string
}

func NewInjector(zConfStatic string) *Injector {
	return &Injector{zConfStatic: zConfStatic}
}

func (i *Injector) InjectScript(html []byte, r *http.Request) []byte {
	return bytes.Replace(
		html,
		[]byte("</title>"),
		[]byte(`</title><script>window.ZCONF={`+i.zConfStatic+`,"policy":`+strconv.Quote(r.Header.Get("X-Policy"))+`};</script>`),
		1,
	)
}
