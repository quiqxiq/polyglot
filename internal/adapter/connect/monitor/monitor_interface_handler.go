package monitor

import (
	"context"
	"strconv"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/iface"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// StreamTraffic streams /interface/monitor-traffic for one interface by name.
func (h *NetworkMonitorConnectHandler) StreamTraffic(ctx context.Context, req *connect.Request[devicepb.StreamTrafficRequest], stream *connect.ServerStream[devicepb.TrafficStreamData]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	ifaceName := req.Msg.Interface
	if ifaceName == "" {
		ifaceName = "ether1"
	}

	cmd := iface.NewMonitorTrafficStreamCommand(ifaceName)
	handle, err := sd.Stream(ctx, cmd)
	if err != nil {
		return response.MapDomainError(err)
	}
	defer func() { _ = handle.Cancel() }()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			stats := iface.ParseInterfaceTrafficStats(res)
			rx, _ := strconv.ParseInt(stats.RxBitsPerSecond, 10, 64)
			tx, _ := strconv.ParseInt(stats.TxBitsPerSecond, 10, 64)

			err := stream.Send(&devicepb.TrafficStreamData{
				DeviceId:      req.Msg.DeviceId,
				Interface:     ifaceName,
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

// StreamInterfaceEthernet streams /interface/ethernet/print interval=<n>.
func (h *NetworkMonitorConnectHandler) StreamInterfaceEthernet(ctx context.Context, req *connect.Request[devicepb.StreamInterfaceEthernetRequest], stream *connect.ServerStream[devicepb.InterfaceEthernetFrame]) error {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return err
	}

	sd, ok := driver.(port.StreamingDeviceDriver)
	if !ok {
		return response.Unimplemented("driver does not support streaming")
	}

	interval := req.Msg.Interval
	if interval == "" {
		interval = "1s"
	}

	handle, err := sd.Stream(ctx, iface.NewStreamInterfacesCommand(req.Msg.TypeFilter, req.Msg.NameFilter, interval))
	if err != nil {
		return response.MapDomainError(err)
	}
	defer func() { _ = handle.Cancel() }()

	for {
		select {
		case <-ctx.Done():
			return nil
		case res, ok := <-handle.Chan():
			if !ok {
				return handle.Err()
			}
			ifaces := iface.ParseInterfaces(res)
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
