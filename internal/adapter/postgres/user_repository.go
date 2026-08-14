package postgres

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

// ErrInvalidArgument is returned when a store mutation is called with
// missing/invalid required values (empty ID, empty password hash, ...).
var ErrInvalidArgument = errors.New("invalid argument")

func (s *Store) CreateUser(user *customer.User) error {
	m := models.UserModelFromDomain(user)
	if err := s.db.Create(m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	return nil
}

func (s *Store) FindUserByID(id uint) (*customer.User, error) {
	var m models.UserModel
	if err := s.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindUserByUsername(username string) (*customer.User, error) {
	var m models.UserModel
	if err := s.db.Where("username = ?", username).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) FindUserByEmail(email string) (*customer.User, error) {
	var m models.UserModel
	if err := s.db.Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return m.ToDomain(), nil
}

func (s *Store) CountUsers() (int64, error) {
	var count int64
	err := s.db.Model(&models.UserModel{}).Count(&count).Error
	return count, err
}

// FindAllUsers returns all registered users, ordered by ID. Used by
// EnsureUserRoleAssignments to sync Casbin role assignments.
func (s *Store) FindAllUsers() ([]*customer.User, error) {
	var ms []models.UserModel
	if err := s.db.Order("id ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	users := make([]*customer.User, 0, len(ms))
	for i := range ms {
		users = append(users, ms[i].ToDomain())
	}
	return users, nil
}

// ListUsers returns a page of users ordered by ID ascending, with an optional
// case-insensitive search on username/email. Returns the users plus the total
// count matching the filter (before pagination).
func (s *Store) ListUsers(page, pageSize int, search string) ([]*customer.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := s.db.Model(&models.UserModel{})
	if s := strings.TrimSpace(search); s != "" {
		like := "%" + s + "%"
		q = q.Where("username ILIKE ? OR email ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var ms []models.UserModel
	if err := q.Order("id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&ms).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*customer.User, 0, len(ms))
	for i := range ms {
		users = append(users, ms[i].ToDomain())
	}
	return users, total, nil
}

// UpdateUser persists the mutable fields (username, email, role) of an
// existing user. PasswordHash/IsActive tidak diubah di sini — gunakan
// SetPasswordHash/SetUserActive.
func (s *Store) UpdateUser(u *customer.User) error {
	if u == nil || u.ID == 0 {
		return ErrInvalidArgument
	}
	res := s.db.Model(&models.UserModel{}).Where("id = ?", u.ID).Updates(map[string]interface{}{
		"username": u.Username,
		"email":    u.Email,
		"role":     u.Role,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteUser removes a user row entirely.
func (s *Store) DeleteUser(id uint) error {
	res := s.db.Delete(&models.UserModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// SetPasswordHash overwrites the bcrypt hash of a user's password.
func (s *Store) SetPasswordHash(id uint, hash string) error {
	if id == 0 || hash == "" {
		return ErrInvalidArgument
	}
	res := s.db.Model(&models.UserModel{}).Where("id = ?", id).Update("password_hash", hash)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// SetUserActive flips the is_active flag of a user.
func (s *Store) SetUserActive(id uint, active bool) error {
	if id == 0 {
		return ErrInvalidArgument
	}
	res := s.db.Model(&models.UserModel{}).Where("id = ?", id).Update("is_active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
