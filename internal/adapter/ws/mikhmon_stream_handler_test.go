package ws

import (
	"context"
	"testing"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyStreamHandle struct {
	ch  chan command.Result
	err error
}

func (h *dummyStreamHandle) Chan() <-chan command.Result {
	return h.ch
}

func (h *dummyStreamHandle) Cancel() error {
	close(h.ch)
	return nil
}

func (h *dummyStreamHandle) Err() error {
	return h.err
}

type dummyStreamingDriver struct {
	streamFn func(ctx context.Context, cmd command.Command) (port.StreamHandle, error)
}

func (d *dummyStreamingDriver) Stream(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
	if d.streamFn != nil {
		return d.streamFn(ctx, cmd)
	}
	ch := make(chan command.Result)
	return &dummyStreamHandle{ch: ch}, nil
}

func (d *dummyStreamingDriver) Execute(ctx context.Context, cmd command.Command) (command.Result, error) {
	return command.Result{}, nil
}

func (d *dummyStreamingDriver) Classify(cmd command.Command) command.Class {
	return command.ClassReadOnly
}

func (d *dummyStreamingDriver) Translate(op command.Operation) (command.Command, error) {
	return command.Command{}, nil
}

func (d *dummyStreamingDriver) Close() error {
	return nil
}

func TestMikhmonStreamHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	driver := &dummyStreamingDriver{
		streamFn: func(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
			ch := make(chan command.Result, 1)
			ch <- command.Result{Rows: []map[string]string{
				{"rx-bits-per-second": "1000", "tx-bits-per-second": "2000"},
			}}
			return &dummyStreamHandle{ch: ch}, nil
		},
	}

	handler := NewMikhmonStreamHandler(func(ctx context.Context, deviceID string) (port.StreamingDeviceDriver, error) {
		return driver, nil
	})

	outChan := make(chan []byte, 5)
	errCh := make(chan error, 1)

	go func() {
		errCh <- handler.StreamTraffic(ctx, "dev-123", "ether1", outChan)
	}()

	select {
	case data := <-outChan:
		assert.Contains(t, string(data), "1000")
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for streaming data")
	}

	cancel()
	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestMikhmonStreamHandler_StreamQueueStats(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	var capturedCmd command.Command
	driver := &dummyStreamingDriver{
		streamFn: func(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
			capturedCmd = cmd
			ch := make(chan command.Result, 1)
			ch <- command.Result{Rows: []map[string]string{
				{
					".id": "*1", "name": "Parent-Total-Bandwidth", "target": "192.168.0.0/16",
					"max-limit": "100M/100M", "rate": "5000000/10000000", "packet-rate": "500/1000",
				},
			}}
			return &dummyStreamHandle{ch: ch}, nil
		},
	}

	handler := NewMikhmonStreamHandler(func(ctx context.Context, deviceID string) (port.StreamingDeviceDriver, error) {
		return driver, nil
	})

	outChan := make(chan []byte, 5)
	errCh := make(chan error, 1)

	params := mikrotik.QueueStreamParams{
		NameFilter:   "Parent-Total-Bandwidth",
		ParentFilter: "none",
		ParentsOnly:  true,
		Interval:     "1s",
	}

	go func() {
		errCh <- handler.StreamQueueStats(ctx, "dev-123", params, outChan)
	}()

	select {
	case data := <-outChan:
		assert.Contains(t, string(data), "Parent-Total-Bandwidth")
		assert.Contains(t, string(data), "5000000/10000000")
		assert.Equal(t, "/queue/simple/print", capturedCmd.Raw)
		assert.Equal(t, "Parent-Total-Bandwidth", capturedCmd.Args["?name"])
		assert.Equal(t, "none", capturedCmd.Args["?parent"])
		assert.Equal(t, "false", capturedCmd.Args["?dynamic"])
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for queue streaming data")
	}

	cancel()
	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestMikhmonStreamHandler_StreamHotspotInactive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	driver := &dummyStreamingDriver{
		streamFn: func(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
			ch := make(chan command.Result, 1)
			if cmd.Raw == "/ip/hotspot/user/print" {
				ch <- command.Result{Rows: []map[string]string{
					{".id": "*1", "name": "user-active"},
					{".id": "*2", "name": "user-inactive"},
				}}
			} else if cmd.Raw == "/ip/hotspot/active/print" {
				ch <- command.Result{Rows: []map[string]string{
					{".id": "*A1", "user": "user-active"},
				}}
			}
			return &dummyStreamHandle{ch: ch}, nil
		},
	}

	handler := NewMikhmonStreamHandler(func(ctx context.Context, deviceID string) (port.StreamingDeviceDriver, error) {
		return driver, nil
	})

	outChan := make(chan []byte, 5)
	errCh := make(chan error, 1)

	go func() {
		errCh <- handler.StreamHotspotInactive(ctx, "dev-123", outChan)
	}()

	var received string
	for i := 0; i < 2; i++ {
		select {
		case data := <-outChan:
			received = string(data)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for hotspot inactive stream data")
		}
	}
	assert.Contains(t, received, "user-inactive")
	assert.NotContains(t, received, "user-active")

	cancel()
	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

func TestMikhmonStreamHandler_StreamPPPOEInactive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	driver := &dummyStreamingDriver{
		streamFn: func(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
			ch := make(chan command.Result, 1)
			if cmd.Raw == "/ppp/secret/print" {
				ch <- command.Result{Rows: []map[string]string{
					{".id": "*1", "name": "ppp-active"},
					{".id": "*2", "name": "ppp-inactive"},
				}}
			} else if cmd.Raw == "/ppp/active/print" {
				ch <- command.Result{Rows: []map[string]string{
					{".id": "*A1", "name": "ppp-active"},
				}}
			}
			return &dummyStreamHandle{ch: ch}, nil
		},
	}

	handler := NewMikhmonStreamHandler(func(ctx context.Context, deviceID string) (port.StreamingDeviceDriver, error) {
		return driver, nil
	})

	outChan := make(chan []byte, 5)
	errCh := make(chan error, 1)

	go func() {
		errCh <- handler.StreamPPPOEInactive(ctx, "dev-123", outChan)
	}()

	var received string
	for i := 0; i < 2; i++ {
		select {
		case data := <-outChan:
			received = string(data)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for ppp inactive stream data")
		}
	}
	assert.Contains(t, received, "ppp-inactive")
	assert.NotContains(t, received, "ppp-active")

	cancel()
	err := <-errCh
	require.ErrorIs(t, err, context.Canceled)
}

