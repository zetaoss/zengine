package main

import (
	"net/http"
	"os"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/cmd/ctl/tablewriter"
	"github.com/zetaoss/zengine/goapp/server"
	"github.com/zetaoss/zengine/goapp/server/runtime/common"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

func runRoutes(cfg *config.Config) error {
	serverCtx, err := serverctx.New(cfg)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	components := &common.Components{RootHandler: http.NotFoundHandler()}
	r, err := server.RegisterRoutes(mux, serverCtx, components)
	if err != nil {
		return err
	}

	routes := r.Routes()
	tw := tablewriter.New(os.Stdout, "method", "path")
	if err := tw.Header(); err != nil {
		return err
	}
	for _, route := range routes {
		if err := tw.Row(route.Method, route.Path); err != nil {
			return err
		}
	}
	return tw.Flush()
}
