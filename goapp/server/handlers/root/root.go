package root

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/server/handlers/root/dev"
	"github.com/zetaoss/zengine/goapp/server/handlers/root/injector"
	"github.com/zetaoss/zengine/goapp/server/handlers/root/prod"
)

func New(cfg *config.Config) (http.Handler, error) {
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
	injector := injector.NewInjector(zConfStatic)

	if cfg.App.DevMode {
		return dev.New(injector), nil
	}

	handler, err := prod.New(injector)
	if err != nil {
		return nil, fmt.Errorf("build prod handler: %w", err)
	}
	return handler, nil
}
