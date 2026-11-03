package mikrotik

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMikhmonGetDashboard_Validation(t *testing.T) {
	h := NewHandler(nil, nil)

	// Missing device_id
	res, out, err := h.MikhmonGetDashboard(context.Background(), nil, mikhmonGetDashboardArgs{})
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Equal(t, "error", out.Status)
	assert.Contains(t, out.Summary, "device_id is required")

	// No registry
	res, out, err = h.MikhmonGetDashboard(context.Background(), nil, mikhmonGetDashboardArgs{DeviceID: "dev1"})
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Equal(t, "error", out.Status)
	assert.Contains(t, out.Summary, "device registry not configured")
}

func TestMikhmonGenerateVoucher_Validation(t *testing.T) {
	h := NewHandler(nil, nil)

	// Missing params
	res, out, err := h.MikhmonGenerateVoucher(context.Background(), nil, mikhmonGenerateVoucherArgs{})
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Equal(t, "error", out.Status)
	assert.Contains(t, out.Summary, "device_id, profile, and positive count are required")
}

func TestMikhmonKickSession_Validation(t *testing.T) {
	h := NewHandler(nil, nil)

	// Missing params
	res, out, err := h.MikhmonKickSession(context.Background(), nil, mikhmonKickSessionArgs{})
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Equal(t, "error", out.Status)
	assert.Contains(t, out.Summary, "device_id and either session_id or user are required")
}
