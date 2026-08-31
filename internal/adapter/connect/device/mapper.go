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
	cfg := d.PingConfig()
	return &devicepb.Device{
		Id:                d.ID,
		TenantId:          d.TenantID,
		Name:              d.Name,
		Vendor:            d.Vendor,
		DriverType:        d.DriverType,
		Host:              d.Host,
		Port:              int32(d.Port),
		SshPort:           sshPort,
		TimeoutMs:         int32(d.TimeoutMS),
		PollIntervalMs:    int32(d.PollIntervalMS),
		Tags:              d.Tags,
		Enabled:           d.Enabled,
		PingEnabled:       cfg.Enabled,
		PingTarget:        cfg.Target,
		PingRetentionDays: int32(cfg.RetentionDays),
		Extra:             d.Extra,
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
	extra := make(map[string]string)
	for k, v := range pb.Extra {
		extra[k] = v
	}
	dev := device.Device{
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
		Extra:          extra,
	}
	if pb.PingTarget != "" || pb.PingEnabled || pb.PingRetentionDays > 0 {
		dev.SetPingConfig(device.DevicePingConfig{
			Enabled:       pb.PingEnabled,
			Target:        pb.PingTarget,
			RetentionDays: int(pb.PingRetentionDays),
		})
	}
	return dev
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

func parsePingLatency(row map[string]string) (int64, string) {
	return ping.ParsePingLatency(row)
}

// IsolationConfigToDomain converts proto isolation config to domain model.
func IsolationConfigToDomain(pb *devicepb.IsolationConfig) device.IsolationConfig {
	if pb == nil {
		return device.DefaultIsolationConfig()
	}
	return device.IsolationConfig{
		PPPoEProfileName:    pb.PppoeProfileName,
		HotspotProfileName:  pb.HotspotProfileName,
		AddressListName:     pb.AddressListName,
		RateLimit:           pb.RateLimit,
		LocalAddress:        pb.LocalAddress,
		RemoteAddressPool:   pb.RemoteAddressPool,
		RedirectIP:          pb.RedirectIp,
		RedirectPort:        int(pb.RedirectPort),
		NATRedirectEnabled:  pb.NatRedirectEnabled,
		PPPoERedirectURL:    pb.PppoeRedirectUrl,
		HotspotRedirectURL:  pb.HotspotRedirectUrl,
		WalledGardenDomains: pb.WalledGardenDomains,
	}
}

// IsolationStatusToPb converts domain isolation status to proto message.
func IsolationStatusToPb(s device.IsolationStatus) *devicepb.IsolationStatus {
	return &devicepb.IsolationStatus{
		PppoeProfileExists:   s.PPPoEProfileExists,
		HotspotProfileExists: s.HotspotProfileExists,
		AddressListExists:    s.AddressListExists,
		NatRedirectExists:    s.NATRedirectExists,
		IsolatedUsersCount:   int32(s.IsolatedUsersCount),
		Config: &devicepb.IsolationConfig{
			PppoeProfileName:    s.Config.PPPoEProfileName,
			HotspotProfileName:  s.Config.HotspotProfileName,
			AddressListName:     s.Config.AddressListName,
			RateLimit:           s.Config.RateLimit,
			LocalAddress:        s.Config.LocalAddress,
			RemoteAddressPool:   s.Config.RemoteAddressPool,
			RedirectIp:          s.Config.RedirectIP,
			RedirectPort:        int32(s.Config.RedirectPort),
			NatRedirectEnabled:  s.Config.NATRedirectEnabled,
			PppoeRedirectUrl:    s.Config.PPPoERedirectURL,
			HotspotRedirectUrl:  s.Config.HotspotRedirectURL,
			WalledGardenDomains: s.Config.WalledGardenDomains,
		},
	}
}
