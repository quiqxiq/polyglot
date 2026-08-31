package provisioner

import (
	"context"
	"fmt"

	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// ProvisionPPPoE provisions a PPPoE subscriber secret and ensures its profile exists.
func (p *Provisioner) ProvisionPPPoE(ctx context.Context, deviceID string, spec domainSub.PPPoEProvisionSpec) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	if spec.Profile.Name != "" {
		if err := p.ensurePPPProfileFromSpec(ctx, driver, spec.Profile); err != nil {
			return fmt.Errorf("ensure ppp profile %s: %w", spec.Profile.Name, err)
		}
	}
	secParams := port.PPPoESecretParams{
		Name:          spec.Secret.Username,
		Password:      spec.Secret.Password,
		Profile:       spec.Secret.Profile,
		Service:       spec.Secret.Service,
		LocalAddress:  spec.Secret.LocalAddress,
		RemoteAddress: spec.Secret.RemoteAddress,
		Comment:       spec.Secret.Comment,
		Disabled:      spec.Secret.Disabled,
	}
	if secParams.Service == "" {
		secParams.Service = "pppoe"
	}
	if _, err := p.ppp.AddSecret(ctx, driver, secParams); err != nil {
		return fmt.Errorf("add ppp secret %s: %w", spec.Secret.Username, err)
	}
	return nil
}

func (p *Provisioner) suspendPPPoE(ctx context.Context, driver port.DeviceDriver, username string) error {
	sec, err := p.findSecret(ctx, driver, username)
	if err != nil {
		return err
	}
	if _, err := p.ppp.SetSecretDisabled(ctx, driver, sec.RosID, true); err != nil {
		return fmt.Errorf("disable ppp secret %s: %w", username, err)
	}
	p.kickPPP(ctx, driver, username)
	return nil
}

func (p *Provisioner) terminatePPPoE(ctx context.Context, driver port.DeviceDriver, username string) error {
	sec, err := p.findSecret(ctx, driver, username)
	if err != nil {
		return err
	}
	p.kickPPP(ctx, driver, username)
	if _, err := p.ppp.RemoveSecret(ctx, driver, sec.RosID); err != nil {
		return fmt.Errorf("remove ppp secret %s: %w", username, err)
	}
	return nil
}

func (p *Provisioner) ensurePPPProfileFromSpec(ctx context.Context, driver port.DeviceDriver, spec domainSub.PPPoEProfileSpec) error {
	params := pppProfileParamsFromSpec(spec)
	existing, err := p.ppp.ListProfiles(ctx, driver, params.Name)
	if err != nil {
		return fmt.Errorf("list ppp profiles: %w", err)
	}
	for _, pr := range existing {
		if pr.Name == params.Name {
			return nil
		}
	}
	if _, err := p.ppp.AddProfile(ctx, driver, params); err != nil {
		return fmt.Errorf("add ppp profile %s: %w", params.Name, err)
	}
	return nil
}

type secretView struct {
	RosID   string
	Name    string
	Profile string
}

// Params returns update parameters that only change the profile.
func (s secretView) Params() port.PPPoESecretParams {
	return port.PPPoESecretParams{Name: s.Name, Profile: s.Profile, Service: "pppoe"}
}

func (p *Provisioner) findSecret(ctx context.Context, driver port.DeviceDriver, username string) (secretView, error) {
	secrets, err := p.ppp.ListSecrets(ctx, driver, username)
	if err != nil {
		return secretView{}, fmt.Errorf("list ppp secrets: %w", err)
	}
	for _, s := range secrets {
		if s.Name == username {
			return secretView{RosID: s.RosID, Name: s.Name, Profile: s.Profile}, nil
		}
	}
	return secretView{}, fmt.Errorf("ppp secret %q not found", username)
}

func (p *Provisioner) kickPPP(ctx context.Context, driver port.DeviceDriver, username string) {
	sessions, err := p.ppp.ListActive(ctx, driver, username)
	if err != nil {
		return
	}
	for _, s := range sessions {
		_, _ = p.ppp.KickActive(ctx, driver, s.RosID)
	}
}
