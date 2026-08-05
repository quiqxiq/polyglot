//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/quixiq/polyglot/internal/domain/device"
	"github.com/quixiq/polyglot/internal/driver/mikrotik"
	"github.com/quixiq/polyglot/internal/usecase/business"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceIntegration_TestConnection(t *testing.T) {
	host := os.Getenv("MIKROTIK_TEST_HOST")
	if host == "" {
		t.Skip("MIKROTIK_TEST_HOST tidak di-set — skip integration test Device TestConnection")
	}

	user := os.Getenv("MIKROTIK_TEST_USER")
	if user == "" {
		user = "admin"
	}
	pass := os.Getenv("MIKROTIK_TEST_PASS")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	target := device.Target{
		Host:     host,
		Port:     8728,
		Username: user,
		Password: pass,
		Timeout:  5 * time.Second,
	}

	driver, err := mikrotik.NewDriver(ctx, target)
	require.NoError(t, err, "Gagal terkoneksi ke MikroTik live")
	defer driver.Close()

	uc := business.NewManageDeviceUseCase(nil, nil)
	result, err := uc.TestConnection(ctx, driver, "dev-integration")
	require.NoError(t, err)

	assert.Equal(t, "dev-integration", result.DeviceID)
	assert.Equal(t, "connected", result.Status)
	assert.NotEmpty(t, result.Version, "Versi RouterOS harus ada")
	assert.NotEmpty(t, result.BoardName, "Board name harus ada")
	t.Logf("TestConnection sukses: Latency=%dms, Board=%s, Version=%s, Uptime=%s, Identity=%s",
		result.LatencyMS, result.BoardName, result.Version, result.Uptime, result.Identity)
}
