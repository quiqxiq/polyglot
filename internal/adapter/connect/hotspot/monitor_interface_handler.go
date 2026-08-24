package hotspot

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamTraffic streams /interface/monitor-traffic for one interface by name
// (RouterOS pushes a new rate row every second natively — no backend
// polling). Replaces legacy get_traffic polling.
func (h *HotspotConnectHandler) StreamTraffic(ctx context.Context, req *connect.Request[devicepb.StreamTrafficRequest], stream *connect.ServerStream[devicepb.TrafficStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	iface := req.Msg.Interface
	if iface == "" {
		iface = "ether1"
	}

	cmd := mikrotik.NewMonitorTrafficStreamCommand(iface)
	handle, err := sd.Stream(ctx, cmd)
	if err != nil {
		return response.MapDomainError(err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			stats := mikrotik.ParseInterfaceTrafficStats(res)
			rx, _ := strconv.ParseInt(stats.RxBitsPerSecond, 10, 64)
			tx, _ := strconv.ParseInt(stats.TxBitsPerSecond, 10, 64)

			err := stream.Send(&devicepb.TrafficStreamData{
				DeviceId:      req.Msg.DeviceId,
				Interface:     iface,
				RxBps:         rx,
				TxBps:         tx,
				TimestampUnix: time.Now().Unix(),
			})
			if err != nil {
				return err
			}
		}
	}
}

// StreamInterfaceEthernet streams /interface/ethernet/print interval=<n>,
// forwarding the full list of ethernet interfaces on every tick.
func (h *HotspotConnectHandler) StreamInterfaceEthernet(ctx context.Context, req *connect.Request[devicepb.StreamInterfaceEthernetRequest], stream *connect.ServerStream[devicepb.InterfaceEthernetFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return connect.NewError(connect.CodeUnimplemented, fmt.Errorf("driver does not support streaming"))
	}

	interval := req.Msg.Interval
	if interval == "" {
		interval = "1s"
	}

	handle, err := sd.Stream(ctx, mikrotik.NewStreamInterfacesCommand(req.Msg.TypeFilter, req.Msg.NameFilter, interval))
	if err != nil {
		return response.MapDomainError(err)
	}
	defer handle.Cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			ifaces := mikrotik.ParseInterfaces(res)
			items := make([]*devicepb.InterfaceEthernetItem, 0, len(ifaces))
			for _, ifc := range ifaces {
				items = append(items, &devicepb.InterfaceEthernetItem{
					Id:         ifc.RosID,
					Name:       ifc.Name,
					Type:       ifc.Type,
					Mtu:        ifc.MTU,
					MacAddress: ifc.MACAddress,
					Running:    ifc.Running,
					Disabled:   ifc.Disabled,
					RxByte:     ifc.RxByte,
					TxByte:     ifc.TxByte,
					RxPacket:   ifc.RxPacket,
					TxPacket:   ifc.TxPacket,
					Comment:    ifc.Comment,
				})
			}

			err := stream.Send(&devicepb.InterfaceEthernetFrame{
				DeviceId:      req.Msg.DeviceId,
				TimestampUnix: time.Now().Unix(),
				Interfaces:    items,
			})
			if err != nil {
				return err
			}
		}
	}
}
