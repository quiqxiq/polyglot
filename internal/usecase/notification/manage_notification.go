package notification

import (
	"context"
	"fmt"
	"time"

	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/port"
	"github.com/quixiq/polyglot/pkg/idgen"
)

// ManageNotificationUseCase orchestrates notification templates, delivery status, and test dispatch.
type ManageNotificationUseCase struct {
	repo   port.NotificationRepository
	sender port.NotificationSender
}

// NewManageNotificationUseCase constructs a new ManageNotificationUseCase.
func NewManageNotificationUseCase(repo port.NotificationRepository, sender port.NotificationSender) *ManageNotificationUseCase {
	return &ManageNotificationUseCase{
		repo:   repo,
		sender: sender,
	}
}

// ListTemplates returns notification templates, optionally filtered by active status.
func (uc *ManageNotificationUseCase) ListTemplates(ctx context.Context, activeOnly bool) ([]domainNotification.NotificationTemplate, error) {
	if uc.repo == nil {
		return nil, domainNotification.ErrRepositoryUnavailable
	}
	templates, err := uc.repo.ListTemplates(ctx, activeOnly)
	if err != nil {
		return nil, fmt.Errorf("list notification templates: %w", err)
	}
	return templates, nil
}

// GetTemplate returns a notification template by key for default tenant.
func (uc *ManageNotificationUseCase) GetTemplate(ctx context.Context, templateKey string) (domainNotification.NotificationTemplate, error) {
	if uc.repo == nil {
		return domainNotification.NotificationTemplate{}, domainNotification.ErrRepositoryUnavailable
	}
	tmpl, err := uc.repo.FindTemplateByKey(ctx, "tenant-default", templateKey)
	if err != nil {
		return domainNotification.NotificationTemplate{}, fmt.Errorf("find template by key: %w", err)
	}
	return tmpl, nil
}

// SaveTemplate creates or updates a notification template.
func (uc *ManageNotificationUseCase) SaveTemplate(ctx context.Context, tmpl domainNotification.NotificationTemplate) (domainNotification.NotificationTemplate, error) {
	if uc.repo == nil {
		return domainNotification.NotificationTemplate{}, domainNotification.ErrRepositoryUnavailable
	}
	if tmpl.ID == "" {
		tmpl.ID = idgen.New("nt")
	}
	if tmpl.TenantID == "" {
		tmpl.TenantID = "tenant-default"
	}
	if err := uc.repo.SaveTemplate(ctx, tmpl); err != nil {
		return domainNotification.NotificationTemplate{}, fmt.Errorf("save template: %w", err)
	}
	return tmpl, nil
}

// ListNotifications returns notification records by customer or pending queue.
func (uc *ManageNotificationUseCase) ListNotifications(ctx context.Context, customerID string, limit int) ([]domainNotification.WANotification, error) {
	if uc.repo == nil {
		return nil, domainNotification.ErrRepositoryUnavailable
	}
	if customerID != "" {
		list, err := uc.repo.ListByCustomer(ctx, customerID, limit)
		if err != nil {
			return nil, fmt.Errorf("list notifications by customer: %w", err)
		}
		return list, nil
	}
	list, err := uc.repo.Pending(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list pending notifications: %w", err)
	}
	return list, nil
}

// PendingCount returns the number of pending notifications.
func (uc *ManageNotificationUseCase) PendingCount(ctx context.Context) (int, error) {
	if uc.repo == nil {
		return 0, domainNotification.ErrRepositoryUnavailable
	}
	pending, err := uc.repo.Pending(ctx, 0)
	if err != nil {
		return 0, fmt.Errorf("count pending notifications: %w", err)
	}
	return len(pending), nil
}

// MarkSent marks a notification as successfully delivered.
func (uc *ManageNotificationUseCase) MarkSent(ctx context.Context, id string) error {
	if uc.repo == nil {
		return domainNotification.ErrRepositoryUnavailable
	}
	if err := uc.repo.MarkSent(ctx, id, time.Now()); err != nil {
		return fmt.Errorf("mark notification sent: %w", err)
	}
	return nil
}

// MarkFailed marks a notification as failed with an error message.
func (uc *ManageNotificationUseCase) MarkFailed(ctx context.Context, id, errMsg string) error {
	if uc.repo == nil {
		return domainNotification.ErrRepositoryUnavailable
	}
	if err := uc.repo.MarkFailed(ctx, id, errMsg); err != nil {
		return fmt.Errorf("mark notification failed: %w", err)
	}
	return nil
}

// TestSend dispatches a test notification message through the notification sender.
func (uc *ManageNotificationUseCase) TestSend(ctx context.Context, phone, content string) error {
	if uc.sender == nil {
		return domainNotification.ErrNoWASession
	}
	if err := uc.sender.Send(ctx, phone, content); err != nil {
		return fmt.Errorf("test send notification: %w", err)
	}
	return nil
}
