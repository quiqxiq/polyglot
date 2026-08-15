package genieacs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func (d *Driver) execGetDevice(ctx context.Context, projection string) (command.Result, error) {
	q := url.Values{}
	q.Set("query", fmt.Sprintf(`{"_id":"%s"}`, d.deviceID))
	if projection != "" {
		q.Set("projection", projection)
	}

	resp, err := d.do(ctx, http.MethodGet, "/devices/", q, nil)
	if err != nil {
		return command.Result{}, fmt.Errorf("genieacs: get device %s: %w", d.deviceID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return command.Result{}, fmt.Errorf("genieacs: get device %s: NBI %d", d.deviceID, resp.StatusCode)
	}

	var arr []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		return command.Result{}, fmt.Errorf("genieacs: decode device response: %w", err)
	}
	if len(arr) == 0 {
		return command.Result{}, fmt.Errorf("genieacs: device %s not found", d.deviceID)
	}

	row := make(map[string]string, len(arr[0]))
	flattenJSON(arr[0], "", row)
	return command.Result{Rows: []map[string]string{row}}, nil
}

func (d *Driver) fetchParamValues(ctx context.Context, parameterNamesJSON string) ([]map[string]string, error) {
	projection := projectionFromParamNames(parameterNamesJSON)
	result, err := d.execGetDevice(ctx, projection)
	if err != nil {
		return nil, err
	}
	return result.Rows, nil
}

func projectionFromParamNames(parameterNamesJSON string) string {
	if parameterNamesJSON == "" {
		return ""
	}
	var names []string
	if err := json.Unmarshal([]byte(parameterNamesJSON), &names); err != nil {
		return ""
	}
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ","
		}
		out += n
	}
	return out
}

func flattenJSON(m map[string]any, prefix string, out map[string]string) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]any:
			flattenJSON(val, key, out)
		default:
			out[key] = fmt.Sprint(val)
		}
	}
}
