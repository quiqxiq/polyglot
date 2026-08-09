package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/quixiq/polyglot/internal/app"
	"github.com/quixiq/polyglot/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	application, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize application: %v", err)
	}

	go func() {
		if err := application.Run(); err != nil {
			log.Fatalf("application run error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
