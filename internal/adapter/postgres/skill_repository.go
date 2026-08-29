package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/internal/port"
)

// SkillRepository implements port.SkillRepository for GORM/Postgres.
type SkillRepository struct {
	db *gorm.DB
}

var _ port.SkillRepository = (*SkillRepository)(nil)

// NewSkillRepository creates a new SkillRepository.
func NewSkillRepository(db *gorm.DB) *SkillRepository {
	return &SkillRepository{db: db}
}

func (r *SkillRepository) ListSkills(ctx context.Context, userID string) ([]skill.SkillMetadataRecord, error) {
	var models []model.SkillMetadataModel
	q := r.db.WithContext(ctx).Order("name ASC")
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if err := q.Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]skill.SkillMetadataRecord, 0, len(models))
	for _, m := range models {
		if d := m.ToDomain(); d != nil {
			result = append(result, *d)
		}
	}
	return result, nil
}

func (r *SkillRepository) GetSkill(ctx context.Context, userID, name string) (*skill.SkillMetadataRecord, error) {
	var m model.SkillMetadataModel
	q := r.db.WithContext(ctx).Where("name = ?", name)
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	if err := q.First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, skill.ErrSkillNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *SkillRepository) SaveSkillMetadata(ctx context.Context, rec *skill.SkillMetadataRecord) error {
	if rec == nil {
		return errors.New("skill metadata record cannot be nil")
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	rec.UpdatedAt = time.Now()
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = rec.UpdatedAt
	}

	// Truncate definition to 500 chars to keep DB light (LocalAI standard)
	if len(rec.Definition) > 500 {
		rec.Definition = rec.Definition[:500]
	}

	var existing model.SkillMetadataModel
	q := r.db.WithContext(ctx).Where("name = ?", rec.Name)
	if rec.UserID != "" {
		q = q.Where("user_id = ?", rec.UserID)
	}
	err := q.First(&existing).Error
	if err == nil {
		rec.ID = existing.ID
		rec.CreatedAt = existing.CreatedAt
		m := model.SkillMetadataModelFromDomain(rec)
		return r.db.WithContext(ctx).Model(&existing).Updates(m).Error
	}

	m := model.SkillMetadataModelFromDomain(rec)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *SkillRepository) DeleteSkillMetadata(ctx context.Context, userID, name string) error {
	q := r.db.WithContext(ctx).Where("name = ?", name)
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	return q.Delete(&model.SkillMetadataModel{}).Error
}

func (r *SkillRepository) ListGitSkills(ctx context.Context) ([]skill.SkillMetadataRecord, error) {
	var models []model.SkillMetadataModel
	if err := r.db.WithContext(ctx).Where("source_type = ? AND enabled = true", "git").Order("name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]skill.SkillMetadataRecord, 0, len(models))
	for _, m := range models {
		if d := m.ToDomain(); d != nil {
			result = append(result, *d)
		}
	}
	return result, nil
}

func (r *SkillRepository) ToggleSkillEnabled(ctx context.Context, userID, name string, enabled bool) error {
	q := r.db.WithContext(ctx).Model(&model.SkillMetadataModel{}).Where("name = ?", name)
	if userID != "" {
		q = q.Where("user_id = ?", userID)
	}
	res := q.Update("enabled", enabled)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return skill.ErrSkillNotFound
	}
	return nil
}

func (r *SkillRepository) GetGlobalSystemPrompt(ctx context.Context) (string, error) {
	var m model.GlobalPromptModel
	if err := r.db.WithContext(ctx).Where("key = ?", "default").First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return m.Content, nil
}

func (r *SkillRepository) SaveGlobalSystemPrompt(ctx context.Context, content string) error {
	var m model.GlobalPromptModel
	err := r.db.WithContext(ctx).Where("key = ?", "default").First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m = model.GlobalPromptModel{
				Key:     "default",
				Content: content,
			}
			return r.db.WithContext(ctx).Create(&m).Error
		}
		return err
	}

	m.Content = content
	return r.db.WithContext(ctx).Save(&m).Error
}

// ─── Backward compatibility delegations on Store ──────────────────────────────

func (s *Store) ListSkills(ctx context.Context, userID string) ([]skill.SkillMetadataRecord, error) {
	return NewSkillRepository(s.db).ListSkills(ctx, userID)
}

func (s *Store) GetSkill(ctx context.Context, userID, name string) (*skill.SkillMetadataRecord, error) {
	return NewSkillRepository(s.db).GetSkill(ctx, userID, name)
}

func (s *Store) SaveSkillMetadata(ctx context.Context, rec *skill.SkillMetadataRecord) error {
	return NewSkillRepository(s.db).SaveSkillMetadata(ctx, rec)
}

func (s *Store) DeleteSkillMetadata(ctx context.Context, userID, name string) error {
	return NewSkillRepository(s.db).DeleteSkillMetadata(ctx, userID, name)
}

func (s *Store) ListGitSkills(ctx context.Context) ([]skill.SkillMetadataRecord, error) {
	return NewSkillRepository(s.db).ListGitSkills(ctx)
}

func (s *Store) ToggleSkillEnabled(ctx context.Context, userID, name string, enabled bool) error {
	return NewSkillRepository(s.db).ToggleSkillEnabled(ctx, userID, name, enabled)
}

func (s *Store) GetGlobalSystemPrompt(ctx context.Context) (string, error) {
	return NewSkillRepository(s.db).GetGlobalSystemPrompt(ctx)
}

func (s *Store) SaveGlobalSystemPrompt(ctx context.Context, content string) error {
	return NewSkillRepository(s.db).SaveGlobalSystemPrompt(ctx, content)
}
