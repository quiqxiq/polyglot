package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/logger"
)

// Run starts the HTTP server and blocks until it stops.
func (a *App) Run() error {
	logger.WithComponent("App").WithField("address", a.httpServer.Addr).Info("engine starting")
	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server failed: %w", err)
	}
	return nil
}

// Shutdown gracefully stops background workers and the HTTP server.
func (a *App) Shutdown(ctx context.Context) error {
	logger.WithComponent("App").Info("shutting down server...")
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
	if a.pingStreamMgr != nil {
		a.pingStreamMgr.Stop()
	}
	if a.sseHub != nil {
		a.sseHub.Close()
	}
	if err := a.httpServer.Shutdown(ctx); err != nil {
		logger.WithComponent("App").WithError(err).Warn("error shutting down HTTP server")
		_ = a.httpServer.Close()
	}
	if a.waManager != nil {
		a.waManager.DisconnectAll()
	}
	if a.registry != nil {
		_ = a.registry.Close()
	}
	logger.WithComponent("App").Info("shutdown complete")
	return nil
}

func loadInitialDevices(pgStore *postgres.Store, encKey string) (port.DeviceRepository, port.CredentialVault) {
	return postgres.NewDeviceRepository(pgStore.DB()), postgres.NewCredentialVault(pgStore.DB(), encKey)
}
