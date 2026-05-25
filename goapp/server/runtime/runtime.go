package runtime

import (
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/server/runtime/common"
	"github.com/zetaoss/zengine/goapp/server/runtime/dev"
	"github.com/zetaoss/zengine/goapp/server/runtime/prod"
)

func NewComponents(cfg *config.Config) (*common.Components, error) {
	if cfg.App.DevMode {
		return dev.NewComponents(cfg)
	}
	return prod.NewComponents(cfg)
}
