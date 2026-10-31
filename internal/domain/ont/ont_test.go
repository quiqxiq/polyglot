package ont

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeMAC(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"78:5C:72:50:00:00", "78:5c:72:50:00:00", true},
		{"78-5c-72-50-00-00", "78:5c:72:50:00:00", true},
		{"785c72500000", "78:5c:72:50:00:00", true},
		{"785c.7250.0000", "78:5c:72:50:00:00", true},
		{"bad", "", false},
	}
	for _, tc := range cases {
		got, ok := NormalizeMAC(tc.in)
		assert.Equal(t, tc.ok, ok, tc.in)
		if tc.ok {
			assert.Equal(t, tc.want, got)
		}
	}
}

func TestDecideState(t *testing.T) {
	t.Parallel()
	assert.Equal(t, StateOnline, DecideState(StateOffline, true, 0, 2))
	assert.Equal(t, StateUnknown, DecideState(StateUnknown, false, 0, 2))
	assert.Equal(t, StateUnknown, DecideState(StateUnknown, false, 1, 2))
	assert.Equal(t, StateOffline, DecideState(StateOnline, false, 2, 2))
	// Strikes di bawah ambang mempertahankan state saat ini (anti-flap).
	assert.Equal(t, StateOnline, DecideState(StateOnline, false, 1, 2))
}

func TestParseOpticalPower(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want *float64
	}{
		{"-21.4", ptr(-21.4)},
		{"-21,4", ptr(-21.4)},
		{"\"-21.40\"", ptr(-21.4)},
		{"-21.40 dBm", ptr(-21.4)},
		{"na", nil},
		{"", nil},
	}
	for _, tc := range cases {
		got, ok := ParseOpticalPower(tc.in)
		if tc.want == nil {
			assert.False(t, ok, tc.in)
			assert.Nil(t, got)
		} else {
			require.True(t, ok, tc.in)
			assert.InDelta(t, *tc.want, *got, 0.001)
		}
	}
}

func TestClassifyRx(t *testing.T) {
	t.Parallel()
	assert.Equal(t, RxUnavailable, ClassifyRx(nil, -27, -30))
	assert.Equal(t, RxNormal, ClassifyRx(ptr(-20.0), -27, -30))
	assert.Equal(t, RxWarning, ClassifyRx(ptr(-28.0), -27, -30))
	assert.Equal(t, RxCritical, ClassifyRx(ptr(-31.0), -27, -30))
}

func TestACSEnvelopeMatching(t *testing.T) {
	t.Parallel()
	assert.True(t, MatchBySerial(" hw123 ", "HW123"))
	assert.False(t, MatchBySerial("", "HW123"))

	mac, ok := MatchExactMAC("78:5c:72:50:00:01", []string{"AA:BB:CC:DD:EE:FF", "785c72500001"})
	require.True(t, ok)
	assert.Equal(t, "78:5c:72:50:00:01", mac)

	src, matched, ok := BestMACMatch("78:5c:72:50:00:01", []string{"78:5c:72:50:00:02"})
	require.True(t, ok)
	assert.Equal(t, MatchMACNeighbor, src)
	assert.Equal(t, "78:5c:72:50:00:02", matched)

	_, _, ok = BestMACMatch("78:5c:72:50:00:01", []string{"00:11:22:33:44:55"})
	assert.False(t, ok)
}

func TestCatalogFromJSON(t *testing.T) {
	t.Parallel()
	cat, err := CatalogFromJSON(`{"rx_power_dbm":["VirtualParameters.RXPower"]}`)
	require.NoError(t, err)
	assert.Equal(t, []string{"VirtualParameters.RXPower"}, cat.RxPowerDBM)
	// Field lain tetap default.
	assert.NotEmpty(t, cat.SerialNumber)
	assert.Contains(t, cat.AllPaths(), "VirtualParameters.RXPower")
}

func ptr[T any](v T) *T { return &v }
