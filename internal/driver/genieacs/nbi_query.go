package genieacs

import (
	"encoding/json"
	"strings"
)

// noMatchQuery is emitted by the query helpers when the identifying value is
// blank. It matches no device, so a missing serial or MAC can never degrade
// into "list every device" (the NBI treats an absent query as no filter).
const noMatchQuery = `{"_id":{"$in":[]}}`

// macParameterPaths lists the TR-069 parameters commonly carrying the ONT's
// own MAC (WAN, LAN, or the radio BSSID printed on the device label). MAC
// addresses are not part of the GenieACS device ID, so they can only be
// queried through parameter values. Models with different paths need a
// custom NBIListOptions.Query.
var macParameterPaths = []string{
	"InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANIPConnection.1.MACAddress",
	"InternetGatewayDevice.WANDevice.1.WANConnectionDevice.1.WANPPPConnection.1.MACAddress",
	"InternetGatewayDevice.LANDevice.1.LANEthernetInterfaceConfig.1.MACAddress",
	"InternetGatewayDevice.LANDevice.1.WLANConfiguration.1.BSSID",
	"Device.Ethernet.Interface.1.MACAddress",
	"Device.WiFi.SSID.1.BSSID",
}

// DeviceBySerialQuery builds a MongoDB query JSON that locates a device by
// serial number. It covers the GenieACS device-ID envelope
// (_deviceId._SerialNumber) as well as the TR-098 and TR-181 DeviceInfo
// parameters, because some CPEs report a serial in the parameter tree that
// differs from the one used in their device ID.
//
// The result is safe to pass as NBIListOptions.Query: it is compact JSON and
// the NBI URL-encodes it. A blank serial yields a match-nothing query.
func DeviceBySerialQuery(serial string) string {
	serial = strings.TrimSpace(serial)
	if serial == "" {
		return noMatchQuery
	}
	return marshalQuery(map[string]any{
		"$or": []map[string]string{
			{"_deviceId._SerialNumber": serial},
			{"InternetGatewayDevice.DeviceInfo.SerialNumber": serial},
			{"Device.DeviceInfo.SerialNumber": serial},
		},
	})
}

// DeviceByMACQuery builds a MongoDB query JSON that locates a device by MAC
// address. MACs are matched case-insensitively by expanding the input to its
// distinct lower/upper-case forms, because CPEs report either casing and the
// NBI does exact value matching.
//
// A blank MAC yields a match-nothing query.
func DeviceByMACQuery(mac string) string {
	mac = strings.TrimSpace(mac)
	if mac == "" {
		return noMatchQuery
	}

	variants := []string{mac}
	if lower := strings.ToLower(mac); lower != mac {
		variants = append(variants, lower)
	}
	if upper := strings.ToUpper(mac); upper != mac {
		variants = append(variants, upper)
	}

	clauses := make([]map[string]any, 0, len(macParameterPaths))
	for _, path := range macParameterPaths {
		clauses = append(clauses, map[string]any{
			path: map[string]any{"$in": variants},
		})
	}
	return marshalQuery(map[string]any{"$or": clauses})
}

// deviceIDQuery builds the exact-match MongoDB query for one GenieACS device
// ID or task ID (both are _id fields on their collections).
func deviceIDQuery(id string) string {
	return marshalQuery(map[string]string{"_id": id})
}

// marshalQuery renders a query for the NBI query parameter. Every caller
// builds the query from strings and maps, so marshalling cannot fail; a
// match-nothing query is returned defensively so an impossible failure can
// never widen a lookup into a full collection scan.
func marshalQuery(query any) string {
	encoded, err := json.Marshal(query)
	if err != nil {
		return noMatchQuery
	}
	return string(encoded)
}
