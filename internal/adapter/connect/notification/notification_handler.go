package notification

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
	"github.com/quixiq/polyglot/pkg/response"
)

// NotificationConnectHandler implements the notification ConnectRPC service.
// NotificationConnectHandler implements the notification ConnectRPC service.
//
//nolint:revive // Explicit transport role is part of the project naming convention.
type NotificationConnectHandler struct {
	repo   port.NotificationRepository
	sender port.NotificationSender
}

// NewNotificationConnectHandler constructs a notification ConnectRPC handler.
func NewNotificationConnectHandler(repo port.NotificationRepository, sender port.NotificationSender) *NotificationConnectHandler {
	return &NotificationConnectHandler{repo: repo, sender: sender}
}

// ListTemplates returns notification templates.
func (h *NotificationConnectHandler) ListTemplates(ctx context.Context, req *connect.Request[devicepb.ListNotificationTemplatesRequest]) (*connect.Response[devicepb.ListNotificationTemplatesResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("notification repository unavailable"))
	}
	list, err := h.repo.ListTemplates(ctx, req.Msg.ActiveOnly)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListNotificationTemplatesResponse{
		Templates: toProtoTemplateList(list),
	}), nil
}

// GetTemplate returns one notification template.
func (h *NotificationConnectHandler) GetTemplate(ctx context.Context, req *connect.Request[devicepb.GetNotificationTemplateRequest]) (*connect.Response[devicepb.GetNotificationTemplateResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("notification repository unavailable"))
	}
	t, err := h.repo.FindTemplateByKey(ctx, "tenant-default", req.Msg.TemplateKey)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.GetNotificationTemplateResponse{
		Template: toProtoTemplate(&t),
	}), nil
}

// SaveTemplate creates or updates a notification template.
func (h *NotificationConnectHandler) SaveTemplate(ctx context.Context, req *connect.Request[devicepb.SaveNotificationTemplateRequest]) (*connect.Response[devicepb.SaveNotificationTemplateResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("notification repository unavailable"))
	}
	pb := req.Msg.Template
	id := pb.Id
	if id == "" {
		id = idgen.New("nt")
	}
	t := domainNotification.NotificationTemplate{
		ID: id, TenantID: "tenant-default", TemplateKey: pb.TemplateKey,
		Name: pb.Name, Content: pb.Content, VariablesJSON: pb.VariablesJson,
		IsActive: pb.IsActive,
	}
	if err := h.repo.SaveTemplate(ctx, t); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.SaveNotificationTemplateResponse{
		Template: toProtoTemplate(&t),
	}), nil
}

// ListNotifications returns notification records.
func (h *NotificationConnectHandler) ListNotifications(ctx context.Context, req *connect.Request[devicepb.ListNotificationsRequest]) (*connect.Response[devicepb.ListNotificationsResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("notification repository unavailable"))
	}
	var list []domainNotification.WANotification
	var err error
	if req.Msg.CustomerId != "" {
		list, err = h.repo.ListByCustomer(ctx, req.Msg.CustomerId, int(req.Msg.Limit))
	} else {
		list, err = h.repo.Pending(ctx, int(req.Msg.Limit))
	}
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.ListNotificationsResponse{
		Notifications: toProtoWANotificationList(list),
	}), nil
}

// PendingCount returns the number of pending notifications.
func (h *NotificationConnectHandler) PendingCount(ctx context.Context, _ *connect.Request[devicepb.PendingCountRequest]) (*connect.Response[devicepb.PendingCountResponse], error) {
	if h.repo == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("notification repository unavailable"))
	}
	pending, err := h.repo.Pending(ctx, 0)
	if err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.PendingCountResponse{
		Queued: int32(len(pending)),
	}), nil
}

// MarkNotificationSent marks a notification as sent.
func (h *NotificationConnectHandler) MarkNotificationSent(ctx context.Context, req *connect.Request[devicepb.MarkNotificationSentRequest]) (*connect.Response[devicepb.MarkNotificationSentResponse], error) {
	if err := h.repo.MarkSent(ctx, req.Msg.Id, time.Now()); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.MarkNotificationSentResponse{}), nil
}

// MarkNotificationFailed marks a notification as failed.
func (h *NotificationConnectHandler) MarkNotificationFailed(ctx context.Context, req *connect.Request[devicepb.MarkNotificationFailedRequest]) (*connect.Response[devicepb.MarkNotificationFailedResponse], error) {
	if err := h.repo.MarkFailed(ctx, req.Msg.Id, req.Msg.ErrorMessage); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.MarkNotificationFailedResponse{}), nil
}

// TestSend sends a test notification.
func (h *NotificationConnectHandler) TestSend(ctx context.Context, req *connect.Request[devicepb.TestSendRequest]) (*connect.Response[devicepb.TestSendResponse], error) {
	if h.sender == nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("notification sender unavailable"))
	}
	if err := h.sender.Send(ctx, req.Msg.Phone, req.Msg.Content); err != nil {
		return nil, response.MapDomainError(err)
	}
	return connect.NewResponse(&devicepb.TestSendResponse{Sent: true}), nil
}
