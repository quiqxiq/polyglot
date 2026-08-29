package network

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/ping"
)

type PingStreamItem struct {
	Seq           int32
	Host          string
	LatencyMS     int64
	Status        string
	TTL           int32
	PacketLoss    int32
	TimestampUnix int64
}

// StreamPing orchestrates streaming ping execution on a device driver.
func StreamPing(ctx context.Context, driver port.DeviceDriver, host string, onResult func(item PingStreamItem) error) error {
	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return command.ErrDriverNotStreaming
	}

	cmd := command.Command{
		Raw: "/ping",
		Args: map[string]string{
			"address":  host,
			"interval": "1s",
		},
	}

	handle, err := sd.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to initiate ping stream: %w", err)
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
			if len(res.Rows) == 0 {
				continue
			}

			for _, row := range res.Rows {
				lat, status := ping.ParsePingLatency(row)
				seq, _ := strconv.ParseInt(row["seq"], 10, 32)
				ttl, _ := strconv.ParseInt(row["ttl"], 10, 32)
				loss, _ := strconv.ParseInt(row["packet-loss"], 10, 32)

				item := PingStreamItem{
					Seq:           int32(seq),
					Host:          host,
					LatencyMS:     lat,
					Status:        status,
					TTL:           int32(ttl),
					PacketLoss:    int32(loss),
					TimestampUnix: time.Now().Unix(),
				}

				if err := onResult(item); err != nil {
					return err
				}
			}
		}
	}
}
