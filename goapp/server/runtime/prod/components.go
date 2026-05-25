package prod

import (
	"fmt"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/server/runtime/common"
)

func NewComponents(cfg *config.Config) (*common.Components, error) {
	injector, err := common.NewInjector(cfg)
	if err != nil {
		return nil, fmt.Errorf("build prod injector: %w", err)
	}
	rootHandler, err := newRootHandler(injector)
	if err != nil {
		return nil, fmt.Errorf("build prod root handler: %w", err)
	}

	return &common.Components{
		RootHandler:         rootHandler,
		AccessLogMiddleware: AccessLogMiddleware,
	}, nil
}
