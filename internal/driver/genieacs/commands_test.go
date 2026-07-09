package genieacs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/quixiq/polyglot/internal/domain/command"
)

func TestClassify(t *testing.T) {
	tests := []struct {
		name string
		cmd  command.Command
		want command.Class
	}{
		{"reboot is destructive", command.Command{Raw: "reboot"}, command.ClassDestructive},
		{"factoryReset is destructive", command.Command{Raw: "factoryReset"}, command.ClassDestructive},
		{"setParameterValues is destructive", command.Command{Raw: "setParameterValues"}, command.ClassDestructive},
		{"addObject is destructive", command.Command{Raw: "addObject"}, command.ClassDestructive},
		{"deleteObject is destructive", command.Command{Raw: "deleteObject"}, command.ClassDestructive},
		{"download is destructive", command.Command{Raw: "download"}, command.ClassDestructive},
		{"getParameterValues is read-only", command.Command{Raw: "getParameterValues"}, command.ClassReadOnly},
		{"refreshObject is read-only", command.Command{Raw: "refreshObject"}, command.ClassReadOnly},
		{"getDevice is read-only", command.Command{Raw: "getDevice"}, command.ClassReadOnly},
		{"unknown name defaults read-only", command.Command{Raw: "no_such_task"}, command.ClassReadOnly},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Classify(tt.cmd))
		})
	}
}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name    string
		op      command.Operation
		want    command.Command
		wantErr bool
	}{
		{"get_status maps to getDevice", command.OpGetStatus, command.Command{Raw: "getDevice"}, false},
		{"reboot maps to reboot task", command.OpReboot, command.Command{Raw: "reboot"}, false},
		{"unsupported op errors", command.Operation("does_not_exist"), command.Command{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Translate(tt.op)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsTaskCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  command.Command
		want bool
	}{
		{"getDevice is direct-GET, not a task", command.Command{Raw: "getDevice"}, false},
		{"reboot is a task", command.Command{Raw: "reboot"}, true},
		{"getParameterValues is a task", command.Command{Raw: "getParameterValues"}, true},
		{"setParameterValues is a task", command.Command{Raw: "setParameterValues"}, true},
		{"download is a task", command.Command{Raw: "download"}, true},
		{"unknown name is not a task", command.Command{Raw: "no_such_task"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isTaskCommand(tt.cmd))
		})
	}
}

func TestBuildTaskBody(t *testing.T) {
	t.Run("reboot has no extra fields", func(t *testing.T) {
		got, err := buildTaskBody(command.Command{Raw: "reboot"})
		require.NoError(t, err)
		var body map[string]any
		require.NoError(t, json.Unmarshal(got, &body))
		assert.Equal(t, "reboot", body["name"])
		_, hasParams := body["parameterNames"]
		assert.False(t, hasParams)
		_, hasObj := body["objectName"]
		assert.False(t, hasObj)
	})

	t.Run("getParameterValues with parameterNames", func(t *testing.T) {
		cmd := command.Command{
			Raw:  "getParameterValues",
			Args: map[string]string{"parameterNames": `["Device.DeviceInfo.SerialNumber","Device.DeviceInfo.ModelName"]`},
		}
		got, err := buildTaskBody(cmd)
		require.NoError(t, err)
		var body taskBody
		require.NoError(t, json.Unmarshal(got, &body))
		assert.Equal(t, "getParameterValues", body.Name)
		assert.Equal(t, []string{"Device.DeviceInfo.SerialNumber", "Device.DeviceInfo.ModelName"}, body.ParameterNames)
	})

	t.Run("setParameterValues with parameterValues tuples", func(t *testing.T) {
		cmd := command.Command{
			Raw:  "setParameterValues",
			Args: map[string]string{"parameterValues": `[["Device.X.Y","hello","xsd:string"],["Device.X.N",42,"xsd:int"]]`},
		}
		got, err := buildTaskBody(cmd)
		require.NoError(t, err)
		var body taskBody
		require.NoError(t, json.Unmarshal(got, &body))
		require.Len(t, body.ParameterValues, 2)
		assert.Equal(t, "Device.X.Y", body.ParameterValues[0][0])
		assert.Equal(t, "hello", body.ParameterValues[0][1])
		assert.Equal(t, "xsd:string", body.ParameterValues[0][2])
	})

	t.Run("refreshObject with objectName", func(t *testing.T) {
		cmd := command.Command{
			Raw:  "refreshObject",
			Args: map[string]string{"objectName": "Device.DeviceInfo"},
		}
		got, err := buildTaskBody(cmd)
		require.NoError(t, err)
		var body taskBody
		require.NoError(t, json.Unmarshal(got, &body))
		assert.Equal(t, "Device.DeviceInfo", body.ObjectName)
	})

	t.Run("refreshObject with empty objectName refreshes root", func(t *testing.T) {
		cmd := command.Command{Raw: "refreshObject"}
		got, err := buildTaskBody(cmd)
		require.NoError(t, err)
		var body taskBody
		require.NoError(t, json.Unmarshal(got, &body))
		assert.Equal(t, "", body.ObjectName)
	})

	t.Run("download with file", func(t *testing.T) {
		cmd := command.Command{
			Raw:  "download",
			Args: map[string]string{"file": "firmware-v2.bin"},
		}
		got, err := buildTaskBody(cmd)
		require.NoError(t, err)
		var body taskBody
		require.NoError(t, json.Unmarshal(got, &body))
		assert.Equal(t, "firmware-v2.bin", body.File)
	})

	t.Run("malformed parameterNames JSON errors", func(t *testing.T) {
		cmd := command.Command{
			Raw:  "getParameterValues",
			Args: map[string]string{"parameterNames": `not json`},
		}
		_, err := buildTaskBody(cmd)
		require.Error(t, err)
	})

	t.Run("malformed parameterValues JSON errors", func(t *testing.T) {
		cmd := command.Command{
			Raw:  "setParameterValues",
			Args: map[string]string{"parameterValues": `["not","a","tuple"]`},
		}
		_, err := buildTaskBody(cmd)
		require.Error(t, err)
	})
}

func TestProjectionFromParamNames(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{"empty input yields no projection", "", ""},
		{"single param", `["Device.X.Y"]`, "Device.X.Y"},
		{"multiple params comma-joined", `["Device.X.Y","Device.X.Z"]`, "Device.X.Y,Device.X.Z"},
		{"malformed JSON yields no projection (safe fallback)", "not json", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, projectionFromParamNames(tt.json))
		})
	}
}

func TestFlattenJSON(t *testing.T) {
	t.Run("flat scalars", func(t *testing.T) {
		m := map[string]any{"_id": "ABC123", "count": float64(42), "active": true}
		out := make(map[string]string)
		flattenJSON(m, "", out)
		assert.Equal(t, "ABC123", out["_id"])
		assert.Equal(t, "42", out["count"])
		assert.Equal(t, "true", out["active"])
	})

	t.Run("nested objects produce dotted keys", func(t *testing.T) {
		m := map[string]any{
			"Device": map[string]any{
				"DeviceInfo": map[string]any{
					"SerialNumber": map[string]any{"_value": "SN001"},
				},
			},
		}
		out := make(map[string]string)
		flattenJSON(m, "", out)
		assert.Equal(t, "SN001", out["Device.DeviceInfo.SerialNumber._value"])
	})

	t.Run("prefix accumulates across levels", func(t *testing.T) {
		m := map[string]any{"leaf": "x"}
		out := make(map[string]string)
		flattenJSON(m, "root", out)
		assert.Equal(t, "x", out["root.leaf"])
	})
}
