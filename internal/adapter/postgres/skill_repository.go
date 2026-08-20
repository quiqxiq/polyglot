package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/internal/port"
)

var _ port.SkillRepository = (*Store)(nil)

func (s *Store) ListSkills(ctx context.Context) ([]skill.Skill, error) {
	var models []model.SkillModel
	if err := s.db.WithContext(ctx).Preload("Files").Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]skill.Skill, 0, len(models))
	for _, m := range models {
		if d := m.ToDomain(); d != nil {
			result = append(result, *d)
		}
	}
	return result, nil
}

func (s *Store) GetSkillByID(ctx context.Context, id uint) (*skill.Skill, error) {
	var m model.SkillModel
	if err := s.db.WithContext(ctx).Preload("Files").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, skill.ErrSkillNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) GetSkillBySlug(ctx context.Context, slug string) (*skill.Skill, error) {
	var m model.SkillModel
	if err := s.db.WithContext(ctx).Preload("Files").Where("slug = ?", slug).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, skill.ErrSkillNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) CreateSkill(ctx context.Context, sEntity *skill.Skill) error {
	m := model.SkillModelFromDomain(sEntity)
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	sEntity.ID = m.ID
	return nil
}

func (s *Store) UpdateSkill(ctx context.Context, sEntity *skill.Skill) error {
	m := model.SkillModelFromDomain(sEntity)
	return s.db.WithContext(ctx).Save(m).Error
}

func (s *Store) DeleteSkill(ctx context.Context, id uint) error {
	_ = s.db.WithContext(ctx).Where("skill_id = ?", id).Delete(&model.SkillFileModel{}).Error
	return s.db.WithContext(ctx).Delete(&model.SkillModel{}, id).Error
}

func (s *Store) ToggleSkillEnabled(ctx context.Context, slug string, enabled bool) error {
	res := s.db.WithContext(ctx).Model(&model.SkillModel{}).Where("slug = ?", slug).Update("is_enabled", enabled)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return skill.ErrSkillNotFound
	}
	return nil
}

func (s *Store) SaveSkillFile(ctx context.Context, skillID uint, f *skill.SkillFile) error {
	f.SkillID = skillID
	var existing model.SkillFileModel
	err := s.db.WithContext(ctx).Where("skill_id = ? AND file_path = ?", skillID, f.FilePath).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m := model.SkillFileModelFromDomain(f)
			if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
				return err
			}
			f.ID = m.ID
			return nil
		}
		return err
	}

	existing.Content = f.Content
	existing.Name = f.Name
	existing.IsReference = f.IsReference
	if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return err
	}
	f.ID = existing.ID
	return nil
}

func (s *Store) DeleteSkillFile(ctx context.Context, fileID uint) error {
	return s.db.WithContext(ctx).Delete(&model.SkillFileModel{}, fileID).Error
}

func (s *Store) DeleteSkillFileByPath(ctx context.Context, skillID uint, filePath string) error {
	return s.db.WithContext(ctx).Where("skill_id = ? AND file_path = ?", skillID, filePath).Delete(&model.SkillFileModel{}).Error
}

func (s *Store) GetGlobalSystemPrompt(ctx context.Context) (string, error) {
	var m model.GlobalPromptModel
	if err := s.db.WithContext(ctx).Where("key = ?", "default").First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return m.Content, nil
}

func (s *Store) SaveGlobalSystemPrompt(ctx context.Context, content string) error {
	var m model.GlobalPromptModel
	err := s.db.WithContext(ctx).Where("key = ?", "default").First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			m = model.GlobalPromptModel{
				Key:     "default",
				Content: content,
			}
			return s.db.WithContext(ctx).Create(&m).Error
		}
		return err
	}

	m.Content = content
	return s.db.WithContext(ctx).Save(&m).Error
}
