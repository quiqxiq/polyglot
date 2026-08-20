package auth

import (
	"context"
	"sync"
	"testing"
	"time"
)

// fakeRefreshStore is an in-memory port.RefreshTokenStore.
type fakeRefreshStore struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeRefreshStore() *fakeRefreshStore {
	return &fakeRefreshStore{data: map[string]string{}}
}

func (f *fakeRefreshStore) Set(_ context.Context, key, value string, _ int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[key] = value
	return nil
}

func (f *fakeRefreshStore) Get(_ context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.data[key], nil
}

func (f *fakeRefreshStore) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, key)
	return nil
}

func (f *fakeRefreshStore) len() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.data)
}

func TestRefreshTokenService(t *testing.T) {
	store := newFakeRefreshStore()
	svc := NewRefreshTokenService(store, 7*24*time.Hour)
	ctx := context.Background()

	t.Run("Issue and Validate", func(t *testing.T) {
		token, err := svc.Issue(ctx, RefreshClaims{UserID: 1, Roles: []string{"admin"}})
		if err != nil {
			t.Fatalf("Issue: %v", err)
		}
		if token == "" {
			t.Fatal("token is empty")
		}
		claims, err := svc.Validate(ctx, token)
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if claims.UserID != 1 || len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
			t.Fatalf("claims mismatch: %+v", claims)
		}
	})

	t.Run("Validate unknown token fails", func(t *testing.T) {
		if _, err := svc.Validate(ctx, "nonexistent"); err == nil {
			t.Fatal("expected error for unknown token")
		}
	})

	t.Run("Rotate revokes old and issues new", func(t *testing.T) {
		before := store.len()
		token, err := svc.Issue(ctx, RefreshClaims{UserID: 2, Roles: []string{"teknisi"}})
		if err != nil {
			t.Fatalf("Issue: %v", err)
		}
		newToken, claims, err := svc.Rotate(ctx, token)
		if err != nil {
			t.Fatalf("Rotate: %v", err)
		}
		if newToken == token {
			t.Fatal("rotated token must differ from old")
		}
		if claims.UserID != 2 {
			t.Fatalf("claims lost identity: %+v", claims)
		}
		// Old token is single-use — must now fail.
		if _, err := svc.Validate(ctx, token); err == nil {
			t.Fatal("old token still valid after rotation")
		}
		// New token works.
		if _, err := svc.Validate(ctx, newToken); err != nil {
			t.Fatalf("new token invalid: %v", err)
		}
		// Rotation should not leak store entries (old deleted, new added).
		if store.len() != before+1 {
			t.Fatalf("store size = %d, want %d", store.len(), before+1)
		}
	})

	t.Run("Revoke invalidates token", func(t *testing.T) {
		token, err := svc.Issue(ctx, RefreshClaims{UserID: 3})
		if err != nil {
			t.Fatalf("Issue: %v", err)
		}
		if err := svc.Revoke(ctx, token); err != nil {
			t.Fatalf("Revoke: %v", err)
		}
		if _, err := svc.Validate(ctx, token); err == nil {
			t.Fatal("token still valid after revoke")
		}
	})
}
