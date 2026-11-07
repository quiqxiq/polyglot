package ppp

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	"github.com/quixiq/polyglot/pkg/response"
)

// ListSecrets retrieves all PPPoE secrets on the router.
func (h *PPPConnectHandler) ListSecrets(ctx context.Context, req *connect.Request[polyglotv1.ListPPPSecretsRequest]) (*connect.Response[polyglotv1.ListPPPSecretsResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	secrets, err := h.useCase.ListSecrets(ctx, driver, req.Msg.NameFilter)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.ListPPPSecretsResponse{
		Secrets: toProtoPPPSecrets(secrets),
	}), nil
}

// GetSecret fetches a single PPPoE secret by its RouterOS ID.
func (h *PPPConnectHandler) GetSecret(ctx context.Context, req *connect.Request[polyglotv1.GetPPPSecretRequest]) (*connect.Response[polyglotv1.GetPPPSecretResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	secret, err := h.useCase.GetSecret(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.GetPPPSecretResponse{
		Secret: toProtoPPPSecret(secret),
	}), nil
}

// CreateSecret adds a new PPPoE secret to the router.
func (h *PPPConnectHandler) CreateSecret(ctx context.Context, req *connect.Request[polyglotv1.CreatePPPSecretRequest]) (*connect.Response[polyglotv1.CreatePPPSecretResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
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
		return nil, response.Internal("secret created but could not be loaded")
	}

	return connect.NewResponse(&polyglotv1.CreatePPPSecretResponse{
		Secret:  toProtoPPPSecret(secrets[0]),
		Message: fmt.Sprintf("secret %q created successfully", req.Msg.Name),
	}), nil
}

// UpdateSecret modifies an existing PPPoE secret on the router.
func (h *PPPConnectHandler) UpdateSecret(ctx context.Context, req *connect.Request[polyglotv1.UpdatePPPSecretRequest]) (*connect.Response[polyglotv1.UpdatePPPSecretResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	params := FromProtoUpdateSecretRequest(req.Msg)
	if _, err := h.useCase.UpdateSecret(ctx, driver, req.Msg.RosId, params); err != nil {
		return nil, response.MapDomainError(err)
	}

	secret, err := h.useCase.GetSecret(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.UpdatePPPSecretResponse{
		Secret:  toProtoPPPSecret(secret),
		Message: fmt.Sprintf("secret %q updated successfully", secret.Name),
	}), nil
}

// DeleteSecret removes a PPPoE secret from the router.
func (h *PPPConnectHandler) DeleteSecret(ctx context.Context, req *connect.Request[polyglotv1.DeletePPPSecretRequest]) (*connect.Response[polyglotv1.DeletePPPSecretResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	res, err := h.useCase.RemoveSecret(ctx, driver, req.Msg.RosId)
	if err != nil {
		return nil, response.MapDomainError(err)
	}

	return connect.NewResponse(&polyglotv1.DeletePPPSecretResponse{
		Message: fmt.Sprintf("secret deleted: output=%s", res.Output),
	}), nil
}

// SetSecretDisabled enables or disables a PPPoE secret on the router.
func (h *PPPConnectHandler) SetSecretDisabled(ctx context.Context, req *connect.Request[polyglotv1.SetPPPSecretDisabledRequest]) (*connect.Response[polyglotv1.SetPPPSecretDisabledResponse], error) {
	driver, err := h.getDriver(ctx, req.Msg.DeviceId)
	if err != nil {
		return nil, err
	}

	if _, err := h.useCase.SetSecretDisabled(ctx, driver, req.Msg.RosId, req.Msg.Disabled); err != nil {
		return nil, response.MapDomainError(err)
	}

	status := "enabled"
	if req.Msg.Disabled {
		status = "disabled"
	}

	return connect.NewResponse(&polyglotv1.SetPPPSecretDisabledResponse{
		Message: fmt.Sprintf("secret %s successfully", status),
	}), nil
}
