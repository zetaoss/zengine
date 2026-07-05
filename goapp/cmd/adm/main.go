package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/zetaoss/zengine/goapp/cmd/adm/extensions"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("adm failed", "err", err)
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		extensions.Usage()
		return nil
	}

	switch args[0] {
	case "extensions":
		return extensions.Run(args[1:])
	case "help", "-h", "--help":
		extensions.Usage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
