package network

import (
	"context"
	"errors"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

type mockStreamHandle struct {
	ch       chan command.Result
	cancelFn func()
}

func (m *mockStreamHandle) Chan() <-chan command.Result {
	return m.ch
}

func (m *mockStreamHandle) Cancel() error {
	if m.cancelFn != nil {
		m.cancelFn()
	}
	return nil
}

func (m *mockStreamHandle) Err() error {
	return nil
}

type mockStreamingDriver struct {
	port.DeviceDriver
	streamFn func(ctx context.Context, cmd command.Command) (port.StreamHandle, error)
}

func (d *mockStreamingDriver) Stream(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
	if d.streamFn != nil {
		return d.streamFn(ctx, cmd)
	}
	return nil, nil
}

func TestStreamPing(t *testing.T) {
	t.Run("successfully streams and parses ping results", func(t *testing.T) {
		ch := make(chan command.Result, 3)
		ch <- command.Result{
			Rows: []map[string]string{
				{
					"seq":         "1",
					"time":        "15ms",
					"status":      "echo reply",
					"ttl":         "56",
					"packet-loss": "0%",
				},
			},
		}
		ch <- command.Result{
			Rows: []map[string]string{
				{
					"sequence":    "2",
					"time":        "25ms",
					"status":      "echo reply",
					"ttl":         "56",
					"packet-loss": "0%",
				},
			},
		}
		ch <- command.Result{
			Rows: []map[string]string{
				{
					"status":      "timeout",
					"packet-loss": "100%",
				},
			},
		}
		close(ch)

		driver := &mockStreamingDriver{
			streamFn: func(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
				if cmd.Raw != "/ping" {
					t.Errorf("expected /ping, got %s", cmd.Raw)
				}
				if cmd.Args["address"] != "8.8.8.8" {
					t.Errorf("expected address 8.8.8.8, got %s", cmd.Args["address"])
				}
				if _, ok := cmd.Args["interval"]; ok {
					t.Errorf("expected no interval in command args, got %s", cmd.Args["interval"])
				}
				return &mockStreamHandle{ch: ch}, nil
			},
		}

		var collected []PingStreamItem
		err := StreamPing(context.Background(), driver, "8.8.8.8:8728", func(item PingStreamItem) error {
			collected = append(collected, item)
			return nil
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(collected) != 3 {
			t.Fatalf("expected 3 items, got %d", len(collected))
		}

		if collected[0].Seq != 1 || collected[0].LatencyMS != 15 || collected[0].PacketLoss != 0 {
			t.Errorf("unexpected item 0: %+v", collected[0])
		}
		if collected[1].Seq != 2 || collected[1].LatencyMS != 25 || collected[1].PacketLoss != 0 {
			t.Errorf("unexpected item 1: %+v", collected[1])
		}
		if collected[2].Seq != 3 || collected[2].PacketLoss != 100 || collected[2].Status != "timeout" {
			t.Errorf("unexpected item 2: %+v", collected[2])
		}
	})

	t.Run("returns ErrDriverNotStreaming when driver does not implement streaming", func(t *testing.T) {
		type nonStreamingDriver struct {
			port.DeviceDriver
		}
		err := StreamPing(context.Background(), nonStreamingDriver{}, "8.8.8.8", func(PingStreamItem) error {
			return nil
		})
		if !errors.Is(err, command.ErrDriverNotStreaming) {
			t.Errorf("expected ErrDriverNotStreaming, got %v", err)
		}
	})
}
