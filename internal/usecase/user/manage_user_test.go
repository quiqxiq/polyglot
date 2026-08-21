package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/usecase/user"
)

type mockUserRepo struct {
	users map[uint]*customer.User
	seq   uint
}

func newMockRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[uint]*customer.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *customer.User) error {
	m.seq++
	u.ID = m.seq
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint) (*customer.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	// Return shallow copy
	copyU := *u
	return &copyU, nil
}

func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (*customer.User, error) {
	for _, u := range m.users {
		if u.Username == username {
			copyU := *u
			return &copyU, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*customer.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			copyU := *u
			return &copyU, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func (m *mockUserRepo) List(ctx context.Context, page, pageSize int, search string) ([]*customer.User, int64, error) {
	var list []*customer.User
	for _, u := range m.users {
		copyU := *u
		list = append(list, &copyU)
	}
	return list, int64(len(list)), nil
}

func (m *mockUserRepo) FindAll(ctx context.Context) ([]*customer.User, error) {
	list, _, err := m.List(ctx, 1, 1000, "")
	return list, err
}

func (m *mockUserRepo) FindByRoles(ctx context.Context, roles []string, activeOnly bool) ([]*customer.User, error) {
	var list []*customer.User
	for _, u := range m.users {
		for _, r := range roles {
			if u.Role == r {
				if !activeOnly || u.IsActive {
					copyU := *u
					list = append(list, &copyU)
				}
				break
			}
		}
	}
	return list, nil
}

func (m *mockUserRepo) Update(ctx context.Context, u *customer.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id uint) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uint, passwordHash string) error {
	if u, ok := m.users[id]; ok {
		u.PasswordHash = passwordHash
		return nil
	}
	return user.ErrUserNotFound
}

func (m *mockUserRepo) UpdateStatus(ctx context.Context, id uint, isActive bool) error {
	if u, ok := m.users[id]; ok {
		u.IsActive = isActive
		return nil
	}
	return user.ErrUserNotFound
}

func TestHierarchy_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Admin cannot create admin user", func(t *testing.T) {
		repo := newMockRepo()
		uc := user.NewManageUserUseCase(repo, nil)

		_, err := uc.CreateUser(ctx, 2, []string{"admin"}, "newadmin", "na@ex.com", "password123", "admin", "Admin 2", "", "")
		if !errors.Is(err, user.ErrAdminCannotCreateAdminOrOwner) {
			t.Fatalf("expected ErrAdminCannotCreateAdminOrOwner, got %v", err)
		}
	})

	t.Run("Admin cannot create owner user", func(t *testing.T) {
		repo := newMockRepo()
		uc := user.NewManageUserUseCase(repo, nil)

		_, err := uc.CreateUser(ctx, 2, []string{"admin"}, "newowner", "no@ex.com", "password123", "owner", "Owner 2", "", "")
		if !errors.Is(err, user.ErrAdminCannotCreateAdminOrOwner) {
			t.Fatalf("expected ErrAdminCannotCreateAdminOrOwner, got %v", err)
		}
	})

	t.Run("Admin can create agent and teknisi user", func(t *testing.T) {
		repo := newMockRepo()
		uc := user.NewManageUserUseCase(repo, nil)

		u1, err := uc.CreateUser(ctx, 2, []string{"admin"}, "agent1", "ag1@ex.com", "password123", "agent", "Agent One", "", "")
		if err != nil || u1 == nil {
			t.Fatalf("unexpected error creating agent: %v", err)
		}

		u2, err := uc.CreateUser(ctx, 2, []string{"admin"}, "tek1", "tk1@ex.com", "password123", "teknisi", "Teknisi One", "", "")
		if err != nil || u2 == nil {
			t.Fatalf("unexpected error creating teknisi: %v", err)
		}
	})

	t.Run("Owner can create any user role", func(t *testing.T) {
		repo := newMockRepo()
		uc := user.NewManageUserUseCase(repo, nil)

		uAdmin, err := uc.CreateUser(ctx, 1, []string{"owner"}, "subadmin", "sa@ex.com", "password123", "admin", "Sub Admin", "", "")
		if err != nil || uAdmin == nil {
			t.Fatalf("unexpected error owner creating admin: %v", err)
		}

		uOwner, err := uc.CreateUser(ctx, 1, []string{"owner"}, "coowner", "co@ex.com", "password123", "owner", "Co Owner", "", "")
		if err != nil || uOwner == nil {
			t.Fatalf("unexpected error owner creating owner: %v", err)
		}
	})
}

func TestHierarchy_UpdateUser(t *testing.T) {
	ctx := context.Background()

	setupRepo := func() (*mockUserRepo, *user.ManageUserUseCase) {
		repo := newMockRepo()
		repo.users[1] = &customer.User{ID: 1, Username: "owner1", Email: "owner1@ex.com", Role: "owner", IsActive: true}
		repo.users[2] = &customer.User{ID: 2, Username: "admin1", Email: "admin1@ex.com", Role: "admin", IsActive: true}
		repo.users[3] = &customer.User{ID: 3, Username: "admin2", Email: "admin2@ex.com", Role: "admin", IsActive: true}
		repo.users[4] = &customer.User{ID: 4, Username: "agent1", Email: "agent1@ex.com", Role: "agent", IsActive: true}
		repo.seq = 4
		return repo, user.NewManageUserUseCase(repo, nil)
	}

	t.Run("Admin cannot edit owner account", func(t *testing.T) {
		_, uc := setupRepo()
		_, err := uc.UpdateUser(ctx, 2, []string{"admin"}, 1, "owner1_edited", "", "", "", "", "")
		if !errors.Is(err, user.ErrCannotModifyOwner) {
			t.Fatalf("expected ErrCannotModifyOwner, got %v", err)
		}
	})

	t.Run("Another owner cannot edit owner account (strict self-only protection)", func(t *testing.T) {
		repo, uc := setupRepo()
		repo.users[5] = &customer.User{ID: 5, Username: "owner2", Role: "owner", IsActive: true}
		_, err := uc.UpdateUser(ctx, 5, []string{"owner"}, 1, "owner1_edited", "", "", "", "", "")
		if !errors.Is(err, user.ErrCannotModifyOwner) {
			t.Fatalf("expected ErrCannotModifyOwner, got %v", err)
		}
	})

	t.Run("Owner can edit own account", func(t *testing.T) {
		_, uc := setupRepo()
		u, err := uc.UpdateUser(ctx, 1, []string{"owner"}, 1, "owner1_new", "owner_new@ex.com", "", "Super Owner", "123", "Lead")
		if err != nil {
			t.Fatalf("expected owner self-edit to succeed, got %v", err)
		}
		if u.Username != "owner1_new" {
			t.Fatalf("expected updated username owner1_new, got %s", u.Username)
		}
	})

	t.Run("Admin can edit own account", func(t *testing.T) {
		_, uc := setupRepo()
		u, err := uc.UpdateUser(ctx, 2, []string{"admin"}, 2, "admin1_new", "admin1_new@ex.com", "", "Admin One", "456", "Ops")
		if err != nil {
			t.Fatalf("expected admin self-edit to succeed, got %v", err)
		}
		if u.Username != "admin1_new" {
			t.Fatalf("expected updated username admin1_new, got %s", u.Username)
		}
	})

	t.Run("Admin cannot edit another admin account", func(t *testing.T) {
		_, uc := setupRepo()
		_, err := uc.UpdateUser(ctx, 2, []string{"admin"}, 3, "admin2_hacked", "", "", "", "", "")
		if !errors.Is(err, user.ErrCannotModifyAdmin) {
			t.Fatalf("expected ErrCannotModifyAdmin, got %v", err)
		}
	})

	t.Run("Owner can edit admin account", func(t *testing.T) {
		_, uc := setupRepo()
		u, err := uc.UpdateUser(ctx, 1, []string{"owner"}, 2, "admin1_byowner", "", "", "", "", "")
		if err != nil {
			t.Fatalf("expected owner to edit admin, got %v", err)
		}
		if u.Username != "admin1_byowner" {
			t.Fatalf("expected admin1_byowner, got %s", u.Username)
		}
	})

	t.Run("Admin editing agent cannot promote agent to admin or owner", func(t *testing.T) {
		_, uc := setupRepo()
		_, err := uc.UpdateUser(ctx, 2, []string{"admin"}, 4, "", "", "admin", "", "", "")
		if !errors.Is(err, user.ErrAdminCannotAssignAdminOrOwner) {
			t.Fatalf("expected ErrAdminCannotAssignAdminOrOwner, got %v", err)
		}

		_, err = uc.UpdateUser(ctx, 2, []string{"admin"}, 4, "", "", "owner", "", "", "")
		if !errors.Is(err, user.ErrCannotAssignOwnerRole) {
			t.Fatalf("expected ErrCannotAssignOwnerRole, got %v", err)
		}
	})

	t.Run("Admin editing agent can change role to teknisi", func(t *testing.T) {
		_, uc := setupRepo()
		u, err := uc.UpdateUser(ctx, 2, []string{"admin"}, 4, "", "", "teknisi", "", "", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Role != "teknisi" {
			t.Fatalf("expected role teknisi, got %s", u.Role)
		}
	})

	t.Run("Sole owner cannot demote self to non-owner", func(t *testing.T) {
		_, uc := setupRepo()
		_, err := uc.UpdateUser(ctx, 1, []string{"owner"}, 1, "", "", "admin", "", "", "")
		if !errors.Is(err, user.ErrLastOwnerDemotion) {
			t.Fatalf("expected ErrLastOwnerDemotion, got %v", err)
		}
	})
}

func TestHierarchy_DeleteAndToggle(t *testing.T) {
	ctx := context.Background()

	setupRepo := func() (*mockUserRepo, *user.ManageUserUseCase) {
		repo := newMockRepo()
		repo.users[1] = &customer.User{ID: 1, Username: "owner1", Role: "owner", IsActive: true}
		repo.users[2] = &customer.User{ID: 2, Username: "admin1", Role: "admin", IsActive: true}
		repo.users[3] = &customer.User{ID: 3, Username: "admin2", Role: "admin", IsActive: true}
		repo.users[4] = &customer.User{ID: 4, Username: "agent1", Role: "agent", IsActive: true}
		return repo, user.NewManageUserUseCase(repo, nil)
	}

	t.Run("Cannot delete or deactivate self", func(t *testing.T) {
		_, uc := setupRepo()
		err := uc.DeleteUser(ctx, 2, []string{"admin"}, 2)
		if !errors.Is(err, user.ErrSelfOperation) {
			t.Fatalf("expected ErrSelfOperation, got %v", err)
		}

		err = uc.ToggleStatus(ctx, 2, []string{"admin"}, 2, false)
		if !errors.Is(err, user.ErrSelfOperation) {
			t.Fatalf("expected ErrSelfOperation, got %v", err)
		}
	})

	t.Run("Owner account cannot be deleted or deactivated", func(t *testing.T) {
		_, uc := setupRepo()
		err := uc.DeleteUser(ctx, 2, []string{"admin"}, 1)
		if !errors.Is(err, user.ErrCannotModifyOwner) {
			t.Fatalf("expected ErrCannotModifyOwner, got %v", err)
		}

		err = uc.ToggleStatus(ctx, 2, []string{"admin"}, 1, false)
		if !errors.Is(err, user.ErrCannotModifyOwner) {
			t.Fatalf("expected ErrCannotModifyOwner, got %v", err)
		}
	})

	t.Run("Admin cannot delete or deactivate another admin", func(t *testing.T) {
		_, uc := setupRepo()
		err := uc.DeleteUser(ctx, 2, []string{"admin"}, 3)
		if !errors.Is(err, user.ErrCannotModifyAdmin) {
			t.Fatalf("expected ErrCannotModifyAdmin, got %v", err)
		}

		err = uc.ToggleStatus(ctx, 2, []string{"admin"}, 3, false)
		if !errors.Is(err, user.ErrCannotModifyAdmin) {
			t.Fatalf("expected ErrCannotModifyAdmin, got %v", err)
		}
	})

	t.Run("Owner can delete or deactivate admin and agent", func(t *testing.T) {
		_, uc := setupRepo()
		err := uc.ToggleStatus(ctx, 1, []string{"owner"}, 3, false)
		if err != nil {
			t.Fatalf("unexpected error owner toggling admin: %v", err)
		}

		err = uc.DeleteUser(ctx, 1, []string{"owner"}, 4)
		if err != nil {
			t.Fatalf("unexpected error owner deleting agent: %v", err)
		}
	})
}
