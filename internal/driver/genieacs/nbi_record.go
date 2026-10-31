package genieacs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DeviceRecord is the normalized view of one GenieACS device document used by
// ACS inventory sync and ACS↔ONT matching. It flattens the parts of the
// document the NBI does not conveniently expose — the CPE identity logged in
// the _deviceId object and the TR-069 parameter tree — into plain fields.
type DeviceRecord struct {
	// ID is the GenieACS document ID (the _id field), for example
	// "00236A-ONT-SN1". It is the identifier other NBI endpoints expect.
	ID string
	// DeviceID is the canonical "OUI-ProductClass-SerialNumber" identity,
	// rebuilt from _deviceId and/or a split of _id.
	DeviceID string
	// OUI is the TR-069 device OUI (the first segment of DeviceID).
	OUI string
	// ProductClass is the TR-069 product class (the second DeviceID segment).
	ProductClass string
	// SerialNumber is the TR-069 serial number (the third DeviceID segment,
	// which may itself contain hyphens).
	SerialNumber string
	// SoftwareVersion is the reported DeviceInfo.SoftwareVersion, resolved
	// from the TR-098 or TR-181 parameter tree when present.
	SoftwareVersion string
	// LastInform is the last time the CPE contacted the ACS; zero when the
	// document carries no parseable _lastInform.
	LastInform time.Time
	// Tags mirrors the NBI _tags array (used to flag managed/bound devices).
	Tags []string
	// Raw is the untouched device document as returned by the NBI, preserved
	// for persistence in acs_devices.raw by the adapter layer.
	Raw json.RawMessage
	// Params holds every scalar leaf of the document flattened to dotted
	// paths, for example
	// "InternetGatewayDevice.DeviceInfo.SoftwareVersion._value". Parameter
	// leaves keep the NBI's _value suffix; arrays are stringified.
	Params map[string]string
}

// deviceMeta is the document envelope needed for indexing, decoded
// separately from the parameter tree.
type deviceMeta struct {
	ID       string `json:"_id"`
	DeviceID struct {
		OUI          string `json:"_OUI"`
		ProductClass string `json:"_ProductClass"`
		SerialNumber string `json:"_SerialNumber"`
	} `json:"_deviceId"`
	LastInform json.RawMessage `json:"_lastInform"`
	Tags       []string        `json:"_tags"`
}

// parseDeviceRecord turns one raw NBI device document into a DeviceRecord.
// Missing identity fields are resolved from the canonical _id and then from
// the TR-069 parameter tree; unknown or malformed timestamps degrade to the
// zero time rather than failing the whole inventory sync.
func parseDeviceRecord(raw json.RawMessage) (DeviceRecord, error) {
	var meta deviceMeta
	if err := json.Unmarshal(raw, &meta); err != nil {
		return DeviceRecord{}, fmt.Errorf("decode device document: %w", err)
	}

	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return DeviceRecord{}, fmt.Errorf("decode device parameters: %w", err)
	}
	params := make(map[string]string, len(document))
	flattenJSON(document, "", params)

	record := DeviceRecord{
		ID:           meta.ID,
		OUI:          meta.DeviceID.OUI,
		ProductClass: meta.DeviceID.ProductClass,
		SerialNumber: meta.DeviceID.SerialNumber,
		LastInform:   parseNBITime(meta.LastInform),
		Tags:         meta.Tags,
		Raw:          append(json.RawMessage(nil), raw...),
		Params:       params,
	}

	// _id is canonical OUI-ProductClass-SerialNumber: use it to fill any
	// identity field the _deviceId object omitted (some NBI versions return
	// a reduced envelope, especially with projections).
	if oui, productClass, serial, ok := splitDeviceID(meta.ID); ok {
		if record.OUI == "" {
			record.OUI = oui
		}
		if record.ProductClass == "" {
			record.ProductClass = productClass
		}
		if record.SerialNumber == "" {
			record.SerialNumber = serial
		}
	}

	// Devices with a non-standard or missing _deviceId still report identity
	// in the TR-069 tree; prefer the document field and fall back to the tree.
	if record.SerialNumber == "" {
		record.SerialNumber = paramValue(params,
			"InternetGatewayDevice.DeviceInfo.SerialNumber",
			"Device.DeviceInfo.SerialNumber")
	}
	if record.ProductClass == "" {
		record.ProductClass = paramValue(params,
			"InternetGatewayDevice.DeviceInfo.ProductClass",
			"Device.DeviceInfo.ProductClass")
	}
	record.SoftwareVersion = paramValue(params,
		"InternetGatewayDevice.DeviceInfo.SoftwareVersion",
		"Device.DeviceInfo.SoftwareVersion")

	record.DeviceID = canonicalDeviceID(record.OUI, record.ProductClass, record.SerialNumber)
	if record.DeviceID == "" {
		record.DeviceID = meta.ID
	}
	return record, nil
}

// splitDeviceID parses the canonical GenieACS device ID
// "OUI-ProductClass-SerialNumber". Serial numbers may include hyphens, so at
// most three fields are split.
func splitDeviceID(deviceID string) (oui, productClass, serial string, ok bool) {
	parts := strings.SplitN(deviceID, "-", 3)
	if len(parts) != 3 {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}

// canonicalDeviceID rebuilds the device ID only when every segment is known;
// a partial identity would be ambiguous and must not be presented as a
// GenieACS device ID.
func canonicalDeviceID(oui, productClass, serial string) string {
	if oui == "" || productClass == "" || serial == "" {
		return ""
	}
	return oui + "-" + productClass + "-" + serial
}

// paramValue returns the first non-empty value among candidate TR-069 paths.
// GenieACS stores parameter leaves as objects with the value under _value, so
// the flattened map normally carries "<path>._value"; a bare "<path>" scalar
// is also accepted for NBI versions or projections that return the raw value.
func paramValue(params map[string]string, paths ...string) string {
	for _, path := range paths {
		if value := params[path+"._value"]; value != "" {
			return value
		}
		if value := params[path]; value != "" {
			return value
		}
	}
	return ""
}

// parseNBITime parses the _lastInform timestamp. The NBI serializes dates as
// ISO-8601 strings; a millisecond epoch number and the {$date: ...}
// extended-JSON envelope are tolerated for deployments that vary.
func parseNBITime(raw json.RawMessage) time.Time {
	if len(raw) == 0 || string(raw) == "null" {
		return time.Time{}
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return parseNBITimeString(text)
	}

	var epochMillis float64
	if err := json.Unmarshal(raw, &epochMillis); err == nil && epochMillis > 0 {
		return time.UnixMilli(int64(epochMillis)).UTC()
	}

	var extended struct {
		Date json.RawMessage `json:"$date"`
	}
	if err := json.Unmarshal(raw, &extended); err == nil && len(extended.Date) > 0 {
		if err := json.Unmarshal(extended.Date, &text); err == nil {
			return parseNBITimeString(text)
		}
		if err := json.Unmarshal(extended.Date, &epochMillis); err == nil && epochMillis > 0 {
			return time.UnixMilli(int64(epochMillis)).UTC()
		}
	}
	return time.Time{}
}

// nbiTimeLayouts are the timestamp shapes accepted for device documents. The
// space-separated layout is the one GenieACS uses inside query filters; it is
// accepted here so stored fixtures copied from queries still parse.
var nbiTimeLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05",
}

// parseNBITimeString parses a timestamp string, returning the zero time when
// the value matches none of the known layouts.
func parseNBITimeString(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	for _, layout := range nbiTimeLayouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}
