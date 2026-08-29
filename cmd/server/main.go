package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/quixiq/polyglot/internal/app"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/pkg/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	logger.Init(cfg.LogLevel, cfg.AppEnv)
	if err := cfg.Validate(); err != nil {
		logger.WithComponent("Server").WithError(err).Error("invalid configuration")
		os.Exit(1)
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		logger.WithComponent("Server").WithError(err).Error("failed to initialize application")
		os.Exit(1)
	}

	runErr := make(chan error, 1)
	go func() {
		runErr <- application.Run()
	}()

	select {
	case <-ctx.Done():
	case err := <-runErr:
		if err != nil {
			logger.WithComponent("Server").WithError(err).Error("application run error")
			os.Exit(1)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		logger.WithComponent("Server").WithError(err).Error("shutdown error")
	}
}
