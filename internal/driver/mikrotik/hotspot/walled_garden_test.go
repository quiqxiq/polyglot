package hotspot

import (
	"context"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureWalledGarden(t *testing.T) {
	ctx := context.Background()

	t.Run("adds missing domains and ip entry", func(t *testing.T) {
		var executedCommands []command.Command

		mockExec := func(ctx context.Context, driver port.DeviceDriver, cmd command.Command) (command.Result, error) {
			executedCommands = append(executedCommands, cmd)
			if cmd.Raw == "/ip/hotspot/walled-garden/print" {
				return command.Result{
					Rows: []map[string]string{
						{"dst-host": "*midtrans.com"},
					},
				}, nil
			}
			if cmd.Raw == "/ip/hotspot/walled-garden/ip/print" {
				return command.Result{
					Rows: []map[string]string{},
				}, nil
			}
			return command.Result{}, nil
		}

		gw := NewGateway(mockExec)
		domains := []string{"midtrans.com", "tripay.co.id", "*.whatsapp.com"}
		err := gw.EnsureWalledGarden(ctx, nil, domains, "192.168.233.195", "5176")
		require.NoError(t, err)

		// midtrans.com should be skipped because it already exists.
		// tripay.co.id and *.whatsapp.com should be added.
		// ip entry should be added.
		var addedHosts []string
		ipAdded := false
		for _, cmd := range executedCommands {
			if cmd.Raw == "/ip/hotspot/walled-garden/add" {
				addedHosts = append(addedHosts, cmd.Args["dst-host"])
			}
			if cmd.Raw == "/ip/hotspot/walled-garden/ip/add" {
				ipAdded = true
				assert.Equal(t, "192.168.233.195", cmd.Args["dst-address"])
				assert.Equal(t, "5176", cmd.Args["dst-port"])
			}
		}

		assert.Contains(t, addedHosts, "*tripay.co.id")
		assert.Contains(t, addedHosts, "*.whatsapp.com")
		assert.NotContains(t, addedHosts, "*midtrans.com")
		assert.True(t, ipAdded)
	})
}
