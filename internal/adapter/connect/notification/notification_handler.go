package notification

import (
	"context"

	"connectrpc.com/connect"

	polyglotv1 "github.com/quixiq/polyglot/api/gen/v1"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	notificationUC "github.com/quixiq/polyglot/internal/usecase/notification"
	"github.com/quixiq/polyglot/pkg/response"
)

// NotificationConnectHandler implements the notification ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type NotificationConnectHandler struct {
	useCase *notificationUC.ManageNotificationUseCase
}

// NewNotificationConnectHandler constructs a notification ConnectRPC handler.
func NewNotificationConnectHandler(uc *notificationUC.ManageNotificationUseCase) *NotificationConnectHandler {
	return &NotificationConnectHandler{useCase: uc}
}

// ListTemplates returns notification templates.
func (h *NotificationConnectHandler) ListTemplates(ctx context.Context, req *connect.Request[polyglotv1.ListNotificationTemplatesRequest]) (*connect.Response[polyglotv1.ListNotificationTemplatesResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("notification usecase unavailable")
	}
	list, err := h.useCase.ListTemplates(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ListNotificationTemplatesResponse{
		Templates: toProtoTemplateList(list),
	}), nil
}

// GetTemplate returns one notification template.
func (h *NotificationConnectHandler) GetTemplate(ctx context.Context, req *connect.Request[polyglotv1.GetNotificationTemplateRequest]) (*connect.Response[polyglotv1.GetNotificationTemplateResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("notification usecase unavailable")
	}
	t, err := h.useCase.GetTemplate(ctx, req.Msg.TemplateKey)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.GetNotificationTemplateResponse{
		Template: toProtoTemplate(&t),
	}), nil
}

// SaveTemplate creates or updates a notification template.
func (h *NotificationConnectHandler) SaveTemplate(ctx context.Context, req *connect.Request[polyglotv1.SaveNotificationTemplateRequest]) (*connect.Response[polyglotv1.SaveNotificationTemplateResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("notification usecase unavailable")
	}
	pb := req.Msg.Template
	t := domainNotification.NotificationTemplate{
		ID:            pb.Id,
		TemplateKey:   pb.TemplateKey,
		Name:          pb.Name,
		Content:       pb.Content,
		VariablesJSON: pb.VariablesJson,
		IsActive:      pb.IsActive,
	}
	saved, err := h.useCase.SaveTemplate(ctx, t)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.SaveNotificationTemplateResponse{
		Template: toProtoTemplate(&saved),
	}), nil
}

// ListNotifications returns notification records.
func (h *NotificationConnectHandler) ListNotifications(ctx context.Context, req *connect.Request[polyglotv1.ListNotificationsRequest]) (*connect.Response[polyglotv1.ListNotificationsResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("notification usecase unavailable")
	}
	list, err := h.useCase.ListNotifications(ctx, req.Msg.CustomerId, int(req.Msg.Limit))
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.ListNotificationsResponse{
		Notifications: toProtoWANotificationList(list),
	}), nil
}

// PendingCount returns the number of pending notifications.
func (h *NotificationConnectHandler) PendingCount(ctx context.Context, _ *connect.Request[polyglotv1.PendingCountRequest]) (*connect.Response[polyglotv1.PendingCountResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("notification usecase unavailable")
	}
	count, err := h.useCase.PendingCount(ctx)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.PendingCountResponse{
		Queued: int32(count),
	}), nil
}

// MarkNotificationSent marks a notification as sent.
func (h *NotificationConnectHandler) MarkNotificationSent(ctx context.Context, req *connect.Request[polyglotv1.MarkNotificationSentRequest]) (*connect.Response[polyglotv1.MarkNotificationSentResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("notification usecase unavailable")
	}
	if err := h.useCase.MarkSent(ctx, req.Msg.Id); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.MarkNotificationSentResponse{}), nil
}

// MarkNotificationFailed marks a notification as failed.
func (h *NotificationConnectHandler) MarkNotificationFailed(ctx context.Context, req *connect.Request[polyglotv1.MarkNotificationFailedRequest]) (*connect.Response[polyglotv1.MarkNotificationFailedResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("notification usecase unavailable")
	}
	if err := h.useCase.MarkFailed(ctx, req.Msg.Id, req.Msg.ErrorMessage); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.MarkNotificationFailedResponse{}), nil
}

// TestSend sends a test notification.
func (h *NotificationConnectHandler) TestSend(ctx context.Context, req *connect.Request[polyglotv1.TestSendRequest]) (*connect.Response[polyglotv1.TestSendResponse], error) {
	if h.useCase == nil {
		return nil, response.Unavailable("notification usecase unavailable")
	}
	if err := h.useCase.TestSend(ctx, req.Msg.Phone, req.Msg.Content); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&polyglotv1.TestSendResponse{Sent: true}), nil
}
