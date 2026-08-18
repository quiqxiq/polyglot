package hotspot

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListHotspotIPBindings returns all /ip/hotspot/ip-binding entries.
func (h *HotspotConnectHandler) ListHotspotIPBindings(ctx context.Context, req *connect.Request[devicepb.ListHotspotIPBindingsRequest]) (*connect.Response[devicepb.ListHotspotIPBindingsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	bindings, err := h.useCase.GetIPBindings(ctx, driver)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	protoBindings := make([]*devicepb.HotspotIPBinding, 0, len(bindings))
	for _, b := range bindings {
		protoBindings = append(protoBindings, &devicepb.HotspotIPBinding{
			Id:         b.RosID,
			MacAddress: b.MACAddress,
			Address:    b.Address,
			ToAddress:  b.ToAddress,
			Server:     b.Server,
			Type:       b.Type,
			Comment:    b.Comment,
			Disabled:   b.Disabled,
		})
	}

	return connect.NewResponse(&devicepb.ListHotspotIPBindingsResponse{
		Bindings: protoBindings,
	}), nil
}

// CreateHotspotIPBinding adds an IP Binding entry.
func (h *HotspotConnectHandler) CreateHotspotIPBinding(ctx context.Context, req *connect.Request[devicepb.CreateHotspotIPBindingRequest]) (*connect.Response[devicepb.CreateHotspotIPBindingResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	p := port.HotspotIPBindingParams{
		MACAddress: req.Msg.MacAddress,
		Address:    req.Msg.Address,
		ToAddress:  req.Msg.ToAddress,
		Server:     req.Msg.Server,
		Type:       req.Msg.Type,
		Comment:    req.Msg.Comment,
		Disabled:   req.Msg.Disabled,
	}

	res, err := h.useCase.CreateIPBinding(ctx, driver, p)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.CreateHotspotIPBindingResponse{
		Binding: &devicepb.HotspotIPBinding{
			Id:         res.Output,
			MacAddress: p.MACAddress,
			Address:    p.Address,
			ToAddress:  p.ToAddress,
			Server:     p.Server,
			Type:       p.Type,
			Comment:    p.Comment,
			Disabled:   p.Disabled,
		},
		Message: "IP binding created successfully",
	}), nil
}

// UpdateHotspotIPBinding updates an existing IP Binding entry.
func (h *HotspotConnectHandler) UpdateHotspotIPBinding(ctx context.Context, req *connect.Request[devicepb.UpdateHotspotIPBindingRequest]) (*connect.Response[devicepb.UpdateHotspotIPBindingResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id is required"))
	}

	p := port.HotspotIPBindingParams{
		MACAddress: req.Msg.MacAddress,
		Address:    req.Msg.Address,
		ToAddress:  req.Msg.ToAddress,
		Server:     req.Msg.Server,
		Type:       req.Msg.Type,
		Comment:    req.Msg.Comment,
		Disabled:   req.Msg.Disabled,
	}

	if _, err := h.useCase.UpdateIPBinding(ctx, driver, req.Msg.RosId, p); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdateHotspotIPBindingResponse{
		Binding: &devicepb.HotspotIPBinding{
			Id:         req.Msg.RosId,
			MacAddress: p.MACAddress,
			Address:    p.Address,
			ToAddress:  p.ToAddress,
			Server:     p.Server,
			Type:       p.Type,
			Comment:    p.Comment,
			Disabled:   p.Disabled,
		},
		Message: "IP binding updated successfully",
	}), nil
}

// DeleteHotspotIPBinding removes an IP Binding entry.
func (h *HotspotConnectHandler) DeleteHotspotIPBinding(ctx context.Context, req *connect.Request[devicepb.DeleteHotspotIPBindingRequest]) (*connect.Response[devicepb.DeleteHotspotIPBindingResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id is required"))
	}

	if _, err := h.useCase.DeleteIPBinding(ctx, driver, req.Msg.RosId); err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeleteHotspotIPBindingResponse{
		Message: "IP binding deleted successfully",
	}), nil
}
