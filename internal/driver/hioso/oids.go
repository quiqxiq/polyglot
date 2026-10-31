package hioso

import (
	"strconv"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/olt"
)

// Enterprise OID HiOSO EPON (HA7302/HA7304).
const enterpriseOID = "1.3.6.1.4.1.25355"

// OID standar MIB-II untuk identifikasi device.
const (
	oidSysDescr    = "1.3.6.1.2.1.1.1.0"
	oidSysObjectID = "1.3.6.1.2.1.1.2.0"
	oidSysUptime   = "1.3.6.1.2.1.1.3.0"
)

// OID private HiOSO yang dipakai inventori ONU (terverifikasi pada HA7302/HA7304).
const (
	// oidFirmwareSignature berisi string mis. "v7.76/HA7302CSM/..." atau
	// "1.0.0.1/HA7304/SN2018-03-00007" — dipakai deteksi varian.
	oidFirmwareSignature = "." + enterpriseOID + ".3.1.8.1.1.2.1"

	// Kolom tabel ONU — root OID; index ditambahkan sesuai varian.
	oidONUName      = "." + enterpriseOID + ".3.2.6.3.2.1.37.1"
	oidONUMAC       = "." + enterpriseOID + ".3.2.6.3.2.1.11.1"
	oidONULLID      = "." + enterpriseOID + ".3.2.6.3.2.1.1.1"
	oidONULinkState = "." + enterpriseOID + ".3.2.6.3.2.1.39.1"
	oidONUDistance  = "." + enterpriseOID + ".3.2.6.3.2.1.25.1"
	oidONURxPower   = "." + enterpriseOID + ".3.2.6.14.2.1.8.1"
	oidONUTxPower   = "." + enterpriseOID + ".3.2.6.14.2.1.4.1"
)

// linkStateUp adalah nilai kolom link-state untuk ONU online.
const linkStateUp = 1

// indexOf membentuk suffix index ONU pada OID kolom.
//   - HA7304: "{pon}.{onu}"
//   - HA7302: "1.{llid}" (LLID datar 1..128)
func indexOf(variant olt.Variant, ponIndex, onuIndex int) string {
	if variant == olt.VariantHA7302 {
		return "1." + strconv.Itoa(onuIndex)
	}
	return strconv.Itoa(ponIndex) + "." + strconv.Itoa(onuIndex)
}

// parseIndex memecah suffix OID menjadi pasangan PON/ONU.
func parseIndex(variant olt.Variant, suffix string) (ponIndex, onuIndex int, ok bool) {
	parts := strings.Split(strings.TrimPrefix(suffix, "."), ".")
	if len(parts) != 2 {
		return 0, 0, false
	}
	pon, errPon := strconv.Atoi(parts[0])
	onu, errOnu := strconv.Atoi(parts[1])
	if errPon != nil || errOnu != nil {
		return 0, 0, false
	}
	if variant == olt.VariantHA7302 {
		// Index datar: komponen pertama selalu domain LLID "1".
		pon = 1
	}
	return pon, onu, true
}
