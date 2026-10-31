package genieacs

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNBI_AcknowledgedMutations(t *testing.T) {
	tests := []struct {
		name       string
		call       func(ctx context.Context, client *NBI) error
		wantMethod string
		wantPath   string
	}{
		{
			name:       "delete preset",
			call:       func(ctx context.Context, client *NBI) error { return client.DeletePreset(ctx, "polyglot-inform") },
			wantMethod: http.MethodDelete,
			wantPath:   "/presets/polyglot-inform",
		},
		{
			name:       "delete fault",
			call:       func(ctx context.Context, client *NBI) error { return client.DeleteFault(ctx, "dev1:default") },
			wantMethod: http.MethodDelete,
			wantPath:   "/faults/dev1:default",
		},
		{
			name:       "delete device",
			call:       func(ctx context.Context, client *NBI) error { return client.DeleteDevice(ctx, "dev1") },
			wantMethod: http.MethodDelete,
			wantPath:   "/devices/dev1",
		},
		{
			name:       "add tag",
			call:       func(ctx context.Context, client *NBI) error { return client.AddTag(ctx, "dev1", "polyglot") },
			wantMethod: http.MethodPost,
			wantPath:   "/devices/dev1/tags/polyglot",
		},
		{
			name:       "remove tag",
			call:       func(ctx context.Context, client *NBI) error { return client.RemoveTag(ctx, "dev1", "polyglot") },
			wantMethod: http.MethodDelete,
			wantPath:   "/devices/dev1/tags/polyglot",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &requestRecorder{}
			client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
				writeBody(t, w, ``)
			})

			require.NoError(t, tc.call(context.Background(), client))

			request := recorder.last()
			assert.Equal(t, tc.wantMethod, request.Method)
			assert.Equal(t, tc.wantPath, request.Path)
			assert.Equal(t, "test-api-key", request.Header.Get("x-api-key"))
		})
	}
}

func TestNBI_DeletePreset_StatusError(t *testing.T) {
	client := newTestNBI(t, &requestRecorder{}, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeBody(t, w, "preset not found")
	})

	err := client.DeletePreset(context.Background(), "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NBI status 404")
	assert.Contains(t, err.Error(), "preset not found")
}

func TestNBI_AddTag_EscapesPathSegments(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, ``)
	})

	require.NoError(t, client.AddTag(context.Background(), "SN%2D1", "polyglot"))
	assert.Equal(t, "/devices/SN%2D1/tags/polyglot", recorder.last().Path)
}

func TestNBI_ListPresets(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, func(w http.ResponseWriter, _ *http.Request) {
		writeBody(t, w, `[
			{"_id":"polyglot-inform","precondition":"{\"_tags\":\"polyglot\"}","configurations":[]},
			{"_id":"polyglot-bootstrap"}
		]`)
	})

	presets, err := client.ListPresets(context.Background())
	require.NoError(t, err)
	require.Len(t, presets, 2)

	var inform map[string]any
	require.NoError(t, json.Unmarshal(presets["polyglot-inform"], &inform))
	assert.Equal(t, `{"_tags":"polyglot"}`, inform["precondition"])
	assert.JSONEq(t, `{"_id":"polyglot-bootstrap"}`, string(presets["polyglot-bootstrap"]))

	request := recorder.last()
	assert.Equal(t, http.MethodGet, request.Method)
	assert.Equal(t, "/presets/", request.Path)
}

func TestNBI_ListPresets_StatusError(t *testing.T) {
	client := newTestNBI(t, &requestRecorder{}, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.ListPresets(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NBI status 500")
}

func TestNBI_Admin_RejectsEmptyArguments(t *testing.T) {
	recorder := &requestRecorder{}
	client := newTestNBI(t, recorder, failResponse)
	ctx := context.Background()

	calls := []struct {
		name string
		call func() error
	}{
		{"put preset name", func() error {
			return client.PutPreset(ctx, " ", map[string]any{})
		}},
		{"put preset body", func() error {
			return client.PutPreset(ctx, "polyglot-inform", nil)
		}},
		{"delete preset", func() error {
			return client.DeletePreset(ctx, "")
		}},
		{"delete fault", func() error {
			return client.DeleteFault(ctx, "")
		}},
		{"delete device", func() error {
			return client.DeleteDevice(ctx, " ")
		}},
		{"list faults device", func() error {
			_, err := client.ListFaults(ctx, "")
			return err
		}},
		{"add tag device", func() error {
			return client.AddTag(ctx, "", "polyglot")
		}},
		{"add tag tag", func() error {
			return client.AddTag(ctx, "dev1", " ")
		}},
		{"remove tag device", func() error {
			return client.RemoveTag(ctx, "", "polyglot")
		}},
		{"remove tag tag", func() error {
			return client.RemoveTag(ctx, "dev1", "")
		}},
	}

	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			require.ErrorIs(t, tc.call(), ErrEmptyArgument)
		})
	}
	assert.Zero(t, recorder.count(), "no request may be sent for an empty identifier")
}
