package network

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

type TrafficStreamItem struct {
	Interface     string
	RxBps         int64
	TxBps         int64
	TimestampUnix int64
}

// StreamTraffic orchestrates realtime interface traffic monitoring on a device driver.
func StreamTraffic(ctx context.Context, driver port.DeviceDriver, iface string, onResult func(item TrafficStreamItem) error) error {
	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return command.ErrDriverNotStreaming
	}

	if iface == "" {
		iface = "ether1"
	}

	cmd := command.Command{
		Raw: "/interface/monitor-traffic",
		Args: map[string]string{
			"interface": iface,
			"once":      "",
		},
	}

	handle, err := sd.Stream(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to initiate traffic stream: %w", err)
	}
	defer func() { _ = handle.Cancel() }()

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

			row := res.Rows[0]
			rx, _ := strconv.ParseInt(row["rx-bits-per-second"], 10, 64)
			tx, _ := strconv.ParseInt(row["tx-bits-per-second"], 10, 64)

			item := TrafficStreamItem{
				Interface:     iface,
				RxBps:         rx,
				TxBps:         tx,
				TimestampUnix: time.Now().Unix(),
			}

			if err := onResult(item); err != nil {
				return err
			}
		}
	}
}
