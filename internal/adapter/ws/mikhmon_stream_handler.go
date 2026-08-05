package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
)

// StreamDriverProvider returns a port.StreamingDeviceDriver for a deviceID.
type StreamDriverProvider func(ctx context.Context, deviceID string) (port.StreamingDeviceDriver, error)

// MikhmonStreamHandler manages real-time WebSocket streaming subscriptions.
type MikhmonStreamHandler struct {
	provider StreamDriverProvider
}

// NewMikhmonStreamHandler constructs a new MikhmonStreamHandler.
func NewMikhmonStreamHandler(provider StreamDriverProvider) *MikhmonStreamHandler {
	return &MikhmonStreamHandler{provider: provider}
}

// StreamTraffic continuously streams interface traffic stats over a channel/writer.
func (h *MikhmonStreamHandler) StreamTraffic(ctx context.Context, deviceID, iface string, out chan<- []byte) error {
	driver, err := h.provider(ctx, deviceID)
	if err != nil {
		return err
	}
	cmd := mikrotik.NewMonitorTrafficStreamCommand(iface)
	handle, err := driver.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start traffic stream: %w", err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			stats := mikrotik.ParseInterfaceTrafficStats(res)
			data, err := json.Marshal(stats)
			if err == nil {
				out <- data
			}
		}
	}
}

// StreamResource continuously streams system resource updates over a channel/writer.
func (h *MikhmonStreamHandler) StreamResource(ctx context.Context, deviceID string, out chan<- []byte) error {
	driver, err := h.provider(ctx, deviceID)
	if err != nil {
		return err
	}
	cmd := mikrotik.NewStreamSystemResourceCommand("1s")
	handle, err := driver.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start resource stream: %w", err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			sysRes := mikrotik.ParseSystemResource(res)
			data, err := json.Marshal(sysRes)
			if err == nil {
				out <- data
			}
		}
	}
}

// StreamActiveSessions continuously streams live hotspot active session changes.
func (h *MikhmonStreamHandler) StreamActiveSessions(ctx context.Context, deviceID, userFilter string, out chan<- []byte) error {
	driver, err := h.provider(ctx, deviceID)
	if err != nil {
		return err
	}
	cmd := mikrotik.NewStreamHotspotActiveCommand(userFilter)
	handle, err := driver.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start active sessions stream: %w", err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			sessions := mikrotik.ParseHotspotActiveSessions(res)
			data, err := json.Marshal(sessions)
			if err == nil {
				out <- data
			}
		}
	}
}

// StreamQueueStats continuously streams simple queue statistics over a channel/writer.
func (h *MikhmonStreamHandler) StreamQueueStats(ctx context.Context, deviceID string, p mikrotik.QueueStreamParams, out chan<- []byte) error {
	driver, err := h.provider(ctx, deviceID)
	if err != nil {
		return err
	}
	cmd := mikrotik.NewStreamQueueStatsCommand(p)
	handle, err := driver.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start queue stats stream: %w", err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			queues := mikrotik.ParseSimpleQueues(res)
			data, err := json.Marshal(queues)
			if err == nil {
				out <- data
			}
		}
	}
}

// StreamHotspotUsers continuously streams live Hotspot user directory updates.
func (h *MikhmonStreamHandler) StreamHotspotUsers(ctx context.Context, deviceID, profileFilter string, out chan<- []byte) error {
	driver, err := h.provider(ctx, deviceID)
	if err != nil {
		return err
	}
	cmd := mikrotik.NewStreamHotspotUsersCommand(profileFilter)
	handle, err := driver.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start hotspot users stream: %w", err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			users := mikrotik.ParseHotspotUsers(res)
			data, err := json.Marshal(users)
			if err == nil {
				out <- data
			}
		}
	}
}

// StreamPPPActive continuously streams live PPPoE active session updates.
func (h *MikhmonStreamHandler) StreamPPPActive(ctx context.Context, deviceID, userFilter string, out chan<- []byte) error {
	driver, err := h.provider(ctx, deviceID)
	if err != nil {
		return err
	}
	cmd := mikrotik.NewStreamPPPActiveCommand(userFilter)
	handle, err := driver.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start ppp active stream: %w", err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			sessions := mikrotik.ParsePPPActiveSessions(res)
			data, err := json.Marshal(sessions)
			if err == nil {
				out <- data
			}
		}
	}
}

// StreamPPPOESecrets continuously streams live PPPoE subscriber secret updates.
func (h *MikhmonStreamHandler) StreamPPPOESecrets(ctx context.Context, deviceID, nameFilter string, out chan<- []byte) error {
	driver, err := h.provider(ctx, deviceID)
	if err != nil {
		return err
	}
	cmd := mikrotik.NewStreamPPPoESecretsCommand(nameFilter)
	handle, err := driver.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to start ppp secrets stream: %w", err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			secrets := mikrotik.ParsePPPoESecrets(res)
			data, err := json.Marshal(secrets)
			if err == nil {
				out <- data
			}
		}
	}
}

// StreamHotspotInactive continuously streams live inactive Hotspot users
// by combining streams of /ip/hotspot/user/print follow and /ip/hotspot/active/print follow.
func (h *MikhmonStreamHandler) StreamHotspotInactive(ctx context.Context, deviceID string, out chan<- []byte) error {
	driver, err := h.provider(ctx, deviceID)
	if err != nil {
		return err
	}

	userCmd := mikrotik.NewStreamHotspotUsersCommand("")
	userHandle, err := driver.Stream(ctx, userCmd)
	if err != nil {
		return fmt.Errorf("failed to start hotspot users stream: %w", err)
	}
	defer userHandle.Cancel()

	activeCmd := mikrotik.NewStreamHotspotActiveCommand("")
	activeHandle, err := driver.Stream(ctx, activeCmd)
	if err != nil {
		return fmt.Errorf("failed to start hotspot active stream: %w", err)
	}
	defer activeHandle.Cancel()

	var (
		users  []mikrotik.HotspotUser
		active []mikrotik.HotspotActiveSession
	)

	emit := func() {
		inactive := mikrotik.FilterInactiveHotspotUsers(users, active)
		data, err := json.Marshal(inactive)
		if err == nil {
			out <- data
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-userHandle.Chan():
			if !ok {
				return userHandle.Err()
			}
			users = mikrotik.ParseHotspotUsers(res)
			emit()
		case res, ok := <-activeHandle.Chan():
			if !ok {
				return activeHandle.Err()
			}
			active = mikrotik.ParseHotspotActiveSessions(res)
			emit()
		}
	}
}

// StreamPPPOEInactive continuously streams live inactive PPPoE subscribers
// by combining streams of /ppp/secret/print follow and /ppp/active/print follow.
func (h *MikhmonStreamHandler) StreamPPPOEInactive(ctx context.Context, deviceID string, out chan<- []byte) error {
	driver, err := h.provider(ctx, deviceID)
	if err != nil {
		return err
	}

	secretCmd := mikrotik.NewStreamPPPoESecretsCommand("")
	secretHandle, err := driver.Stream(ctx, secretCmd)
	if err != nil {
		return fmt.Errorf("failed to start ppp secrets stream: %w", err)
	}
	defer secretHandle.Cancel()

	activeCmd := mikrotik.NewStreamPPPActiveCommand("")
	activeHandle, err := driver.Stream(ctx, activeCmd)
	if err != nil {
		return fmt.Errorf("failed to start ppp active stream: %w", err)
	}
	defer activeHandle.Cancel()

	var (
		secrets []mikrotik.PPPoESecret
		active  []mikrotik.PPPActiveSession
	)

	emit := func() {
		inactive := mikrotik.FilterInactivePPPoESecrets(secrets, active)
		data, err := json.Marshal(inactive)
		if err == nil {
			out <- data
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-secretHandle.Chan():
			if !ok {
				return secretHandle.Err()
			}
			secrets = mikrotik.ParsePPPoESecrets(res)
			emit()
		case res, ok := <-activeHandle.Chan():
			if !ok {
				return activeHandle.Err()
			}
			active = mikrotik.ParsePPPActiveSessions(res)
			emit()
		}
	}
}

// HTTPHandlerUpgrade placeholder for HTTP WebSocket upgrade.
func HTTPHandlerUpgrade(w http.ResponseWriter, r *http.Request) {
	// Handled by WebSocket upgrader in production server setup
}
