package ppp

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListSecrets retrieves all PPPoE secrets on the router.
func (h *PPPConnectHandler) ListSecrets(ctx context.Context, req *connect.Request[devicepb.ListPPPSecretsRequest]) (*connect.Response[devicepb.ListPPPSecretsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	secrets, err := h.useCase.ListSecrets(ctx, driver, req.Msg.NameFilter)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.ListPPPSecretsResponse{
		Secrets: ToProtoPPPSecrets(secrets),
	}), nil
}

// GetSecret fetches a single PPPoE secret by its RouterOS ID.
func (h *PPPConnectHandler) GetSecret(ctx context.Context, req *connect.Request[devicepb.GetPPPSecretRequest]) (*connect.Response[devicepb.GetPPPSecretResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id is required"))
	}

	secret, err := h.useCase.GetSecret(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.GetPPPSecretResponse{
		Secret: ToProtoPPPSecret(secret),
	}), nil
}

// CreateSecret adds a new PPPoE secret to the router.
func (h *PPPConnectHandler) CreateSecret(ctx context.Context, req *connect.Request[devicepb.CreatePPPSecretRequest]) (*connect.Response[devicepb.CreatePPPSecretResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}

	params := FromProtoCreateSecretRequest(req.Msg)
	if _, err := h.useCase.AddSecret(ctx, driver, params); err != nil {
		return nil, response.MapDomainError(err)
	}

	// Re-print to return the created secret (with RouterOS .id)
	secrets, err := h.useCase.ListSecrets(ctx, driver, req.Msg.Name)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	if len(secrets) == 0 {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("secret %q created but not found", req.Msg.Name))
	}

	return connect.NewResponse(&devicepb.CreatePPPSecretResponse{
		Secret:  ToProtoPPPSecret(secrets[0]),
		Message: fmt.Sprintf("secret %q created successfully", req.Msg.Name),
	}), nil
}

// UpdateSecret modifies an existing PPPoE secret on the router.
func (h *PPPConnectHandler) UpdateSecret(ctx context.Context, req *connect.Request[devicepb.UpdatePPPSecretRequest]) (*connect.Response[devicepb.UpdatePPPSecretResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id is required"))
	}

	params := FromProtoUpdateSecretRequest(req.Msg)
	if _, err := h.useCase.UpdateSecret(ctx, driver, req.Msg.RosId, params); err != nil {
		return nil, response.MapDomainError(err)
	}

	secret, err := h.useCase.GetSecret(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.UpdatePPPSecretResponse{
		Secret:  ToProtoPPPSecret(secret),
		Message: fmt.Sprintf("secret %q updated successfully", secret.Name),
	}), nil
}

// DeleteSecret removes a PPPoE secret from the router.
func (h *PPPConnectHandler) DeleteSecret(ctx context.Context, req *connect.Request[devicepb.DeletePPPSecretRequest]) (*connect.Response[devicepb.DeletePPPSecretResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id is required"))
	}

	res, err := h.useCase.RemoveSecret(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&devicepb.DeletePPPSecretResponse{
		Message: fmt.Sprintf("secret deleted: output=%s", res.Output),
	}), nil
}

// SetSecretDisabled enables or disables a PPPoE secret on the router.
func (h *PPPConnectHandler) SetSecretDisabled(ctx context.Context, req *connect.Request[devicepb.SetPPPSecretDisabledRequest]) (*connect.Response[devicepb.SetPPPSecretDisabledResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}
	if req.Msg.RosId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("ros_id is required"))
	}

	if _, err := h.useCase.SetSecretDisabled(ctx, driver, req.Msg.RosId, req.Msg.Disabled); err != nil {
		return nil, response.MapDomainError(err)
	}

	status := "enabled"
	if req.Msg.Disabled {
		status = "disabled"
	}

	return connect.NewResponse(&devicepb.SetPPPSecretDisabledResponse{
		Message: fmt.Sprintf("secret %s successfully", status),
	}), nil
}
