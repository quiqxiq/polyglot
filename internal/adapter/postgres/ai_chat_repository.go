package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/aichat"
	"github.com/quixiq/polyglot/internal/port"
)

type AIChatRepository struct {
	db *gorm.DB
}

var _ port.AIChatRepository = (*AIChatRepository)(nil)

func NewAIChatRepository(db *gorm.DB) *AIChatRepository { return &AIChatRepository{db: db} }

func (r *AIChatRepository) CreateConversation(ctx context.Context, c aichat.Conversation) error {
	return r.db.WithContext(ctx).Create(model.AIConversationModelFromDomain(c)).Error
}

func (r *AIChatRepository) FindConversation(ctx context.Context, id string) (aichat.Conversation, error) {
	var m model.AIConversationModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return aichat.Conversation{}, aichat.ErrNotFound
		}
		return aichat.Conversation{}, err
	}
	return m.ToDomain(), nil
}

func (r *AIChatRepository) UpdateConversation(ctx context.Context, c aichat.Conversation) error {
	m := model.AIConversationModelFromDomain(c)
	m.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *AIChatRepository) AppendEvent(ctx context.Context, conversationID, eventType string, payload json.RawMessage) (aichat.Event, error) {
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	var out aichat.Event
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var last model.AIChatEventModel
		q := tx.Where("conversation_id = ?", conversationID).Order("seq DESC").Limit(1)
		if r.db.Dialector.Name() == "postgres" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		seq := int64(1)
		if err := q.First(&last).Error; err == nil {
			seq = last.Seq + 1
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		m := &model.AIChatEventModel{ConversationID: conversationID, Seq: seq, EventType: eventType, Payload: payload}
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		out = m.ToDomain()
		return nil
	})
	return out, err
}

func (r *AIChatRepository) ListEvents(ctx context.Context, conversationID string, sinceSeq int64, limit int) ([]aichat.Event, error) {
	if limit <= 0 {
		limit = 200
	}
	var models []model.AIChatEventModel
	q := r.db.WithContext(ctx).Where("conversation_id = ? AND seq > ?", conversationID, sinceSeq).Order("seq ASC").Limit(limit)
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]aichat.Event, len(models))
	for i := range models {
		out[i] = models[i].ToDomain()
	}
	return out, nil
}

// DeleteEventsBefore removes chat events older than cutoff.
func (r *AIChatRepository) DeleteEventsBefore(ctx context.Context, cutoff time.Time) (int, error) {
	res := r.db.WithContext(ctx).Where("created_at < ?", cutoff).Delete(&model.AIChatEventModel{})
	if res.Error != nil {
		return 0, res.Error
	}
	return int(res.RowsAffected), nil
}
