package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"golang.org/x/crypto/bcrypt"

	devicepb "github.com/quixiq/polyglot/api/gen/v1"
	authadapter "github.com/quixiq/polyglot/internal/adapter/auth"
	iconnect "github.com/quixiq/polyglot/internal/adapter/connect"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

// KnownRoles adalah katalog role sistem — harus sinkron dengan policy seeder
// (internal/adapter/auth/policy_seeder.go). Role di luar katalog ini ditolak.
var KnownRoles = []string{"owner", "admin", "agent", "teknisi"}

// MinPasswordLength adalah panjang minimum password (sama dengan seed).
const MinPasswordLength = 8

// DefaultPageSize dan MaxPageSize membatasi pagination ListUsers.
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// UserConnectHandler implements the UserService ConnectRPC procedures.
// Memakai *postgres.Store langsung (pola sama dengan AuthConnectHandler) dan
// enforcer Casbin untuk sinkronisasi role assignment saat create/update/delete.
type UserConnectHandler struct {
	pgStore    *postgres.Store
	enforcer   *authadapter.CasbinEnforcer
	jwtService *authadapter.JWTService
}

func NewUserConnectHandler(
	pgStore *postgres.Store,
	enforcer *authadapter.CasbinEnforcer,
	jwtService *authadapter.JWTService,
) *UserConnectHandler {
	return &UserConnectHandler{pgStore: pgStore, enforcer: enforcer, jwtService: jwtService}
}

// ListUsers — user:read. Mengembalikan halaman user dengan pencarian
// username/email dan role efektif (multi-role dari Casbin).
func (h *UserConnectHandler) ListUsers(ctx context.Context, req *connect.Request[devicepb.ListUsersRequest]) (*connect.Response[devicepb.ListUsersResponse], error) {
	if h.pgStore == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("database store unavailable"))
	}

	page := int(req.Msg.GetPage())
	pageSize := int(req.Msg.GetPageSize())
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	users, total, err := h.pgStore.ListUsers(page, pageSize, req.Msg.GetSearch())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	items := make([]*devicepb.User, 0, len(users))
	for _, u := range users {
		items = append(items, h.toProtoUser(u))
	}

	return connect.NewResponse(&devicepb.ListUsersResponse{
		Users:    items,
		Total:    total,
		Page:     int32(page),
		PageSize: int32(pageSize),
	}), nil
}

// CreateUser — user:manage. Validasi, bcrypt, simpan, lalu assign role utama
// ke Casbin supaya enforcement langsung berlaku.
func (h *UserConnectHandler) CreateUser(ctx context.Context, req *connect.Request[devicepb.CreateUserRequest]) (*connect.Response[devicepb.CreateUserResponse], error) {
	username := strings.TrimSpace(req.Msg.GetUsername())
	email := strings.TrimSpace(req.Msg.GetEmail())
	password := req.Msg.GetPassword()
	role := strings.TrimSpace(req.Msg.GetRole())

	if err := validateCreateUser(username, email, password, role); err != nil {
		return nil, err
	}
	if h.pgStore == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("database store unavailable"))
	}

	if existing, err := h.pgStore.FindUserByUsername(username); err == nil && existing != nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("username %q already exists", username))
	}
	if existing, err := h.pgStore.FindUserByEmail(email); err == nil && existing != nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("email %q already in use", email))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("hash password: %w", err))
	}

	user := &customer.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		IsActive:     true,
		TenantID:     "tenant-default",
	}
	if err := h.pgStore.CreateUser(user); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("create user: %w", err))
	}

	// Sinkron role ke Casbin — tanpa ini user baru tidak punya permission apa pun.
	if h.enforcer != nil {
		if ok, err := h.enforcer.AddRoleForUser(fmt.Sprintf("%d", user.ID), role); err != nil || !ok {
			// Role tetap tersimpan di kolom users.role (fallback JWT) —
			// kegagalan Casbin bukan alasan membatalkan pembuatan user.
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("assign role in casbin: ok=%v err=%v", ok, err))
		}
	}

	return connect.NewResponse(&devicepb.CreateUserResponse{User: h.toProtoUser(user)}), nil
}

// UpdateUser — user:manage. Full-update username/email/role. Guard: tidak
// bisa mengubah role sendiri, dan owner terakhir tidak bisa di-demote.
func (h *UserConnectHandler) UpdateUser(ctx context.Context, req *connect.Request[devicepb.UpdateUserRequest]) (*connect.Response[devicepb.UpdateUserResponse], error) {
	id := req.Msg.GetId()
	username := strings.TrimSpace(req.Msg.GetUsername())
	email := strings.TrimSpace(req.Msg.GetEmail())
	role := strings.TrimSpace(req.Msg.GetRole())

	if id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user id is required"))
	}
	if username == "" || email == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("username and email are required"))
	}
	if !isKnownRole(role) {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown role %q (known: %s)", role, strings.Join(KnownRoles, ", ")))
	}
	if h.pgStore == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("database store unavailable"))
	}

	existing, err := h.pgStore.FindUserByID(uint(id))
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	callerID, err := h.callerID(req.Header())
	if err != nil {
		return nil, err
	}
	// Tidak bisa mengubah role sendiri — mencegah self-demote untuk melepas
	// tanggung jawab owner/admin lalu kehilangan akses permanen.
	if callerID == uint(id) && existing.Role != role {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot change your own role"))
	}
	// Owner terakhir tidak boleh di-demote.
	if existing.Role == "owner" && role != "owner" {
		owners, err := h.countOwners()
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		if owners <= 1 {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot demote the last owner"))
		}
	}

	if username != existing.Username {
		if existing2, err := h.pgStore.FindUserByUsername(username); err == nil && existing2 != nil && existing2.ID != existing.ID {
			return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("username %q already exists", username))
		}
	}
	if email != existing.Email {
		if existing2, err := h.pgStore.FindUserByEmail(email); err == nil && existing2 != nil && existing2.ID != existing.ID {
			return nil, connect.NewError(connect.CodeAlreadyExists, fmt.Errorf("email %q already in use", email))
		}
	}

	oldRole := existing.Role
	existing.Username = username
	existing.Email = email
	existing.Role = role
	if err := h.pgStore.UpdateUser(existing); err != nil {
		return nil, mapUserError(err)
	}

	// Sinkron perubahan role ke Casbin: hapus assignment lama, assign baru.
	if h.enforcer != nil && oldRole != role {
		if _, err := h.enforcer.DeleteRoleForUser(fmt.Sprintf("%d", existing.ID), oldRole); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("sync role in casbin: %w", err))
		}
		if ok, err := h.enforcer.AddRoleForUser(fmt.Sprintf("%d", existing.ID), role); err != nil || !ok {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("sync role in casbin: ok=%v err=%v", ok, err))
		}
	}

	return connect.NewResponse(&devicepb.UpdateUserResponse{User: h.toProtoUser(existing)}), nil
}

// ResetPassword — user:manage. Set password baru (bcrypt) tanpa menonaktifkan
// sesi refresh token yang sedang aktif.
func (h *UserConnectHandler) ResetPassword(ctx context.Context, req *connect.Request[devicepb.ResetPasswordRequest]) (*connect.Response[devicepb.ResetPasswordResponse], error) {
	id := req.Msg.GetId()
	password := req.Msg.GetNewPassword()
	if id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user id is required"))
	}
	if len(password) < MinPasswordLength {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("password must be at least %d characters", MinPasswordLength))
	}
	if h.pgStore == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("database store unavailable"))
	}

	if _, err := h.pgStore.FindUserByID(uint(id)); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("hash password: %w", err))
	}
	if err := h.pgStore.SetPasswordHash(uint(id), string(hash)); err != nil {
		return nil, mapUserError(err)
	}

	return connect.NewResponse(&devicepb.ResetPasswordResponse{Success: true}), nil
}

// ToggleActive — user:manage. Membalik is_active. Guard: tidak bisa
// menonaktifkan diri sendiri, dan owner terakhir tidak bisa dinonaktifkan.
func (h *UserConnectHandler) ToggleActive(ctx context.Context, req *connect.Request[devicepb.ToggleActiveRequest]) (*connect.Response[devicepb.ToggleActiveResponse], error) {
	id := req.Msg.GetId()
	if id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user id is required"))
	}
	if h.pgStore == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("database store unavailable"))
	}

	existing, err := h.pgStore.FindUserByID(uint(id))
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	callerID, err := h.callerID(req.Header())
	if err != nil {
		return nil, err
	}
	if callerID == uint(id) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot deactivate your own account"))
	}
	if existing.Role == "owner" && existing.IsActive {
		owners, err := h.countOwners()
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		if owners <= 1 {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot deactivate the last owner"))
		}
	}

	if err := h.pgStore.SetUserActive(uint(id), !existing.IsActive); err != nil {
		return nil, mapUserError(err)
	}
	existing.IsActive = !existing.IsActive

	return connect.NewResponse(&devicepb.ToggleActiveResponse{User: h.toProtoUser(existing)}), nil
}

// DeleteUser — user:manage. Guard: tidak bisa menghapus diri sendiri, dan
// owner terakhir tidak bisa dihapus. Role Casbin ikut dibersihkan.
func (h *UserConnectHandler) DeleteUser(ctx context.Context, req *connect.Request[devicepb.DeleteUserRequest]) (*connect.Response[devicepb.DeleteUserResponse], error) {
	id := req.Msg.GetId()
	if id == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("user id is required"))
	}
	if h.pgStore == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("database store unavailable"))
	}

	existing, err := h.pgStore.FindUserByID(uint(id))
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, errors.New("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	callerID, err := h.callerID(req.Header())
	if err != nil {
		return nil, err
	}
	if callerID == uint(id) {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot delete your own account"))
	}
	if existing.Role == "owner" {
		owners, err := h.countOwners()
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		if owners <= 1 {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("cannot delete the last owner"))
		}
	}

	if err := h.pgStore.DeleteUser(uint(id)); err != nil {
		return nil, mapUserError(err)
	}
	if h.enforcer != nil {
		if _, err := h.enforcer.DeleteRolesForUser(fmt.Sprintf("%d", id)); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("clean up casbin roles: %w", err))
		}
	}

	return connect.NewResponse(&devicepb.DeleteUserResponse{Success: true}), nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────

func (h *UserConnectHandler) toProtoUser(u *customer.User) *devicepb.User {
	if u == nil {
		return &devicepb.User{}
	}
	roles := []string{u.Role}
	if h.enforcer != nil {
		if r, err := h.enforcer.GetRolesForUser(fmt.Sprintf("%d", u.ID)); err == nil && len(r) > 0 {
			roles = r
		}
	}
	return &devicepb.User{
		Id:            uint64(u.ID),
		Username:      u.Username,
		Email:         u.Email,
		Role:          u.Role,
		Roles:         roles,
		IsActive:      u.IsActive,
		TenantId:      u.TenantID,
		CreatedAtUnix: u.CreatedAt.Unix(),
		UpdatedAtUnix: u.UpdatedAt.Unix(),
	}
}

// callerID mengekstrak user ID dari token Bearer di header Authorization —
// pola yang sama dengan GetMe. Dipakai guard self-action (delete/deactivate/
// change-own-role).
func (h *UserConnectHandler) callerID(hdr http.Header) (uint, error) {
	if h.jwtService == nil {
		return 0, connect.NewError(connect.CodeInternal, errors.New("jwt service unavailable"))
	}
	authHeader := hdr.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return 0, connect.NewError(connect.CodeUnauthenticated, errors.New("missing bearer token"))
	}
	claims, err := h.jwtService.ValidateToken(parts[1])
	if err != nil {
		return 0, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token: %w", err))
	}
	return claims.UserID, nil
}

// countOwners menghitung user dengan role owner — baik sebagai role utama
// (kolom users.role) maupun role tambahan via Casbin.
func (h *UserConnectHandler) countOwners() (int, error) {
	users, err := h.pgStore.FindAllUsers()
	if err != nil {
		return 0, err
	}
	n := 0
	for i := range users {
		if users[i].Role == "owner" {
			n++
			continue
		}
		if h.enforcer != nil {
			if roles, err := h.enforcer.GetRolesForUser(fmt.Sprintf("%d", users[i].ID)); err == nil {
				for _, r := range roles {
					if r == "owner" {
						n++
						break
					}
				}
			}
		}
	}
	return n, nil
}

func validateCreateUser(username, email, password, role string) error {
	if username == "" {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("username is required"))
	}
	if email == "" || !strings.Contains(email, "@") {
		return connect.NewError(connect.CodeInvalidArgument, errors.New("a valid email is required"))
	}
	if len(password) < MinPasswordLength {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("password must be at least %d characters", MinPasswordLength))
	}
	if !isKnownRole(role) {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown role %q (known: %s)", role, strings.Join(KnownRoles, ", ")))
	}
	return nil
}

func isKnownRole(role string) bool {
	for _, r := range KnownRoles {
		if r == role {
			return true
		}
	}
	return false
}

func mapUserError(err error) error {
	switch {
	case errors.Is(err, postgres.ErrNotFound):
		return connect.NewError(connect.CodeNotFound, errors.New("user not found"))
	case errors.Is(err, postgres.ErrInvalidArgument):
		return connect.NewError(connect.CodeInvalidArgument, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}

// NewUserServiceHandler exposes the UserService ConnectRPC procedures.
func NewUserServiceHandler(
	pgStore *postgres.Store,
	enforcer *authadapter.CasbinEnforcer,
	jwtService *authadapter.JWTService,
) (string, http.Handler) {
	handler := NewUserConnectHandler(pgStore, enforcer, jwtService)
	mux := http.NewServeMux()
	codecOpt := connect.WithCodec(iconnect.JSONCodec())

	serviceName := "polyglot.v1.UserService"
	mux.Handle("/"+serviceName+"/ListUsers", connect.NewUnaryHandler("/"+serviceName+"/ListUsers", handler.ListUsers, codecOpt))
	mux.Handle("/"+serviceName+"/CreateUser", connect.NewUnaryHandler("/"+serviceName+"/CreateUser", handler.CreateUser, codecOpt))
	mux.Handle("/"+serviceName+"/UpdateUser", connect.NewUnaryHandler("/"+serviceName+"/UpdateUser", handler.UpdateUser, codecOpt))
	mux.Handle("/"+serviceName+"/ResetPassword", connect.NewUnaryHandler("/"+serviceName+"/ResetPassword", handler.ResetPassword, codecOpt))
	mux.Handle("/"+serviceName+"/ToggleActive", connect.NewUnaryHandler("/"+serviceName+"/ToggleActive", handler.ToggleActive, codecOpt))
	mux.Handle("/"+serviceName+"/DeleteUser", connect.NewUnaryHandler("/"+serviceName+"/DeleteUser", handler.DeleteUser, codecOpt))

	return "/" + serviceName + "/", mux
}
