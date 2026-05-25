package dev

import (
	"fmt"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/server/runtime/common"
)

func NewComponents(cfg *config.Config) (*common.Components, error) {
	injector, err := common.NewInjector(cfg)
	if err != nil {
		return nil, fmt.Errorf("build dev injector: %w", err)
	}
	rootHandler := newRootHandler(injector)

	return &common.Components{
		RootHandler:         rootHandler,
		AccessLogMiddleware: AccessLogMiddleware,
	}, nil
}
