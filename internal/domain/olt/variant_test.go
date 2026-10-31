package olt

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectVariant(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		sysOID string
		descr  string
		fw     string
		want   Variant
	}{
		{"ha7302 firmware signature", "1.3.6.1.4.1.25355", "", "v7.76/HA7302CSM/SN2020", VariantHA7302},
		{"ha7302 descr", "1.3.6.1.4.1.25355", "HA-7302 OLT", "", VariantHA7302},
		{"ha7304 firmware", "1.3.6.1.4.1.25355", "", "1.0.0.1/HA7304/SN2018", VariantHA7304},
		{"enterprise only", "1.3.6.1.4.1.25355.3.1.1", "", "", VariantHA7304},
		{"unknown vendor", "1.3.6.1.4.1.999", "Other OLT", "", VariantUnknown},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, DetectVariant(tc.sysOID, tc.descr, tc.fw))
		})
	}
}

func TestCapabilities(t *testing.T) {
	t.Parallel()
	ha7302 := CapabilitiesFor(VariantHA7302)
	assert.True(t, ha7302.FlatLLID)
	assert.True(t, ha7302.RenameViaSNMP)

	ha7304 := CapabilitiesFor(VariantHA7304)
	assert.False(t, ha7304.FlatLLID)
	assert.Equal(t, 4, ha7304.PONCount)
}

func TestValidateONUName(t *testing.T) {
	t.Parallel()
	require.NoError(t, ValidateONUName("CUST-001"))
	require.NoError(t, ValidateONUName("KIOS_R28"))
	assert.ErrorIs(t, ValidateONUName("nama dengan spasi"), ErrInvalidInput)
	assert.ErrorIs(t, ValidateONUName(""), ErrInvalidInput)
	assert.ErrorIs(t, ValidateONUName("abc;drop"), ErrInvalidInput)
}

func TestValidatePosition(t *testing.T) {
	t.Parallel()
	ha7302 := CapabilitiesFor(VariantHA7302)
	require.NoError(t, ValidatePosition(ha7302, 1, 5))
	assert.ErrorIs(t, ValidatePosition(ha7302, 2, 5), ErrInvalidInput)
	assert.ErrorIs(t, ValidatePosition(ha7302, 1, 129), ErrInvalidInput)

	ha7304 := CapabilitiesFor(VariantHA7304)
	require.NoError(t, ValidatePosition(ha7304, 4, 128))
	assert.ErrorIs(t, ValidatePosition(ha7304, 5, 1), ErrInvalidInput)
}

func TestErrorsAreSentinels(t *testing.T) {
	t.Parallel()
	assert.True(t, errors.Is(ErrUnreachable, ErrUnreachable))
	assert.True(t, errors.Is(ErrCommandRejected, ErrCommandRejected))
}
