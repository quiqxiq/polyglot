package network

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectStream(t *testing.T) {
	ctx := context.Background()

	t.Run("nil driver returns ErrDriverUnavailable", func(t *testing.T) {
		_, err := CollectStream(ctx, nil, command.Command{Raw: "/ping"})
		assert.ErrorIs(t, err, command.ErrDriverUnavailable)
	})

	t.Run("collects ping results bounded by count", func(t *testing.T) {
		ch := make(chan command.Result, 5)
		ch <- command.Result{
			Rows: []map[string]string{
				{"seq": "1", "host": "8.8.8.8", "time": "12ms", "ttl": "56"},
			},
		}
		ch <- command.Result{
			Rows: []map[string]string{
				{"seq": "2", "host": "8.8.8.8", "time": "15ms", "ttl": "56"},
			},
		}
		cancelled := false
		handle := &mockStreamHandle{
			ch: ch,
			cancelFn: func() {
				cancelled = true
			},
		}

		driver := &mockStreamingDriver{
			streamFn: func(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
				assert.Equal(t, "/ping", cmd.Raw)
				assert.Equal(t, "2", cmd.Args["count"])
				return handle, nil
			},
		}

		cmd := command.Command{
			Raw:  "/ping",
			Args: map[string]string{"address": "8.8.8.8", "count": "2"},
		}

		res, err := CollectStream(ctx, driver, cmd)
		require.NoError(t, err)
		assert.Len(t, res.Rows, 2)
		assert.Contains(t, res.Output, "seq=1 host=8.8.8.8 time=12ms ttl=56")
		assert.Contains(t, res.Output, "seq=2 host=8.8.8.8 time=15ms ttl=56")
		assert.True(t, cancelled)
	})

	t.Run("sets default count=4 for ping if count is omitted", func(t *testing.T) {
		ch := make(chan command.Result, 5)
		for i := 1; i <= 4; i++ {
			ch <- command.Result{
				Rows: []map[string]string{
					{"seq": "1", "host": "1.1.1.1", "time": "10ms", "ttl": "58"},
				},
			}
		}
		handle := &mockStreamHandle{ch: ch}

		driver := &mockStreamingDriver{
			streamFn: func(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
				assert.Equal(t, "/ping", cmd.Raw)
				assert.Equal(t, "4", cmd.Args["count"])
				return handle, nil
			},
		}

		cmd := command.Command{
			Raw:  "/ping",
			Args: map[string]string{"address": "1.1.1.1"},
		}

		res, err := CollectStream(ctx, driver, cmd)
		require.NoError(t, err)
		assert.Len(t, res.Rows, 4)
	})

	t.Run("handles generic monitor traffic stream and timeout bounds", func(t *testing.T) {
		ch := make(chan command.Result, 5)
		ch <- command.Result{
			Rows: []map[string]string{
				{"rx-bits-per-second": "1000", "tx-bits-per-second": "500"},
			},
		}
		ch <- command.Result{
			Rows: []map[string]string{
				{"rx-bits-per-second": "1200", "tx-bits-per-second": "600"},
			},
		}
		ch <- command.Result{
			Rows: []map[string]string{
				{"rx-bits-per-second": "1100", "tx-bits-per-second": "550"},
			},
		}
		handle := &mockStreamHandle{ch: ch}

		driver := &mockStreamingDriver{
			streamFn: func(ctx context.Context, cmd command.Command) (port.StreamHandle, error) {
				assert.Equal(t, "/interface/monitor-traffic", cmd.Raw)
				return handle, nil
			},
		}

		cmd := command.Command{
			Raw:  "/interface/monitor-traffic",
			Args: map[string]string{"interface": "ether1"},
		}

		res, err := CollectStream(ctx, driver, cmd)
		require.NoError(t, err)
		assert.Len(t, res.Rows, 3)
		assert.Contains(t, res.Output, "rx-bits-per-second=1000")
	})
}
