package util

import (
	"bytes"
	"net/http"
)

type Injector struct {
	zConfStatic string
}

func NewInjector(zConfStatic string) *Injector {
	return &Injector{zConfStatic: zConfStatic}
}

func (i *Injector) InjectScript(html []byte, r *http.Request) []byte {
	policy := "strict"
	if r.Header.Get("X-Policy") == "standard" {
		policy = "standard"
	}

	return bytes.Replace(
		html,
		[]byte("</title>"),
		[]byte(`</title><script>window.ZCONF={`+i.zConfStatic+`,"policy":"`+policy+`"};</script>`),
		1,
	)
}
