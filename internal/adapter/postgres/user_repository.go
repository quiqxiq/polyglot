package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/port"
)

// ErrInvalidArgument indicates invalid user repository input parameters.
var ErrInvalidArgument = customer.ErrInvalidInput

type UserRepository struct {
	db *gorm.DB
}

var _ port.UserRepository = (*UserRepository)(nil)

// NewUserRepository returns an implementation of port.UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *customer.User) error {
	m := model.UserModelFromDomain(user)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*customer.User, error) {
	var m model.UserModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	u := m.ToDomain()
	u.AssignedDeviceIDs, _ = r.GetAssignedDeviceIDs(ctx, u.ID)
	return u, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*customer.User, error) {
	var m model.UserModel
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	u := m.ToDomain()
	u.AssignedDeviceIDs, _ = r.GetAssignedDeviceIDs(ctx, u.ID)
	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*customer.User, error) {
	var m model.UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	u := m.ToDomain()
	u.AssignedDeviceIDs, _ = r.GetAssignedDeviceIDs(ctx, u.ID)
	return u, nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserModel{}).Count(&count).Error
	return count, err
}

func (r *UserRepository) FindAll(ctx context.Context) ([]*customer.User, error) {
	var ms []model.UserModel
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	users := make([]*customer.User, 0, len(ms))
	userIDs := make([]uint, 0, len(ms))
	for i := range ms {
		u := ms[i].ToDomain()
		users = append(users, u)
		userIDs = append(userIDs, u.ID)
	}
	if len(userIDs) > 0 {
		var udList []model.UserDeviceModel
		if err := r.db.WithContext(ctx).Where("user_id IN (?)", userIDs).Find(&udList).Error; err == nil {
			devMap := make(map[uint][]string)
			for _, ud := range udList {
				devMap[ud.UserID] = append(devMap[ud.UserID], ud.DeviceID)
			}
			for _, u := range users {
				u.AssignedDeviceIDs = devMap[u.ID]
			}
		}
	}
	return users, nil
}

func (r *UserRepository) List(ctx context.Context, page, pageSize int, search string) ([]*customer.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&model.UserModel{})
	if s := strings.TrimSpace(search); s != "" {
		like := "%" + s + "%"
		if r.db.Dialector.Name() == "postgres" {
			q = q.Where("username ILIKE ? OR email ILIKE ?", like, like)
		} else {
			q = q.Where("username LIKE ? OR email LIKE ?", like, like)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var ms []model.UserModel
	if err := q.Order("id ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&ms).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*customer.User, 0, len(ms))
	userIDs := make([]uint, 0, len(ms))
	for i := range ms {
		u := ms[i].ToDomain()
		users = append(users, u)
		userIDs = append(userIDs, u.ID)
	}
	if len(userIDs) > 0 {
		var udList []model.UserDeviceModel
		if err := r.db.WithContext(ctx).Where("user_id IN (?)", userIDs).Find(&udList).Error; err == nil {
			devMap := make(map[uint][]string)
			for _, ud := range udList {
				devMap[ud.UserID] = append(devMap[ud.UserID], ud.DeviceID)
			}
			for _, u := range users {
				u.AssignedDeviceIDs = devMap[u.ID]
			}
		}
	}
	return users, total, nil
}

func (r *UserRepository) Update(ctx context.Context, u *customer.User) error {
	if u == nil || u.ID == 0 {
		return ErrInvalidArgument
	}
	res := r.db.WithContext(ctx).Model(&model.UserModel{}).Where("id = ?", u.ID).Updates(map[string]any{
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

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.UserModel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uint, hash string) error {
	if id == 0 || hash == "" {
		return ErrInvalidArgument
	}
	res := r.db.WithContext(ctx).Model(&model.UserModel{}).Where("id = ?", id).Update("password_hash", hash)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id uint, active bool) error {
	if id == 0 {
		return ErrInvalidArgument
	}
	res := r.db.WithContext(ctx).Model(&model.UserModel{}).Where("id = ?", id).Update("is_active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Backward compatibility methods on Store delegating to UserRepository
func (s *Store) CreateUser(ctx context.Context, user *customer.User) error {
	return NewUserRepository(s.db).Create(ctx, user)
}

func (s *Store) FindUserByID(ctx context.Context, id uint) (*customer.User, error) {
	return NewUserRepository(s.db).FindByID(ctx, id)
}

func (s *Store) FindUserByUsername(ctx context.Context, username string) (*customer.User, error) {
	return NewUserRepository(s.db).FindByUsername(ctx, username)
}

func (s *Store) FindUserByEmail(ctx context.Context, email string) (*customer.User, error) {
	return NewUserRepository(s.db).FindByEmail(ctx, email)
}

func (s *Store) CountUsers(ctx context.Context) (int64, error) {
	return NewUserRepository(s.db).Count(ctx)
}

func (s *Store) FindAllUsers(ctx context.Context) ([]*customer.User, error) {
	return NewUserRepository(s.db).FindAll(ctx)
}

func (s *Store) ListUsers(ctx context.Context, page, pageSize int, search string) ([]*customer.User, int64, error) {
	return NewUserRepository(s.db).List(ctx, page, pageSize, search)
}

func (s *Store) UpdateUser(ctx context.Context, u *customer.User) error {
	return NewUserRepository(s.db).Update(ctx, u)
}

func (s *Store) DeleteUser(ctx context.Context, id uint) error {
	return NewUserRepository(s.db).Delete(ctx, id)
}

func (s *Store) SetPasswordHash(ctx context.Context, id uint, hash string) error {
	return NewUserRepository(s.db).UpdatePassword(ctx, id, hash)
}

func (r *UserRepository) FindByRoles(ctx context.Context, roles []string, activeOnly bool) ([]*customer.User, error) {
	q := r.db.WithContext(ctx).Model(&model.UserModel{})
	if len(roles) > 0 {
		q = q.Where("role IN (?)", roles)
	}
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	var ms []model.UserModel
	if err := q.Order("id ASC").Find(&ms).Error; err != nil {
		return nil, err
	}
	users := make([]*customer.User, 0, len(ms))
	for i := range ms {
		users = append(users, ms[i].ToDomain())
	}
	return users, nil
}

func (s *Store) SetUserActive(ctx context.Context, id uint, active bool) error {
	return NewUserRepository(s.db).UpdateStatus(ctx, id, active)
}

func (s *Store) FindUsersByRoles(ctx context.Context, roles []string, activeOnly bool) ([]*customer.User, error) {
	return NewUserRepository(s.db).FindByRoles(ctx, roles, activeOnly)
}

func (r *UserRepository) AssignDevices(ctx context.Context, userID uint, deviceIDs []string, assignedBy *uint) error {
	if userID == 0 {
		return ErrInvalidArgument
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserDeviceModel{}).Error; err != nil {
			return err
		}

		if len(deviceIDs) == 0 {
			return nil
		}

		seen := make(map[string]bool)
		var models []model.UserDeviceModel
		for _, devID := range deviceIDs {
			devID = strings.TrimSpace(devID)
			if devID != "" && !seen[devID] {
				seen[devID] = true
				models = append(models, model.UserDeviceModel{
					UserID:     userID,
					DeviceID:   devID,
					AssignedBy: assignedBy,
				})
			}
		}

		if len(models) > 0 {
			if err := tx.Create(&models).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *UserRepository) GetAssignedDeviceIDs(ctx context.Context, userID uint) ([]string, error) {
	if userID == 0 {
		return nil, nil
	}
	var deviceIDs []string
	err := r.db.WithContext(ctx).Model(&model.UserDeviceModel{}).
		Where("user_id = ?", userID).
		Order("device_id ASC").
		Pluck("device_id", &deviceIDs).Error
	if err != nil {
		return nil, err
	}
	return deviceIDs, nil
}

func (r *UserRepository) IsDeviceAccessibleByUser(ctx context.Context, userID uint, deviceID string) (bool, error) {
	if userID == 0 || deviceID == "" {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserDeviceModel{}).
		Where("user_id = ? AND device_id = ?", userID, deviceID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
