package network

import (
	"context"
	"errors"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validatingMockDriver wraps mockDriver (the DeviceDriver stub used by the
// package's other tests) and adds the optional Validate capability, so tests
// can exercise ExecuteCommand's validation wiring without touching a device.
type validatingMockDriver struct {
	mockDriver
	validateFn func(ctx context.Context, cmd command.Command) error
}

func (m *validatingMockDriver) Validate(ctx context.Context, cmd command.Command) error {
	if m.validateFn != nil {
		return m.validateFn(ctx, cmd)
	}
	return nil
}

func TestExecuteCommandValidation(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid command short-circuits before Execute", func(t *testing.T) {
		executed := false
		driver := &validatingMockDriver{
			validateFn: func(ctx context.Context, cmd command.Command) error {
				return errors.New(`mikrotik: validate "/ip/address/add": unknown attribute(s): bogus`)
			},
		}
		driver.executeFn = func(ctx context.Context, cmd command.Command) (command.Result, error) {
			executed = true
			return command.Result{}, nil
		}

		_, err := ExecuteCommand(ctx, driver, command.Command{Raw: "/ip/address/add", Args: map[string]string{"bogus": "1"}})
		require.Error(t, err)
		assert.False(t, executed, "an invalid command must never reach Execute")
	})

	t.Run("valid command passes validation and executes", func(t *testing.T) {
		executed := false
		driver := &validatingMockDriver{
			validateFn: func(ctx context.Context, cmd command.Command) error { return nil },
		}
		driver.executeFn = func(ctx context.Context, cmd command.Command) (command.Result, error) {
			executed = true
			return command.Result{}, nil
		}

		_, err := ExecuteCommand(ctx, driver, command.Command{Raw: "/system/resource/print"})
		require.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("non-validating driver skips validation entirely", func(t *testing.T) {
		driver := &mockDriver{}
		_, err := ExecuteCommand(ctx, driver, command.Command{Raw: "/system/resource/print"})
		require.NoError(t, err)
	})
}
