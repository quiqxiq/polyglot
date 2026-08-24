package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/notification"
	"github.com/quixiq/polyglot/internal/port"
)

type NotificationRepository struct {
	db *gorm.DB
}

var _ port.NotificationRepository = (*NotificationRepository)(nil)

// NewNotificationRepository returns a port.NotificationRepository backed by GORM/Postgres.
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) SaveTemplate(ctx context.Context, t notification.NotificationTemplate) error {
	return r.db.WithContext(ctx).Save(model.NotificationTemplateModelFromDomain(t)).Error
}

func (r *NotificationRepository) FindTemplateByKey(ctx context.Context, tenantID, key string) (notification.NotificationTemplate, error) {
	var m model.NotificationTemplateModel
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND template_key = ?", tenantID, key).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notification.NotificationTemplate{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *NotificationRepository) ListTemplates(ctx context.Context, activeOnly bool) ([]notification.NotificationTemplate, error) {
	q := r.db.WithContext(ctx)
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var mList []model.NotificationTemplateModel
	if err := q.Order("template_key").Find(&mList).Error; err != nil {
		return nil, err
	}
	out := make([]notification.NotificationTemplate, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

func (r *NotificationRepository) Queue(ctx context.Context, n notification.WANotification) error {
	m := model.WANotificationModelFromDomain(n)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	n.ID = m.ID
	n.CreatedAt = m.CreatedAt
	return nil
}

func (r *NotificationRepository) FindByID(ctx context.Context, id string) (notification.WANotification, error) {
	var m model.WANotificationModel
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notification.WANotification{}, ErrNotFound
	}
	return m.ToDomain(), err
}

func (r *NotificationRepository) Pending(ctx context.Context, limit int) ([]notification.WANotification, error) {
	var mList []model.WANotificationModel
	q := r.db.WithContext(ctx).Where("status = ?", notification.StatusQueued)
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Order("created_at asc").Find(&mList).Error; err != nil {
		return nil, err
	}
	out := make([]notification.WANotification, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

func (r *NotificationRepository) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.WANotificationModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": notification.StatusSent, "sent_at": sentAt})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) MarkFailed(ctx context.Context, id string, errMsg string) error {
	res := r.db.WithContext(ctx).Model(&model.WANotificationModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": notification.StatusFailed, "error_message": errMsg})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) ListByCustomer(ctx context.Context, customerID string, limit int) ([]notification.WANotification, error) {
	q := r.db.WithContext(ctx).Where("customer_id = ?", customerID)
	if limit > 0 {
		q = q.Limit(limit)
	}
	var mList []model.WANotificationModel
	if err := q.Order("created_at desc").Find(&mList).Error; err != nil {
		return nil, err
	}
	out := make([]notification.WANotification, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

// PendingWithRetryLimit implements port.NotificationRetryRepository.
func (r *NotificationRepository) PendingWithRetryLimit(ctx context.Context, limit, maxAttempts int) ([]notification.WANotification, error) {
	var mList []model.WANotificationModel
	// Retryable: QUEUED, atau FAILED yang masih punya sisa percobaan.
	q := r.db.WithContext(ctx).
		Where("(status = ?) OR (status = ? AND attempts < ?)",
			notification.StatusQueued, notification.StatusFailed, maxAttempts)
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Order("created_at asc").Find(&mList).Error; err != nil {
		return nil, err
	}
	out := make([]notification.WANotification, len(mList))
	for i := range mList {
		out[i] = mList[i].ToDomain()
	}
	return out, nil
}

// MarkFailedWithAttempt implements port.NotificationRetryRepository.
func (r *NotificationRepository) MarkFailedWithAttempt(ctx context.Context, id, errMsg string, attempts int) error {
	res := r.db.WithContext(ctx).Model(&model.WANotificationModel{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": notification.StatusFailed, "error_message": errMsg, "attempts": attempts})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
