package hotspot

import (
	"context"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

// ListIPBindings implements port.HotspotGateway.
func (g *Gateway) ListIPBindings(ctx context.Context, driver port.DeviceDriver) ([]port.HotspotIPBinding, error) {
	cmd := command.Command{Raw: "/ip/hotspot/ip-binding/print"}
	res, err := g.exec(ctx, driver, cmd)
	if err != nil {
		return nil, err
	}

	bindings := make([]port.HotspotIPBinding, 0, len(res.Rows))
	for _, row := range res.Rows {
		b := port.HotspotIPBinding{
			RosID:      row[".id"],
			MACAddress: row["mac-address"],
			Address:    row["address"],
			ToAddress:  row["to-address"],
			Server:     row["server"],
			Type:       row["type"],
			Comment:    row["comment"],
			Disabled:   row["disabled"] == "true" || row["disabled"] == "yes",
		}
		if b.Type == "" {
			b.Type = "regular"
		}
		bindings = append(bindings, b)
	}
	return bindings, nil
}

// CreateIPBinding implements port.HotspotGateway.
func (g *Gateway) CreateIPBinding(ctx context.Context, driver port.DeviceDriver, p port.HotspotIPBindingParams) (command.Result, error) {
	args := map[string]string{}
	if p.MACAddress != "" {
		args["mac-address"] = p.MACAddress
	}
	if p.Address != "" {
		args["address"] = p.Address
	}
	if p.ToAddress != "" {
		args["to-address"] = p.ToAddress
	}
	if p.Server != "" {
		args["server"] = p.Server
	}
	if p.Type != "" {
		args["type"] = p.Type
	} else {
		args["type"] = "bypassed"
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}

	cmd := command.Command{
		Raw:  "/ip/hotspot/ip-binding/add",
		Args: args,
	}
	return g.exec(ctx, driver, cmd)
}

// UpdateIPBinding implements port.HotspotGateway.
func (g *Gateway) UpdateIPBinding(ctx context.Context, driver port.DeviceDriver, rosID string, p port.HotspotIPBindingParams) (command.Result, error) {
	args := map[string]string{".id": rosID}
	if p.MACAddress != "" {
		args["mac-address"] = p.MACAddress
	}
	if p.Address != "" {
		args["address"] = p.Address
	}
	if p.ToAddress != "" {
		args["to-address"] = p.ToAddress
	}
	if p.Server != "" {
		args["server"] = p.Server
	}
	if p.Type != "" {
		args["type"] = p.Type
	}
	if p.Comment != "" {
		args["comment"] = p.Comment
	}
	if p.Disabled {
		args["disabled"] = "yes"
	} else {
		args["disabled"] = "no"
	}

	cmd := command.Command{
		Raw:  "/ip/hotspot/ip-binding/set",
		Args: args,
	}
	return g.exec(ctx, driver, cmd)
}

// DeleteIPBinding implements port.HotspotGateway.
func (g *Gateway) DeleteIPBinding(ctx context.Context, driver port.DeviceDriver, rosID string) (command.Result, error) {
	cmd := command.Command{
		Raw:  "/ip/hotspot/ip-binding/remove",
		Args: map[string]string{".id": rosID},
	}
	return g.exec(ctx, driver, cmd)
}
