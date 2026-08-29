package port

import (
	domainHotspot "github.com/quixiq/polyglot/internal/domain/hotspot"
)

// HotspotIPBinding alias to domain model.
type HotspotIPBinding = domainHotspot.HotspotIPBinding

// HotspotIPBindingParams represents parameters for adding or setting an IP Binding.
type HotspotIPBindingParams struct {
	MACAddress string
	Address    string
	ToAddress  string
	Server     string
	Type       string
	Comment    string
	Disabled   bool
}

// HotspotCookie alias to domain model.
type HotspotCookie = domainHotspot.HotspotCookie

// VoucherStatusDetails alias to domain model.
type VoucherStatusDetails = domainHotspot.VoucherStatusDetails
