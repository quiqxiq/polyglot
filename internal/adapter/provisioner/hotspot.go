package provisioner

import (
	"context"
	"fmt"

	domainSub "github.com/quixiq/polyglot/internal/domain/subscription"
	"github.com/quixiq/polyglot/internal/port"
)

// ProvisionHotspot provisions a Hotspot subscriber user and ensures its profile exists.
func (p *Provisioner) ProvisionHotspot(ctx context.Context, deviceID string, spec domainSub.HotspotProvisionSpec) error {
	driver, err := p.resolve(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("resolve driver %s: %w", deviceID, err)
	}
	if spec.Profile.Name != "" {
		if err := p.ensureHotspotProfileFromSpec(ctx, driver, spec.Profile); err != nil {
			return fmt.Errorf("ensure hotspot profile %s: %w", spec.Profile.Name, err)
		}
	}
	userParams := port.HotspotUserParams{
		Name:          spec.User.Username,
		Password:      spec.User.Password,
		Profile:       spec.User.Profile,
		Server:        spec.User.Server,
		MACAddress:    spec.User.MacAddress,
		Address:       spec.User.IPAddress,
		LimitUptime:   spec.User.LimitUptime,
		LimitBytesOut: spec.User.LimitBytes,
		Comment:       spec.User.Comment,
		Disabled:      spec.User.Disabled,
	}
	if userParams.Server == "" {
		userParams.Server = "all"
	}
	if _, err := p.hot.AddUser(ctx, driver, userParams); err != nil {
		return fmt.Errorf("add hotspot user %s: %w", spec.User.Username, err)
	}
	return nil
}

func (p *Provisioner) suspendHotspot(ctx context.Context, driver port.DeviceDriver, username string) error {
	u, rosID, err := p.findHotspotUser(ctx, driver, username)
	if err != nil {
		return err
	}
	u.Disabled = true
	if _, err := p.hot.UpdateUser(ctx, driver, rosID, u); err != nil {
		return fmt.Errorf("suspend hotspot user %s: %w", username, err)
	}
	p.kickHotspotIfActive(ctx, driver, username)
	return nil
}

func (p *Provisioner) terminateHotspot(ctx context.Context, driver port.DeviceDriver, username string) error {
	_, rosID, err := p.findHotspotUser(ctx, driver, username)
	if err != nil {
		return err
	}
	if _, err := p.hot.RemoveUser(ctx, driver, rosID); err != nil {
		return fmt.Errorf("remove hotspot user %s: %w", username, err)
	}
	return nil
}

func (p *Provisioner) ensureHotspotProfileFromSpec(ctx context.Context, driver port.DeviceDriver, spec domainSub.HotspotProfileSpec) error {
	params := hotspotProfileParamsFromSpec(spec)
	existing, err := p.hot.GetUserProfiles(ctx, driver)
	if err != nil {
		return fmt.Errorf("list hotspot profiles: %w", err)
	}
	for _, pr := range existing {
		if pr.Name == params.Name {
			return nil
		}
	}
	if _, err := p.hot.CreateUserProfile(ctx, driver, params); err != nil {
		return fmt.Errorf("add hotspot profile %s: %w", params.Name, err)
	}
	return nil
}

func (p *Provisioner) findHotspotUser(ctx context.Context, driver port.DeviceDriver, username string) (port.HotspotUserParams, string, error) {
	users, err := p.hot.ListUsers(ctx, driver, port.ListUsersFilter{Name: username})
	if err != nil {
		return port.HotspotUserParams{}, "", fmt.Errorf("list hotspot users: %w", err)
	}
	for _, u := range users {
		if u.Name == username {
			return port.HotspotUserParams{
				Name: u.Name, Password: u.Password, Profile: u.Profile, Server: u.Server,
				MACAddress: u.MACAddress, Address: u.Address,
				LimitUptime: u.LimitUptime, LimitBytesIn: u.LimitBytesIn, LimitBytesOut: u.LimitBytesOut,
				Comment: u.Comment, Disabled: u.Disabled,
			}, u.RosID, nil
		}
	}
	return port.HotspotUserParams{}, "", fmt.Errorf("hotspot user %q not found", username)
}

func (p *Provisioner) kickHotspotIfActive(ctx context.Context, driver port.DeviceDriver, username string) {
	sessions, err := p.hot.ListActiveSessions(ctx, driver)
	if err == nil {
		for _, s := range sessions {
			if s.User == username {
				_, _ = p.hot.RemoveActiveSession(ctx, driver, s.RosID)
			}
		}
	}
	cookies, err := p.hot.ListCookies(ctx, driver)
	if err == nil {
		for _, c := range cookies {
			if c.User == username {
				_, _ = p.hot.DeleteCookie(ctx, driver, c.RosID)
			}
		}
	}
}
