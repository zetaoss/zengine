package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
)

type Injector struct {
	zConfStatic string
}

func NewInjector(cfg *config.Config) (*Injector, error) {
	zConfPayload, err := json.Marshal(app.H{
		"avatarBaseUrl":   cfg.App.AvatarBaseURL,
		"gaMeasurementId": cfg.Analytics.GAMeasurementID,
		"adClient":        cfg.Ads.Client,
		"adSlots":         cfg.Ads.Slots,
	})
	if err != nil {
		return nil, fmt.Errorf("build zconf: %w", err)
	}
	if len(zConfPayload) < 2 || zConfPayload[0] != '{' || zConfPayload[len(zConfPayload)-1] != '}' {
		return nil, fmt.Errorf("invalid zconf payload")
	}
	zConfStatic := string(zConfPayload[1 : len(zConfPayload)-1])
	return &Injector{zConfStatic: zConfStatic}, nil
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
