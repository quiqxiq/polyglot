package mapper

import (
	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

// DeviceToProto maps a domain Device entity to its Protobuf message representation.
func DeviceToProto(d device.Device) *devicepb.Device {
	sshPort := int32(d.SSHPort)
	if sshPort <= 0 {
		sshPort = 22
	}
	return &devicepb.Device{
		Id:             d.ID,
		TenantId:       d.TenantID,
		Name:           d.Name,
		Vendor:         d.Vendor,
		DriverType:     d.DriverType,
		Host:           d.Host,
		Port:           int32(d.Port),
		SshPort:        sshPort,
		TimeoutMs:      int32(d.TimeoutMS),
		PollIntervalMs: int32(d.PollIntervalMS),
		Tags:           d.Tags,
		Enabled:        d.Enabled,
	}
}

// DeviceListToProto maps a slice of domain Device entities to a slice of Protobuf Device messages.
func DeviceListToProto(devices []device.Device) []*devicepb.Device {
	res := make([]*devicepb.Device, len(devices))
	for i, d := range devices {
		res[i] = DeviceToProto(d)
	}
	return res
}

// ProtoToDevice maps a Protobuf Device message to a domain Device entity.
func ProtoToDevice(pb *devicepb.Device) device.Device {
	if pb == nil {
		return device.Device{}
	}
	sshPort := int(pb.SshPort)
	if sshPort <= 0 {
		sshPort = 22
	}
	return device.Device{
		ID:             pb.Id,
		TenantID:       pb.TenantId,
		Name:           pb.Name,
		Vendor:         pb.Vendor,
		DriverType:     pb.DriverType,
		Host:           pb.Host,
		Port:           int(pb.Port),
		SSHPort:        sshPort,
		TimeoutMS:      int(pb.TimeoutMs),
		PollIntervalMS: int(pb.PollIntervalMs),
		Tags:           pb.Tags,
		Enabled:        pb.Enabled,
	}
}

// ConnectionTestToProto maps a DeviceTestResult domain object to Protobuf response.
func ConnectionTestToProto(res business.DeviceTestResult) *devicepb.TestDeviceConnectionResponse {
	return &devicepb.TestDeviceConnectionResponse{
		DeviceId:  res.DeviceID,
		Status:    res.Status,
		LatencyMs: res.LatencyMS,
		Uptime:    res.Uptime,
		Version:   res.Version,
		BoardName: res.BoardName,
		Identity:  res.Identity,
		Message:   res.Message,
		Success:   res.Status == "connected" || res.Status == "online",
	}
}
