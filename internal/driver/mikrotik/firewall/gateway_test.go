package firewall

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureIsolationFilter(t *testing.T) {
	ctx := context.Background()

	t.Run("adds portal accept and drop rules", func(t *testing.T) {
		var executedCommands []command.Command

		mockExec := func(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
			executedCommands = append(executedCommands, cmd)
			if cmd.Raw == "/ip/firewall/filter/print" {
				return command.Result{
					Rows: []map[string]string{},
				}, nil
			}
			return command.Result{}, nil
		}

		gw := NewGateway(mockExec)
		err := gw.EnsureIsolationFilter(ctx, nil, "ISOLIR_USERS", "192.168.233.195")
		require.NoError(t, err)

		var actions []string
		for _, cmd := range executedCommands {
			if cmd.Raw == "/ip/firewall/filter/add" {
				actions = append(actions, cmd.Args["action"])
			}
		}

		assert.Contains(t, actions, "accept")
		assert.Contains(t, actions, "drop")
	})
}
