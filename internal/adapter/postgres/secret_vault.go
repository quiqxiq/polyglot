package postgres

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/port"
)

var _ port.SecretVault = (*SecretVault)(nil)

// SecretVault implements port.SecretVault with AES-256-GCM encrypted rows
// in the secrets table (same scheme as device CredentialVault).
type SecretVault struct {
	db  *gorm.DB
	key string
}

func NewSecretVault(db *gorm.DB, encryptionKey ...string) *SecretVault {
	v := &SecretVault{db: db}
	if len(encryptionKey) > 0 {
		v.key = encryptionKey[0]
	}
	return v
}

func (v *SecretVault) Put(ctx context.Context, key, secret string) error {
	m, err := model.SecretModelFromDomainWithKey(key, secret, v.key)
	if err != nil {
		return err
	}
	return v.db.WithContext(ctx).Save(m).Error
}

func (v *SecretVault) Get(ctx context.Context, key string) (string, error) {
	var m model.SecretModel
	err := v.db.WithContext(ctx).First(&m, "\"key\" = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("secret %q not found", key)
	}
	if err != nil {
		return "", err
	}
	return m.ToSecret(v.key)
}

func (v *SecretVault) Delete(ctx context.Context, key string) error {
	return v.db.WithContext(ctx).Delete(&model.SecretModel{}, "\"key\" = ?", key).Error
}
