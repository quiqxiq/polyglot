package hotspot

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// ListCookies implements port.HotspotGateway.
func (g *Gateway) ListCookies(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotCookie, error) {
	cmd := command.Command{Raw: "/ip/hotspot/cookie/print"}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}

	cookies := make([]port.HotspotCookie, 0, len(res.Rows))
	for _, row := range res.Rows {
		c := port.HotspotCookie{
			RosID:      row[".id"],
			User:       row["user"],
			MACAddress: row["mac-address"],
			ExpiresIn:  row["expires-in"],
			Domain:     row["domain"],
		}
		cookies = append(cookies, c)
	}
	return cookies, nil
}

// DeleteCookie implements port.HotspotGateway.
func (g *Gateway) DeleteCookie(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	if rosID == "" || rosID == "all" {
		cookies, err := g.ListCookies(ctx, driver)
		if err != nil {
			return command.Result{}, err
		}
		for _, c := range cookies {
			if c.RosID != "" {
				_, _ = g.exec(ctx, driver, command.Command{
					Raw:  "/ip/hotspot/cookie/remove",
					Args: map[string]string{".id": c.RosID},
				})
			}
		}
		return command.Result{Output: "success"}, nil
	}

	cmd := command.Command{
		Raw:  "/ip/hotspot/cookie/remove",
		Args: map[string]string{".id": rosID},
	}
	return g.exec(ctx, driver, cmd)
}
