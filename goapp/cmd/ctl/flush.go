package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zetaoss/zengine/goapp/worker"
)

func runFlush(wkr *worker.Worker, args []string) error {
	if len(args) == 1 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "all", "running", "pending":
			return flushByMode(wkr, args[0])
		default:
			return fmt.Errorf("unknown flush mode: %s (use all|running|pending)", args[0])
		}
	}

	fs := flag.NewFlagSet("flush", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	all := fs.Bool("all", false, "flush all jobs (default)")
	running := fs.Bool("running", false, "flush running jobs only")
	pending := fs.Bool("pending", false, "flush pending jobs only")
	if err := fs.Parse(args); err != nil {
		return err
	}

	selected := 0
	if *all {
		selected++
	}
	if *running {
		selected++
	}
	if *pending {
		selected++
	}
	if selected > 1 {
		return fmt.Errorf("flush flags are mutually exclusive: use one of --all, --running, --pending")
	}

	mode := "all"
	if *running {
		mode = "running"
	}
	if *pending {
		mode = "pending"
	}

	return flushByMode(wkr, mode)
}

func flushByMode(wkr *worker.Worker, mode string) error {
	reason := "flushed by ctl"
	switch mode {
	case "running":
		n, err := wkr.FlushRunning(reason)
		if err != nil {
			return err
		}
		_, _ = fmt.Printf("flushed running: %d\n", n)
	case "pending":
		n, err := wkr.FlushPending(reason)
		if err != nil {
			return err
		}
		_, _ = fmt.Printf("flushed pending: %d\n", n)
	default:
		rn, err := wkr.FlushRunning(reason)
		if err != nil {
			return err
		}
		pn, err := wkr.FlushPending(reason)
		if err != nil {
			return err
		}
		_, _ = fmt.Printf("flushed all: running=%d pending=%d total=%d\n", rn, pn, rn+pn)
	}
	return nil
}
