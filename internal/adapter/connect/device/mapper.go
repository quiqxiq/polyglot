package device

import (
	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/domain/device"
	deviceUC "github.com/quixiq/polyglot/internal/usecase/device"
	"github.com/quixiq/polyglot/pkg/ping"
)

// DomainToPb converts domain device entity into proto device message.
func DomainToPb(d device.Device) *devicepb.Device {
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

// PbToDomain converts proto device message into domain device entity.
func PbToDomain(pb *devicepb.Device) device.Device {
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

// ToProtoInterfaceDetails converts usecase interface details to proto list.
func ToProtoInterfaceDetails(details []deviceUC.DeviceInterfaceDetail) []*devicepb.DeviceInterfaceInfo {
	list := make([]*devicepb.DeviceInterfaceInfo, 0, len(details))
	for _, ifc := range details {
		list = append(list, &devicepb.DeviceInterfaceInfo{
			Name:       ifc.Name,
			Type:       ifc.Type,
			Running:    ifc.Running,
			Disabled:   ifc.Disabled,
			MacAddress: ifc.MACAddress,
			RxBps:      ifc.RxBps,
			TxBps:      ifc.TxBps,
		})
	}
	return list
}

func domainToPb(d device.Device) *devicepb.Device {
	return DomainToPb(d)
}

func pbToDomain(pb *devicepb.Device) device.Device {
	return PbToDomain(pb)
}

func parsePingLatency(row map[string]string) (int64, string) {
	return ping.ParsePingLatency(row)
}
