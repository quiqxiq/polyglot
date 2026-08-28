package ppp

import (
	"context"
	"fmt"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// Gateway implements port.PPPGateway for MikroTik RouterOS.
type Gateway struct {
	exec port.CommandExecutor
}

// NewGateway creates a Gateway bound to exec.
func NewGateway(exec port.CommandExecutor) *Gateway {
	return &Gateway{exec: exec}
}

var _ port.PPPGateway = (*Gateway)(nil)

// ListSecrets implements port.PPPGateway.
func (g *Gateway) ListSecrets(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPoESecret, error) {
	res, err := g.exec(ctx, driver, NewPrintSecretsCommand(nameFilter))
	if err != nil {
		return nil, err
	}
	return ParseSecrets(res), nil
}

// GetSecret implements port.PPPGateway.
func (g *Gateway) GetSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPoESecret, error) {
	cmd := command.Command{
		Raw:  "/ppp/secret/print",
		Args: map[string]string{"?.id": rosID},
	}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return port.PPPoESecret{}, err
	}
	secrets := ParseSecrets(res)
	if len(secrets) == 0 {
		return port.PPPoESecret{}, fmt.Errorf("ppp secret %q not found", rosID)
	}
	return secrets[0], nil
}

// AddSecret implements port.PPPGateway.
func (g *Gateway) AddSecret(ctx context.Context, driver port.DeviceDriver, p port.PPPoESecretParams) (command.Result, error) {
	return g.exec(ctx, driver, NewAddSecretCommand(p))
}

// UpdateSecret implements port.PPPGateway.
func (g *Gateway) UpdateSecret(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPoESecretParams) (command.Result, error) {
	return g.exec(ctx, driver, NewSetSecretCommand(rosID, p))
}

// RemoveSecret implements port.PPPGateway.
func (g *Gateway) RemoveSecret(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewRemoveSecretCommand(rosID))
}

// SetSecretDisabled implements port.PPPGateway.
func (g *Gateway) SetSecretDisabled(ctx context.Context, driver port.DeviceDriver, rosID string, disabled bool) (command.Result, error) {
	return g.exec(ctx, driver, NewSetSecretDisabledCommand(rosID, disabled))
}

// ListProfiles implements port.PPPGateway.
func (g *Gateway) ListProfiles(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPProfile, error) {
	res, err := g.exec(ctx, driver, NewPrintProfilesCommand(nameFilter))
	if err != nil {
		return nil, err
	}
	return ParseProfiles(res), nil
}

// GetProfile implements port.PPPGateway.
func (g *Gateway) GetProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (port.PPPProfile, error) {
	cmd := command.Command{
		Raw:  "/ppp/profile/print",
		Args: map[string]string{"?.id": rosID},
	}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return port.PPPProfile{}, err
	}
	profiles := ParseProfiles(res)
	if len(profiles) == 0 {
		return port.PPPProfile{}, fmt.Errorf("ppp profile %q not found", rosID)
	}
	return profiles[0], nil
}

// AddProfile implements port.PPPGateway.
func (g *Gateway) AddProfile(ctx context.Context, driver port.DeviceDriver, p port.PPPProfileParams) (command.Result, error) {
	return g.exec(ctx, driver, NewAddProfileCommand(p))
}

// UpdateProfile implements port.PPPGateway.
func (g *Gateway) UpdateProfile(ctx context.Context, driver port.DeviceDriver, rosID string, p port.PPPProfileParams) (command.Result, error) {
	return g.exec(ctx, driver, NewSetProfileCommand(rosID, p))
}

// RemoveProfile implements port.PPPGateway.
func (g *Gateway) RemoveProfile(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewRemoveProfileCommand(rosID))
}

// ListActive implements port.PPPGateway.
func (g *Gateway) ListActive(ctx context.Context, driver port.DeviceDriver, nameFilter string) ([]port.PPPActiveSession, error) {
	res, err := g.exec(ctx, driver, NewPrintActiveCommand(nameFilter))
	if err != nil {
		return nil, err
	}
	active := ParseActiveSessions(res)

	// Enrich with profile names if possible
	if secRes, err := g.exec(ctx, driver, NewPrintSecretsCommand(nameFilter)); err == nil {
		secrets := ParseSecrets(secRes)
		active = EnrichActiveSessionsWithProfiles(active, secrets)
	}
	return active, nil
}

// KickActive implements port.PPPGateway.
func (g *Gateway) KickActive(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	return g.exec(ctx, driver, NewKickActiveCommand(rosID))
}

// ListInactive implements port.PPPGateway.
func (g *Gateway) ListInactive(ctx context.Context, driver port.DeviceDriver) ([]port.PPPoESecret, error) {
	secRes, err := g.exec(ctx, driver, NewPrintSecretsCommand(""))
	if err != nil {
		return nil, err
	}
	actRes, err := g.exec(ctx, driver, NewPrintActiveCommand(""))
	if err != nil {
		return nil, err
	}
	secrets := ParseSecrets(secRes)
	active := ParseActiveSessions(actRes)
	return FilterInactiveSecrets(secrets, active), nil
}

