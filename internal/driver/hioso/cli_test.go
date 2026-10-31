package hioso

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/olt"
)

// fakeCLITransport mensimulasikan sesi CLI HiOSO (mode config/enable).
type fakeCLITransport struct {
	commands []string
	outputs  map[string]string
	closed   bool
}

func (f *fakeCLITransport) Run(_ context.Context, command string) (string, error) {
	f.commands = append(f.commands, command)
	if out, ok := f.outputs[command]; ok {
		return out + "\nEPON#", nil
	}
	return "EPON#", nil
}

func (f *fakeCLITransport) Close() error { f.closed = true; return nil }

func TestCLIListONUsParsing(t *testing.T) {
	t.Parallel()
	tr := &fakeCLITransport{outputs: map[string]string{
		"show onu info epon 0/1 all": strings.Join([]string{
			"OnuIndex   Llid   Name        MacAddress          State   Distance   RxPower   TxPower",
			"0/1/1      1      CUST-001    78:5c:72:50:00:01   Up      1234       -21.40    2.10",
			"0/1/2      2      CUST-002    78:5c:72:50:00:02   Down    0          na        na",
		}, "\n"),
	}}
	cli := newCLIClient(tr)
	rows, err := cli.ListONUs(context.Background(), olt.CapabilitiesFor(olt.VariantHA7304), 4)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, "CUST-001", rows[0].Name)
	assert.Equal(t, "78:5c:72:50:00:01", rows[0].MACAddress)
	assert.True(t, rows[0].LinkUp)
	assert.Equal(t, 1, rows[0].PONIndex)
	assert.Equal(t, 1, rows[0].ONUIndex)
	require.NotNil(t, rows[0].RxPowerDBM)
	assert.InDelta(t, -21.4, *rows[0].RxPowerDBM, 0.001)
	assert.False(t, rows[1].LinkUp)
	assert.Nil(t, rows[1].RxPowerDBM)
	// Hanya 1 PON untuk HA7302, 4 PON untuk HA7304 → 4 perintah show.
	assert.Len(t, tr.commands, 4)
}

func TestCLIRenameAndRebootHA7304(t *testing.T) {
	t.Parallel()
	tr := &fakeCLITransport{outputs: map[string]string{}}
	cli := newCLIClient(tr)
	caps := olt.CapabilitiesFor(olt.VariantHA7304)
	require.NoError(t, cli.RenameONU(context.Background(), caps, 2, 5, "CUST-009"))
	assert.Contains(t, tr.commands, "conf t")
	assert.Contains(t, tr.commands, "interface epon 0/2")
	assert.Contains(t, tr.commands, "onu 5 name CUST-009")

	tr.commands = nil
	require.NoError(t, cli.RebootONU(context.Background(), caps, 1, 3))
	assert.Contains(t, tr.commands, "onu 3 reboot")
}

func TestCLIRenameHA7302Rejected(t *testing.T) {
	t.Parallel()
	cli := newCLIClient(&fakeCLITransport{})
	err := cli.RenameONU(context.Background(), olt.CapabilitiesFor(olt.VariantHA7302), 1, 5, "X")
	require.Error(t, err)
	assert.True(t, errors.Is(err, olt.ErrUnsupportedVariant), "rename HA7302 harus via SNMP SET")
}

func TestCLICommandRejectedError(t *testing.T) {
	t.Parallel()
	tr := &fakeCLITransport{outputs: map[string]string{
		"delete onu 1/1/2": "Error: ONU not found",
	}}
	cli := newCLIClient(tr)
	err := cli.DeleteONU(context.Background(), olt.CapabilitiesFor(olt.VariantHA7302), 1, 2)
	require.Error(t, err)
	assert.True(t, errors.Is(err, olt.ErrCommandRejected))
}

func TestCLISaveConfigFallback(t *testing.T) {
	t.Parallel()
	tr := &fakeCLITransport{outputs: map[string]string{
		"write file": "Invalid command",
	}}
	cli := newCLIClient(tr)
	require.NoError(t, cli.SaveConfig(context.Background()))
	assert.Contains(t, tr.commands, "write")
}
