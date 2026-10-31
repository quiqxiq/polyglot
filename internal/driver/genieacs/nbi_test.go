package genieacs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fullNBIDeviceDoc = `{
	"_id": "00236A-ONT-SN1",
	"_deviceId": {"_Manufacturer": "Acme", "_OUI": "00236A", "_ProductClass": "ONT", "_SerialNumber": "SN1"},
	"_lastInform": "2024-01-02T03:04:05.000Z",
	"_tags": ["acs", "polyglot"],
	"InternetGatewayDevice": {
		"DeviceInfo": {
			"_object": true,
			"SerialNumber": {"_value": "SN1", "_type": "xsd:string", "_writable": false},
			"SoftwareVersion": {"_value": "V1.2.3", "_type": "xsd:string", "_writable": false}
		}
	}
}`

const minimalNBIDeviceDoc = `{
	"_id": "00236a-96318REF-SR360NA0A4-0003196",
	"_lastInform": null
}`

func TestNewNBI_Validation(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr bool
	}{
		{name: "http", rawURL: "http://localhost:7557"},
		{name: "https with base path", rawURL: "https://acs.example:7557/genieacs/"},
		{name: "surrounding whitespace", rawURL: "  http://127.0.0.1:7557  "},
		{name: "empty", rawURL: "", wantErr: true},
		{name: "missing scheme", rawURL: "genieacs:7557", wantErr: true},
		{name: "unsupported scheme", rawURL: "ftp://acs.example:7557", wantErr: true},
		{name: "missing host", rawURL: "http://", wantErr: true},
		{name: "query string", rawURL: "http://acs.example:7557?token=1", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewNBI(tc.rawURL, "secret-key", 3*time.Second)
			if tc.wantErr {
				require.ErrorIs(t, err, ErrInvalidNBIURL)
				assert.Nil(t, client)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, client)
			assert.Equal(t, "secret-key", client.apiKey)
			assert.Equal(t, 3*time.Second, client.httpClient.Timeout)
		})
	}
}

func TestNBI_Ping(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[]`)
	})

	require.NoError(t, client.Ping(context.Background()))

	request := recorder.last()
	assert.Equal(t, http.MethodGet, request.Method)
	assert.Equal(t, "/devices/", request.Path)
	assert.JSONEq(t, `{"_id":"__ping__"}`, request.Query.Get("query"))
	assert.Equal(t, "test-api-key", request.Header.Get("x-api-key"))
}

func TestNBI_Ping_BasePathPreserved(t *testing.T) {
	recorder := &requestRecorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder.record(r)
		writeBody(t, w, `[]`)
	}))
	defer server.Close()

	client, err := NewNBI(server.URL+"/genieacs", "", 5*time.Second)
	require.NoError(t, err)
	require.NoError(t, client.Ping(context.Background()))

	assert.Equal(t, "/genieacs/devices/", recorder.last().Path)
	assert.Empty(t, recorder.last().Header.Get("x-api-key"), "no API key header when the key is empty")
}

func TestNBI_Ping_StatusError(t *testing.T) {
	client := newTestNBI(t, &requestRecorder{}, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		writeBody(t, w, "NBI unavailable")
	})

	err := client.Ping(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NBI status 503")
	assert.Contains(t, err.Error(), "NBI unavailable")
}

func TestNBI_ListDevices(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, "["+fullNBIDeviceDoc+","+minimalNBIDeviceDoc+"]")
	})

	records, err := client.ListDevices(context.Background(), NBIListOptions{
		Query: `{"_deviceId._SerialNumber":"SN1"}`,
		Projection: []string{
			"InternetGatewayDevice.DeviceInfo.SoftwareVersion",
			" Device.DeviceInfo.SoftwareVersion ",
		},
		Skip:  10,
		Limit: 25,
	})
	require.NoError(t, err)
	require.Len(t, records, 2)

	request := recorder.last()
	assert.Equal(t, http.MethodGet, request.Method)
	assert.Equal(t, "/devices/", request.Path)
	assert.JSONEq(t, `{"_deviceId._SerialNumber":"SN1"}`, request.Query.Get("query"))
	assert.Equal(t, "InternetGatewayDevice.DeviceInfo.SoftwareVersion,Device.DeviceInfo.SoftwareVersion",
		request.Query.Get("projection"))
	assert.Equal(t, "10", request.Query.Get("skip"))
	assert.Equal(t, "25", request.Query.Get("limit"))

	full := records[0]
	assert.Equal(t, "00236A-ONT-SN1", full.ID)
	assert.Equal(t, "00236A-ONT-SN1", full.DeviceID)
	assert.Equal(t, "00236A", full.OUI)
	assert.Equal(t, "ONT", full.ProductClass)
	assert.Equal(t, "SN1", full.SerialNumber)
	assert.Equal(t, "V1.2.3", full.SoftwareVersion)
	assert.Equal(t, time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC), full.LastInform)
	assert.Equal(t, []string{"acs", "polyglot"}, full.Tags)
	assert.Equal(t, "V1.2.3", full.Params["InternetGatewayDevice.DeviceInfo.SoftwareVersion._value"])
	assert.JSONEq(t, fullNBIDeviceDoc, string(full.Raw))

	minimal := records[1]
	assert.Equal(t, "00236a-96318REF-SR360NA0A4-0003196", minimal.DeviceID)
	assert.Equal(t, "00236a", minimal.OUI)
	assert.Equal(t, "96318REF", minimal.ProductClass)
	assert.Equal(t, "SR360NA0A4-0003196", minimal.SerialNumber, "serial hyphens must not be split")
	assert.Empty(t, minimal.SoftwareVersion)
	assert.True(t, minimal.LastInform.IsZero())
}

func TestNBI_ListDevices_NoFilters(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[]`)
	})

	records, err := client.ListDevices(context.Background(), NBIListOptions{})
	require.NoError(t, err)
	assert.Empty(t, records)
	assert.Empty(t, recorder.last().Query.Encode())
}

func TestNBI_ListDevices_StatusError(t *testing.T) {
	client := newTestNBI(t, &requestRecorder{}, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		writeBody(t, w, "bad query")
	})

	_, err := client.ListDevices(context.Background(), NBIListOptions{Query: `{"_id":`})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NBI status 400")
	assert.Contains(t, err.Error(), "bad query")
}

func TestNBI_GetDevice(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[{"_id":"dev1","_deviceId":{"_OUI":"AA","_ProductClass":"PC","_SerialNumber":"SN"}}]`)
	})

	record, err := client.GetDevice(context.Background(), "dev1", []string{"Device.DeviceInfo.SoftwareVersion"})
	require.NoError(t, err)
	assert.Equal(t, "dev1", record.ID)
	assert.Equal(t, "AA-PC-SN", record.DeviceID)

	request := recorder.last()
	assert.JSONEq(t, `{"_id":"dev1"}`, request.Query.Get("query"))
	assert.Equal(t, "Device.DeviceInfo.SoftwareVersion", request.Query.Get("projection"))
	assert.Equal(t, "1", request.Query.Get("limit"))
}

func TestNBI_GetDevice_NotFound(t *testing.T) {
	client := newTestNBI(t, &requestRecorder{}, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[]`)
	})

	_, err := client.GetDevice(context.Background(), "missing", nil)
	require.ErrorIs(t, err, ErrDeviceNotFound)
	assert.Contains(t, err.Error(), "missing")
}

func TestNBI_GetDevice_RejectsEmptyID(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, failResponse)

	_, err := client.GetDevice(context.Background(), "  ", nil)
	require.ErrorIs(t, err, ErrEmptyArgument)
	assert.Zero(t, recorder.count())
}

func TestNBI_PostTask_Immediate(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeBody(t, w, `{"_id":"task-1","name":"reboot","device":"dev1"}`)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	taskID, immediate, err := client.PostTask(ctx, "dev1", map[string]any{"name": "reboot"}, true)
	require.NoError(t, err)
	assert.Equal(t, "task-1", taskID)
	assert.True(t, immediate)

	request := recorder.last()
	assert.Equal(t, http.MethodPost, request.Method)
	assert.Equal(t, "/devices/dev1/tasks", request.Path)
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "reboot", request.bodyJSON(t)["name"])
	_, hasConnectionRequest := request.Query["connection_request"]
	assert.True(t, hasConnectionRequest, "connection_request flag must be sent when requested")
	timeoutMillis, err := strconv.Atoi(request.Query.Get("timeout"))
	require.NoError(t, err)
	assert.Greater(t, timeoutMillis, 0)
	assert.LessOrEqual(t, timeoutMillis, 10_000)
}

func TestNBI_PostTask_Queued(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		writeBody(t, w, `{"_id":"task-2","name":"reboot"}`)
	})

	taskID, immediate, err := client.PostTask(context.Background(), "dev1", map[string]any{"name": "reboot"}, false)
	require.NoError(t, err)
	assert.Equal(t, "task-2", taskID)
	assert.False(t, immediate, "202 means the task is queued")

	request := recorder.last()
	_, hasConnectionRequest := request.Query["connection_request"]
	assert.False(t, hasConnectionRequest)
	assert.Equal(t, "reboot", request.bodyJSON(t)["name"])
}

func TestNBI_PostTask_EscapesDeviceID(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeBody(t, w, `{"_id":"task-1"}`)
	})

	_, _, err := client.PostTask(context.Background(), "SN%2D1", map[string]string{"name": "reboot"}, false)
	require.NoError(t, err)
	assert.Equal(t, "/devices/SN%2D1/tasks", recorder.last().Path,
		"a literal percent in a GenieACS device ID must survive path escaping")
}

func TestNBI_PostTask_StatusError(t *testing.T) {
	client := newTestNBI(t, &requestRecorder{}, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		writeBody(t, w, "task rejected")
	})

	_, _, err := client.PostTask(context.Background(), "dev1", map[string]any{"name": "reboot"}, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NBI status 500")
	assert.Contains(t, err.Error(), "task rejected")
}

func TestNBI_PostTask_RejectsEmptyArguments(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, failResponse)

	_, _, err := client.PostTask(context.Background(), "", map[string]any{"name": "reboot"}, false)
	require.ErrorIs(t, err, ErrEmptyArgument)

	_, _, err = client.PostTask(context.Background(), "dev1", nil, false)
	require.ErrorIs(t, err, ErrEmptyArgument)

	assert.Zero(t, recorder.count())
}

func TestNBI_WaitTask_Completes(t *testing.T) {
	recorder := &requestRecorder{}
	var polls atomic.Int32
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		if polls.Add(1) == 1 {
			writeBody(t, w, `[{"_id":"task-9"}]`)
			return
		}
		writeBody(t, w, `[]`)
	})

	gone := client.WaitTask(context.Background(), "task-9", 2*time.Millisecond)
	assert.True(t, gone)
	assert.GreaterOrEqual(t, polls.Load(), int32(2))
	assert.JSONEq(t, `{"_id":"task-9"}`, recorder.last().Query.Get("query"))
	assert.Equal(t, "/tasks/", recorder.last().Path)
}

func TestNBI_WaitTask_ContextExpiry(t *testing.T) {
	client := newTestNBI(t, &requestRecorder{}, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[{"_id":"task-9"}]`)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	assert.False(t, client.WaitTask(ctx, "task-9", 2*time.Millisecond))
}

const expectedPresetBody = `{
	"weight": 0,
	"precondition": "{\"_tags\":\"polyglot\"}",
	"configurations": [
		{"type": "value", "name": "InternetGatewayDevice.ManagementServer.PeriodicInformInterval", "value": "300"}
	]
}`

func TestNBI_PutPreset(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, ``)
	})

	preset := map[string]any{
		"weight":       0,
		"precondition": `{"_tags":"polyglot"}`,
		"configurations": []map[string]any{
			{
				"type":  "value",
				"name":  "InternetGatewayDevice.ManagementServer.PeriodicInformInterval",
				"value": "300",
			},
		},
	}
	require.NoError(t, client.PutPreset(context.Background(), "polyglot-inform", preset))

	request := recorder.last()
	assert.Equal(t, http.MethodPut, request.Method)
	assert.Equal(t, "/presets/polyglot-inform", request.Path)
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.JSONEq(t, expectedPresetBody, string(request.Body))
}

func TestNBI_PutPreset_StatusError(t *testing.T) {
	client := newTestNBI(t, &requestRecorder{}, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	err := client.PutPreset(context.Background(), "polyglot-inform", map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NBI status 400")
}

func TestNBI_ListFaults(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[{"_id":"dev1:default","message":"device rejected reboot"},{"_id":"dev1:task_1"}]`)
	})

	messages, err := client.ListFaults(context.Background(), "dev1")
	require.NoError(t, err)
	assert.Equal(t, []string{"device rejected reboot", "dev1:task_1"}, messages,
		"a fault without a message falls back to its fault ID")

	request := recorder.last()
	assert.Equal(t, http.MethodGet, request.Method)
	assert.Equal(t, "/faults/", request.Path)
	assert.JSONEq(t, `{"_id":{"$regex":"^dev1:"}}`, request.Query.Get("query"))
}

func TestNBI_ListFaults_EscapesDeviceIDRegex(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[]`)
	})

	_, err := client.ListFaults(context.Background(), "dev.1")
	require.NoError(t, err)
	assert.JSONEq(t, `{"_id":{"$regex":"^dev\\.1:"}}`, recorder.last().Query.Get("query"))
}

func TestNBI_ListFaults_StatusError(t *testing.T) {
	client := newTestNBI(t, &requestRecorder{}, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.ListFaults(context.Background(), "dev1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NBI status 500")
}

func TestDeviceBySerialQuery(t *testing.T) {
	query := DeviceBySerialQuery(" SN1 ")
	assert.JSONEq(t, `{
		"$or": [
			{"_deviceId._SerialNumber": "SN1"},
			{"InternetGatewayDevice.DeviceInfo.SerialNumber": "SN1"},
			{"Device.DeviceInfo.SerialNumber": "SN1"}
		]
	}`, query)

	assert.Equal(t, noMatchQuery, DeviceBySerialQuery("   "),
		"a blank serial must not turn into a full-device scan")
}

func TestDeviceByMACQuery(t *testing.T) {
	query := DeviceByMACQuery("AA:BB:CC:DD:EE:FF")

	var decoded struct {
		Or []map[string]any `json:"$or"`
	}
	require.NoError(t, json.Unmarshal([]byte(query), &decoded))
	require.Len(t, decoded.Or, len(macParameterPaths))

	first := decoded.Or[0]
	condition, ok := first[macParameterPaths[0]]
	require.True(t, ok, "first clause must use the first MAC parameter path")
	clause, ok := condition.(map[string]any)
	require.True(t, ok)
	values, ok := clause["$in"].([]any)
	require.True(t, ok)
	assert.ElementsMatch(t, []any{"AA:BB:CC:DD:EE:FF", "aa:bb:cc:dd:ee:ff"}, values,
		"both cases CPEs report must match")

	assert.Equal(t, noMatchQuery, DeviceByMACQuery(""),
		"a blank MAC must not turn into a full-device scan")
}

func TestParseDeviceRecord_FallsBackToIDAndParams(t *testing.T) {
	raw := json.RawMessage(`{
		"_id": "001122-IAD-SN-WITH-DASH",
		"_lastInform": "2023-05-06 07:08:09 +0000",
		"Device": {"DeviceInfo": {
			"SoftwareVersion": {"_value": "9.9"}
		}}
	}`)

	record, err := parseDeviceRecord(raw)
	require.NoError(t, err)
	assert.Equal(t, "001122", record.OUI)
	assert.Equal(t, "IAD", record.ProductClass)
	assert.Equal(t, "SN-WITH-DASH", record.SerialNumber)
	assert.Equal(t, "001122-IAD-SN-WITH-DASH", record.DeviceID)
	assert.Equal(t, "9.9", record.SoftwareVersion)
	assert.Equal(t, time.Date(2023, 5, 6, 7, 8, 9, 0, time.UTC), record.LastInform)
}

func TestParseDeviceRecord_RejectsMalformedJSON(t *testing.T) {
	_, err := parseDeviceRecord(json.RawMessage(`not-json`))
	require.Error(t, err)
}
