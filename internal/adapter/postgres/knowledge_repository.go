package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

func (s *Store) CreateKnowledgeEntry(ctx context.Context, entry *knowledge.Entry) error {
	m := model.KnowledgeEntryModelFromDomain(entry)
	if err := s.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	entry.ID = m.ID
	return nil
}

func (s *Store) CreateKnowledge(ctx context.Context, entry *knowledge.Entry) error {
	return s.CreateKnowledgeEntry(ctx, entry)
}

func (s *Store) FindKnowledgeEntryByID(ctx context.Context, id uint) (*knowledge.Entry, error) {
	var m model.KnowledgeEntryModel
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, knowledge.ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindAllKnowledgeEntries(ctx context.Context) ([]knowledge.Entry, error) {
	var mList []model.KnowledgeEntryModel
	if err := s.db.WithContext(ctx).Order("created_at DESC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []knowledge.Entry
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (s *Store) FindAllKnowledge(ctx context.Context) ([]knowledge.Entry, error) {
	return s.FindAllKnowledgeEntries(ctx)
}

func (s *Store) FindKnowledgeByID(ctx context.Context, id uint) (*knowledge.Entry, error) {
	return s.FindKnowledgeEntryByID(ctx, id)
}

func (s *Store) UpdateKnowledge(ctx context.Context, entry *knowledge.Entry) error {
	return s.UpdateKnowledgeEntry(ctx, entry)
}

func (s *Store) DeleteKnowledge(ctx context.Context, id uint) error {
	return s.DeleteKnowledgeEntry(ctx, id)
}

func (s *Store) UpdateKnowledgeEntry(ctx context.Context, entry *knowledge.Entry) error {
	m := model.KnowledgeEntryModelFromDomain(entry)
	return s.db.WithContext(ctx).Save(m).Error
}

func (s *Store) DeleteKnowledgeEntry(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&model.KnowledgeEntryModel{}, id).Error
}

func (s *Store) SearchKnowledgeByTags(ctx context.Context, tags []string) ([]knowledge.Entry, error) {
	if len(tags) == 0 {
		return s.FindAllKnowledgeEntries(ctx)
	}

	query := s.db.WithContext(ctx)
	for _, tag := range tags {
		query = query.Or("LOWER(tags) LIKE ?", "%"+tag+"%")
	}

	var mList []model.KnowledgeEntryModel
	if err := query.Find(&mList).Error; err != nil {
		return nil, err
	}

	var res []knowledge.Entry
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (s *Store) Retrieve(ctx context.Context, query string) ([]knowledge.Entry, error) {
	return s.FindAllKnowledgeEntries(ctx)
}
