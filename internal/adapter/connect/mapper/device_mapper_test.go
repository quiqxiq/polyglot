package mapper_test

import (
	"testing"

	devicepb "github.com/quixiq/polyglot/api/proto/v1"
	"github.com/quixiq/polyglot/internal/adapter/connect/mapper"
	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/usecase/business"
)

func TestDeviceToProto(t *testing.T) {
	d := device.Device{
		ID:             "dev-1",
		TenantID:       "tenant-1",
		Name:           "Core Router",
		Vendor:         "mikrotik",
		DriverType:     "mikrotik",
		Host:           "192.168.88.1",
		Port:           8728,
		SSHPort:        22,
		TimeoutMS:      5000,
		PollIntervalMS: 10000,
		Tags:           []string{"core", "gateway"},
		Enabled:        true,
	}

	var pb *devicepb.Device = mapper.DeviceToProto(d)
	if pb == nil {
		t.Fatal("expected non-nil protobuf device")
	}

	if pb.Id != d.ID || pb.Name != d.Name || pb.Host != d.Host || pb.Port != int32(d.Port) {
		t.Errorf("field mismatch between domain and proto: got %+v, want %+v", pb, d)
	}

	// Test back conversion
	domain := mapper.ProtoToDevice(pb)
	if domain.ID != d.ID || domain.Name != d.Name || domain.Port != d.Port {
		t.Errorf("field mismatch in roundtrip: got %+v, want %+v", domain, d)
	}
}

func TestConnectionTestToProto(t *testing.T) {
	res := business.DeviceTestResult{
		DeviceID:  "dev-1",
		Status:    "connected",
		LatencyMS: 15,
		Uptime:    "5d12h",
		Version:   "7.12",
		BoardName: "RB4011",
		Identity:  "Router-Gateway",
		Message:   "connection established",
	}

	pb := mapper.ConnectionTestToProto(res)
	if pb == nil {
		t.Fatal("expected non-nil response")
	}

	if !pb.Success {
		t.Errorf("expected Success=true for status 'connected'")
	}
	if pb.DeviceId != res.DeviceID || pb.BoardName != res.BoardName {
		t.Errorf("field mismatch: got %+v, want %+v", pb, res)
	}
}
