package notification

import (
	"context"
	"testing"
	"time"

	domainNotification "github.com/quixiq/polyglot/internal/domain/notification"
)

type mockNotificationRepo struct {
	templates []domainNotification.NotificationTemplate
	pending   []domainNotification.WANotification
	sentIDs   []string
	failedIDs []string
}

func (m *mockNotificationRepo) SaveTemplate(ctx context.Context, t domainNotification.NotificationTemplate) error {
	m.templates = append(m.templates, t)
	return nil
}

func (m *mockNotificationRepo) FindTemplateByKey(ctx context.Context, tenantID, key string) (domainNotification.NotificationTemplate, error) {
	for _, t := range m.templates {
		if t.TemplateKey == key {
			return t, nil
		}
	}
	return domainNotification.NotificationTemplate{}, domainNotification.ErrInvalidInput
}

func (m *mockNotificationRepo) ListTemplates(ctx context.Context, activeOnly bool) ([]domainNotification.NotificationTemplate, error) {
	return m.templates, nil
}

func (m *mockNotificationRepo) Queue(ctx context.Context, n domainNotification.WANotification) error {
	m.pending = append(m.pending, n)
	return nil
}

func (m *mockNotificationRepo) FindByID(ctx context.Context, id string) (domainNotification.WANotification, error) {
	for _, n := range m.pending {
		if n.ID == id {
			return n, nil
		}
	}
	return domainNotification.WANotification{}, domainNotification.ErrInvalidInput
}

func (m *mockNotificationRepo) Pending(ctx context.Context, limit int) ([]domainNotification.WANotification, error) {
	return m.pending, nil
}

func (m *mockNotificationRepo) MarkSent(ctx context.Context, id string, at time.Time) error {
	m.sentIDs = append(m.sentIDs, id)
	return nil
}

func (m *mockNotificationRepo) MarkFailed(ctx context.Context, id, err string) error {
	m.failedIDs = append(m.failedIDs, id)
	return nil
}

func (m *mockNotificationRepo) ListByCustomer(ctx context.Context, customerID string, limit int) ([]domainNotification.WANotification, error) {
	return m.pending, nil
}

type mockNotificationSender struct {
	lastPhone   string
	lastContent string
}

func (m *mockNotificationSender) Send(ctx context.Context, phone, content string) error {
	m.lastPhone = phone
	m.lastContent = content
	return nil
}

func TestManageNotificationUseCase(t *testing.T) {
	ctx := context.Background()
	repo := &mockNotificationRepo{
		templates: []domainNotification.NotificationTemplate{
			{ID: "nt_1", TemplateKey: "BILL_REMINDER", Name: "Bill Reminder", IsActive: true},
		},
		pending: []domainNotification.WANotification{
			{ID: "wa_1", RecipientPhone: "08123456789", MessageContent: "Hello"},
		},
	}
	sender := &mockNotificationSender{}
	uc := NewManageNotificationUseCase(repo, sender)

	// ListTemplates
	tmpls, err := uc.ListTemplates(ctx, true)
	if err != nil || len(tmpls) != 1 {
		t.Fatalf("unexpected ListTemplates: %v, %v", tmpls, err)
	}

	// GetTemplate
	tmpl, err := uc.GetTemplate(ctx, "BILL_REMINDER")
	if err != nil || tmpl.ID != "nt_1" {
		t.Fatalf("unexpected GetTemplate: %v, %v", tmpl, err)
	}

	// SaveTemplate
	saved, err := uc.SaveTemplate(ctx, domainNotification.NotificationTemplate{
		TemplateKey: "NEW_KEY",
		Name:        "New Key",
	})
	if err != nil || saved.ID == "" || saved.TenantID != "tenant-default" {
		t.Fatalf("unexpected SaveTemplate: %v, %v", saved, err)
	}

	// ListNotifications
	notifs, err := uc.ListNotifications(ctx, "", 10)
	if err != nil || len(notifs) != 1 {
		t.Fatalf("unexpected ListNotifications: %v, %v", notifs, err)
	}

	// PendingCount
	count, err := uc.PendingCount(ctx)
	if err != nil || count != 1 {
		t.Fatalf("unexpected PendingCount: %d, %v", count, err)
	}

	// MarkSent
	if err := uc.MarkSent(ctx, "wa_1"); err != nil || len(repo.sentIDs) != 1 {
		t.Fatalf("unexpected MarkSent: %v", err)
	}

	// MarkFailed
	if err := uc.MarkFailed(ctx, "wa_1", "timeout"); err != nil || len(repo.failedIDs) != 1 {
		t.Fatalf("unexpected MarkFailed: %v", err)
	}

	// TestSend
	if err := uc.TestSend(ctx, "08111111", "Test message"); err != nil || sender.lastPhone != "08111111" {
		t.Fatalf("unexpected TestSend: %v", err)
	}
}
