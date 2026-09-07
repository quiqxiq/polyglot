package network

import (
	"context"
	"fmt"
	"net"
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

	if hostOnly, _, err := net.SplitHostPort(host); err == nil {
		host = hostOnly
	}

	cmd := command.Command{
		Raw: "/ping",
		Args: map[string]string{
			"address": host,
		},
	}

	handle, err := sd.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to initiate ping stream: %w", err)
	}
	defer func() { _ = handle.Cancel() }()

	var streamSeq int32
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
				seqVal, _ := strconv.ParseInt(row["seq"], 10, 32)
				if seqVal == 0 {
					if s, ok := row["sequence"]; ok {
						seqVal, _ = strconv.ParseInt(s, 10, 32)
					}
					if seqVal == 0 {
						seqVal = int64(streamSeq)
					}
				}
				streamSeq = int32(seqVal + 1)

				ttl, _ := strconv.ParseInt(row["ttl"], 10, 32)
				loss := ping.ParsePacketLoss(row["packet-loss"])

				item := PingStreamItem{
					Seq:           int32(seqVal),
					Host:          host,
					LatencyMS:     lat,
					Status:        status,
					TTL:           int32(ttl),
					PacketLoss:    loss,
					TimestampUnix: time.Now().Unix(),
				}

				if err := onResult(item); err != nil {
					return err
				}
			}
		}
	}
}
