package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/auth"
	"github.com/quixiq/polyglot/internal/port"
)

type userRepository struct {
	db *gorm.DB
}

var _ port.UserRepository = (*userRepository)(nil)

// NewUserRepository creates a port.UserRepository backed by GORM/PostgreSQL.
func NewUserRepository(db *gorm.DB) port.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Save(ctx context.Context, user *auth.User) error {
	m := models.UserModelFromDomain(user)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*auth.User, error) {
	var m models.UserModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*auth.User, error) {
	var m models.UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.UserModel{}).Count(&count).Error
	return count, err
}

// Store backward compatibility methods
func (s *Store) CreateUser(user *auth.User) error {
	repo := NewUserRepository(s.db)
	return repo.Save(context.Background(), user)
}

func (s *Store) FindUserByID(id uint) (*auth.User, error) {
	repo := NewUserRepository(s.db)
	return repo.FindByID(context.Background(), id)
}

func (s *Store) FindUserByEmail(email string) (*auth.User, error) {
	repo := NewUserRepository(s.db)
	return repo.FindByEmail(context.Background(), email)
}

func (s *Store) CountUsers() (int64, error) {
	repo := NewUserRepository(s.db)
	return repo.Count(context.Background())
}
