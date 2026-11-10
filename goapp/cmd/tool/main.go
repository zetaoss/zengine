package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/database"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/tasks/aiedit"
	"github.com/zetaoss/zengine/goapp/worker/registry"

	"github.com/hibiken/asynq"
)

func main() {
	args := os.Args[1:]
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}
	config.ConfigureSlog(cfg.App.LogLevel)
	reg := registry.New()
	if err := run(cfg, reg, args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(cfg *config.Config, reg *registry.Registry, args []string) error {
	if len(args) == 0 {
		printUsage(reg)
		return nil
	}
	switch args[0] {
	case "migrate":
		return runMigrate(cfg)
	case "stat-interpolate-daily":
		return runStatInterpolateDaily(cfg, args[1:])
	case "tasks":
		return withInspector(cfg, func(inspector *asynq.Inspector) error {
			return runTasks(reg, inspector, args[1:])
		})
	case "flush":
		return withInspector(cfg, func(inspector *asynq.Inspector) error {
			return runFlush(registry.ConsumeQueues(reg.AllSpecs()), inspector, args[1:])
		})
	case "routes":
		return runRoutes(cfg)
	case "metrics":
		return runMetrics(cfg, args[1:])
	case "help", "-h", "--help":
		printUsage(reg)
		return nil
	default:
		return runTaskCommand(cfg, reg, args[0], args[1:])
	}
}

func withInspector(cfg *config.Config, fn func(*asynq.Inspector) error) error {
	conn, err := appredis.AsynqConnOpt(cfg)
	if err != nil {
		return err
	}
	inspector := asynq.NewInspector(conn)
	err = fn(inspector)
	closeErr := inspector.Close()
	if err != nil {
		return err
	}
	return closeErr
}

// runMigrate applies the SQL migrations, then converts the data that needs Go to convert: AI edit prompt
// templates saved before they had a spec (aiedit.ConvertLegacyPrompts).
func runMigrate(cfg *config.Config) error {
	if err := database.RunMigrate(cfg); err != nil {
		return err
	}
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	if sqlDB, err := db.DB(); err == nil {
		defer func() { _ = sqlDB.Close() }()
	}
	n, err := aiedit.ConvertLegacyPrompts(db)
	if err != nil {
		return fmt.Errorf("convert AI edit prompt templates: %w", err)
	}
	if n > 0 {
		fmt.Fprintf(os.Stderr, "[worker:migrate] converted %d AI edit prompt templates\n", n)
	}
	return nil
}
