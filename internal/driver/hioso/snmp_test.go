package hioso

import (
	"context"
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/olt"
)

// fakeSNMPTransport memenuhi pduTransport untuk unit test tanpa socket.
type fakeSNMPTransport struct {
	get   map[string]gosnmp.SnmpPDU
	walks map[string][]gosnmp.SnmpPDU
	err   error
}

func (f *fakeSNMPTransport) Get(oids []string) (*gosnmp.SnmpPacket, error) {
	if f.err != nil {
		return nil, f.err
	}
	packet := &gosnmp.SnmpPacket{}
	for _, oid := range oids {
		if pdu, ok := f.get[oid]; ok {
			packet.Variables = append(packet.Variables, pdu)
		}
	}
	return packet, nil
}

func (f *fakeSNMPTransport) Walk(root string) ([]gosnmp.SnmpPDU, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.walks[root], nil
}

func (f *fakeSNMPTransport) Set(pdus []gosnmp.SnmpPDU) error {
	if f.err != nil {
		return f.err
	}
	if f.get == nil {
		f.get = map[string]gosnmp.SnmpPDU{}
	}
	for _, pdu := range pdus {
		f.get[pdu.Name] = pdu
	}
	return nil
}

func (f *fakeSNMPTransport) Close() error { return nil }

func octet(name, value string) gosnmp.SnmpPDU {
	return gosnmp.SnmpPDU{Name: name, Type: gosnmp.OctetString, Value: value}
}

func integer(name string, value int) gosnmp.SnmpPDU {
	return gosnmp.SnmpPDU{Name: name, Type: gosnmp.Integer, Value: value}
}

func TestSysInfoAndFirmwareSignature(t *testing.T) {
	t.Parallel()
	tr := &fakeSNMPTransport{
		get: map[string]gosnmp.SnmpPDU{
			oidSysDescr:                 octet(oidSysDescr, "HA7302CSM"),
			oidSysObjectID:              octet(oidSysObjectID, "1.3.6.1.4.1.25355.3.1.1"),
			oidSysUptime:                {Name: oidSysUptime, Type: gosnmp.TimeTicks, Value: uint32(360000)},
			oidFirmwareSignature + ".1": octet(oidFirmwareSignature+".1", "v7.76/HA7302CSM/SN2020"),
		},
	}
	c := newSNMPClient(tr)
	sysDescr, sysObj, uptime, err := c.SysInfo(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "HA7302CSM", sysDescr)
	assert.Equal(t, "1.3.6.1.4.1.25355.3.1.1", sysObj)
	assert.Equal(t, "1h0m0s", uptime)

	fw, err := c.FirmwareSignature(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "v7.76/HA7302CSM/SN2020", fw)
}

func TestListONUsHA7304(t *testing.T) {
	t.Parallel()
	idx := func(root, suffix string) gosnmp.SnmpPDU { return octet(root+"."+suffix, "") }
	tr := &fakeSNMPTransport{walks: map[string][]gosnmp.SnmpPDU{
		oidONUName: {
			octet(oidONUName+".1.3", "CUST-001"),
			octet(oidONUName+".2.7", "CUST-002"),
		},
		oidONUMAC: {
			octet(oidONUMAC+".1.3", "78:5C:72:50:00:01"),
			octet(oidONUMAC+".2.7", "78-5c-72-50-00-02"),
		},
		oidONULinkState: {
			integer(oidONULinkState+".1.3", 1),
			integer(oidONULinkState+".2.7", 2),
		},
		oidONURxPower: {
			octet(oidONURxPower+".1.3", "-21.40"),
			octet(oidONURxPower+".2.7", "na"),
		},
		oidONUTxPower: {
			octet(oidONUTxPower+".1.3", "2.10"),
		},
		oidONUDistance: {
			{Name: oidONUDistance + ".1.3", Type: gosnmp.Integer, Value: 1234},
		},
		oidONULLID: {
			integer(oidONULLID+".1.3", 3),
		},
	}}
	_ = idx // helper sengaja tidak dipakai untuk variasi value

	c := newSNMPClient(tr)
	rows, err := c.ListONUs(context.Background(), olt.VariantHA7304, 0)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	assert.Equal(t, 1, rows[0].PONIndex)
	assert.Equal(t, 3, rows[0].ONUIndex)
	assert.Equal(t, "78:5c:72:50:00:01", rows[0].MACAddress)
	assert.Equal(t, "CUST-001", rows[0].Name)
	assert.True(t, rows[0].LinkUp)
	require.NotNil(t, rows[0].RxPowerDBM)
	assert.InDelta(t, -21.4, *rows[0].RxPowerDBM, 0.001)
	require.NotNil(t, rows[0].DistanceM)
	assert.Equal(t, 1234, *rows[0].DistanceM)
	assert.Equal(t, 3, rows[0].LLID)

	assert.Equal(t, 2, rows[1].PONIndex)
	assert.Equal(t, 7, rows[1].ONUIndex)
	assert.False(t, rows[1].LinkUp)
	assert.Nil(t, rows[1].RxPowerDBM)

	// Filter PON 2 hanya mengembalikan satu baris.
	rows, err = c.ListONUs(context.Background(), olt.VariantHA7304, 2)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, 7, rows[0].ONUIndex)
}

func TestListONUsHA7302Flat(t *testing.T) {
	t.Parallel()
	tr := &fakeSNMPTransport{walks: map[string][]gosnmp.SnmpPDU{
		oidONUName:      {octet(oidONUName+".1.5", "KIOS_R28")},
		oidONUMAC:       {octet(oidONUMAC+".1.5", "785c72500000")},
		oidONULinkState: {integer(oidONULinkState+".1.5", 1)},
	}}
	c := newSNMPClient(tr)
	rows, err := c.ListONUs(context.Background(), olt.VariantHA7302, 0)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, 1, rows[0].PONIndex, "HA7302 memakai LLID datar, PON dinormalisasi ke 1")
	assert.Equal(t, 5, rows[0].ONUIndex)
	assert.Equal(t, "KIOS_R28", rows[0].Name)
}

func TestIndexEncoding(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "2.7", indexOf(olt.VariantHA7304, 2, 7))
	assert.Equal(t, "1.7", indexOf(olt.VariantHA7302, 2, 7))

	pon, onu, ok := parseIndex(olt.VariantHA7304, "2.7")
	require.True(t, ok)
	assert.Equal(t, 2, pon)
	assert.Equal(t, 7, onu)

	pon, onu, ok = parseIndex(olt.VariantHA7302, "1.7")
	require.True(t, ok)
	assert.Equal(t, 1, pon)
	assert.Equal(t, 7, onu)
}
