package postgres

import (
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
)

func (s *Store) CreateKnowledgeEntry(entry *knowledge.KnowledgeEntry) error {
	m := models.KnowledgeEntryModelFromDomain(entry)
	if err := s.db.Create(m).Error; err != nil {
		return err
	}
	entry.ID = m.ID
	return nil
}

func (s *Store) CreateKnowledge(entry *knowledge.KnowledgeEntry) error {
	return s.CreateKnowledgeEntry(entry)
}

func (s *Store) FindKnowledgeEntryByID(id uint) (*knowledge.KnowledgeEntry, error) {
	var m models.KnowledgeEntryModel
	if err := s.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindAllKnowledgeEntries() ([]knowledge.KnowledgeEntry, error) {
	var mList []models.KnowledgeEntryModel
	if err := s.db.Order("created_at DESC").Find(&mList).Error; err != nil {
		return nil, err
	}
	var res []knowledge.KnowledgeEntry
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (s *Store) FindAllKnowledge() ([]knowledge.KnowledgeEntry, error) {
	return s.FindAllKnowledgeEntries()
}

func (s *Store) FindKnowledgeByID(id uint) (*knowledge.KnowledgeEntry, error) {
	return s.FindKnowledgeEntryByID(id)
}

func (s *Store) UpdateKnowledge(entry *knowledge.KnowledgeEntry) error {
	return s.UpdateKnowledgeEntry(entry)
}

func (s *Store) DeleteKnowledge(id uint) error {
	return s.DeleteKnowledgeEntry(id)
}

func (s *Store) UpdateKnowledgeEntry(entry *knowledge.KnowledgeEntry) error {
	m := models.KnowledgeEntryModelFromDomain(entry)
	return s.db.Save(m).Error
}

func (s *Store) DeleteKnowledgeEntry(id uint) error {
	return s.db.Delete(&models.KnowledgeEntryModel{}, id).Error
}

func (s *Store) SearchKnowledgeByTags(tags []string) ([]knowledge.KnowledgeEntry, error) {
	if len(tags) == 0 {
		return s.FindAllKnowledgeEntries()
	}

	query := s.db
	for _, tag := range tags {
		query = query.Or("LOWER(tags) LIKE ?", "%"+tag+"%")
	}

	var mList []models.KnowledgeEntryModel
	if err := query.Find(&mList).Error; err != nil {
		return nil, err
	}

	var res []knowledge.KnowledgeEntry
	for _, m := range mList {
		if d := m.ToDomain(); d != nil {
			res = append(res, *d)
		}
	}
	return res, nil
}
