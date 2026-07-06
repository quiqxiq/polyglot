package zteolt

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/device"
)

// SNMPClient is a read-only monitoring client for ZTE OLT via gosnmp
// (github.com/gosnmp/gosnmp).
type SNMPClient struct{}

// NewSNMPClient connects to target for SNMP-based monitoring.
// target.Extra["community"] holds the SNMP community string.
// TODO: use gosnmp.Default-style setup per TECH-STACK-DAN-PERSIAPAN.md §8.
func NewSNMPClient(ctx context.Context, target device.Target) (*SNMPClient, error) {
	return &SNMPClient{}, nil
}

// Get retrieves the value of a single OID.
func (c *SNMPClient) Get(ctx context.Context, oid string) error {
	return nil
}

// Close releases the underlying SNMP socket.
func (c *SNMPClient) Close() error {
	return nil
}
