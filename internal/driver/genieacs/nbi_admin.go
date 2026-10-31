package genieacs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

// PutPreset creates or replaces a GenieACS preset (an idempotent upsert keyed
// by name) used to bootstrap devices: periodic inform interval, management
// server URL, and similar provisioning configuration.
//
// preset is the preset object documented by GenieACS
// ({"weight":..., "precondition":..., "configurations":[...]}). The name in
// the URL always wins over any name-like field in the body, matching the NBI's
// own behavior.
func (n *NBI) PutPreset(ctx context.Context, name string, preset any) error {
	if err := requireArgument("preset name", name); err != nil {
		return err
	}
	if preset == nil {
		return fmt.Errorf("%w: preset body", ErrEmptyArgument)
	}

	payload, err := json.Marshal(preset)
	if err != nil {
		return fmt.Errorf("genieacs: put preset %q: marshal body: %w", name, err)
	}
	if err := n.doAcknowledged(ctx, http.MethodPut, "/presets/"+url.PathEscape(name), payload); err != nil {
		return fmt.Errorf("genieacs: put preset %q: %w", name, err)
	}
	return nil
}

// DeletePreset removes a preset by name. Deleting a preset that does not exist
// is a no-op at the NBI: it still acknowledges success.
func (n *NBI) DeletePreset(ctx context.Context, name string) error {
	if err := requireArgument("preset name", name); err != nil {
		return err
	}
	if err := n.doAcknowledged(ctx, http.MethodDelete, "/presets/"+url.PathEscape(name), nil); err != nil {
		return fmt.Errorf("genieacs: delete preset %q: %w", name, err)
	}
	return nil
}

// ListPresets returns every preset known to the ACS, keyed by preset name
// (the NBI _id) with the untouched preset document as value. Callers needing
// only drift detection can compare against PutPreset bodies without
// redefining the preset schema in Go.
func (n *NBI) ListPresets(ctx context.Context) (map[string]json.RawMessage, error) {
	resp, err := n.do(ctx, http.MethodGet, "/presets/", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("genieacs: list presets: %w", err)
	}
	defer closeResponse(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("genieacs: list presets: %w", statusError(resp))
	}

	var documents []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&documents); err != nil {
		return nil, fmt.Errorf("genieacs: list presets: decode response: %w", err)
	}

	presets := make(map[string]json.RawMessage, len(documents))
	for _, document := range documents {
		var meta struct {
			ID string `json:"_id"`
		}
		if err := json.Unmarshal(document, &meta); err != nil {
			return nil, fmt.Errorf("genieacs: list presets: decode preset: %w", err)
		}
		if meta.ID == "" {
			continue
		}
		presets[meta.ID] = document
	}
	return presets, nil
}

// ListFaults returns the fault messages currently recorded for a device across
// all fault channels. A fault means the CPE rejected or failed the last task
// on that channel; the adapter surfaces the messages so an operator can see
// why a remote action did not take effect.
//
// GenieACS keys faults by "<device_id>:<channel>", so the lookup is a prefix
// match with the device ID regex-escaped. When a fault document has no message
// its fault ID is returned instead of dropping it silently.
func (n *NBI) ListFaults(ctx context.Context, deviceID string) ([]string, error) {
	if err := requireArgument("device id", deviceID); err != nil {
		return nil, err
	}

	query := marshalQuery(map[string]any{
		"_id": map[string]string{"$regex": "^" + regexp.QuoteMeta(deviceID) + ":"},
	})
	q := url.Values{}
	q.Set("query", query)

	resp, err := n.do(ctx, http.MethodGet, "/faults/", q, nil)
	if err != nil {
		return nil, fmt.Errorf("genieacs: list faults for %q: %w", deviceID, err)
	}
	defer closeResponse(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("genieacs: list faults for %q: %w", deviceID, statusError(resp))
	}

	var faults []struct {
		ID      string `json:"_id"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&faults); err != nil {
		return nil, fmt.Errorf("genieacs: list faults for %q: decode response: %w", deviceID, err)
	}

	messages := make([]string, 0, len(faults))
	for _, fault := range faults {
		if fault.Message != "" {
			messages = append(messages, fault.Message)
			continue
		}
		messages = append(messages, fault.ID)
	}
	return messages, nil
}

// DeleteFault clears one fault by its NBI fault ID ("<device_id>:<channel>"),
// for example after an operator retried the task deliberately.
func (n *NBI) DeleteFault(ctx context.Context, faultID string) error {
	if err := requireArgument("fault id", faultID); err != nil {
		return err
	}
	if err := n.doAcknowledged(ctx, http.MethodDelete, "/faults/"+url.PathEscape(faultID), nil); err != nil {
		return fmt.Errorf("genieacs: delete fault %q: %w", faultID, err)
	}
	return nil
}

// DeleteDevice removes a device from the ACS database. Note that GenieACS
// re-registers the device on its next inform, so this is only useful for
// cleaning up stale records whose CPE is gone.
func (n *NBI) DeleteDevice(ctx context.Context, deviceID string) error {
	if err := requireArgument("device id", deviceID); err != nil {
		return err
	}
	if err := n.doAcknowledged(ctx, http.MethodDelete, "/devices/"+url.PathEscape(deviceID), nil); err != nil {
		return fmt.Errorf("genieacs: delete device %q: %w", deviceID, err)
	}
	return nil
}

// AddTag assigns a tag to a device (a no-op if the device already has it).
// Polyglot uses tags to mark devices it manages and to target presets.
func (n *NBI) AddTag(ctx context.Context, deviceID, tag string) error {
	if err := requireDeviceAndTag(deviceID, tag); err != nil {
		return err
	}
	if err := n.doAcknowledged(ctx, http.MethodPost, tagPath(deviceID, tag), nil); err != nil {
		return fmt.Errorf("genieacs: add tag %q to device %q: %w", tag, deviceID, err)
	}
	return nil
}

// RemoveTag removes a tag from a device (a no-op if the device does not have
// it).
func (n *NBI) RemoveTag(ctx context.Context, deviceID, tag string) error {
	if err := requireDeviceAndTag(deviceID, tag); err != nil {
		return err
	}
	if err := n.doAcknowledged(ctx, http.MethodDelete, tagPath(deviceID, tag), nil); err != nil {
		return fmt.Errorf("genieacs: remove tag %q from device %q: %w", tag, deviceID, err)
	}
	return nil
}

// tagPath builds the NBI tag resource path. Both segments are escaped because
// GenieACS device IDs can contain characters that are special in URL paths.
func tagPath(deviceID, tag string) string {
	return "/devices/" + url.PathEscape(deviceID) + "/tags/" + url.PathEscape(tag)
}

// requireDeviceAndTag validates both identifiers used by the tag endpoints.
func requireDeviceAndTag(deviceID, tag string) error {
	if err := requireArgument("device id", deviceID); err != nil {
		return err
	}
	return requireArgument("tag", tag)
}
