package importer_test

import (
	"context"
	"errors"

	domainDevice "github.com/quixiq/polyglot/internal/domain/device"
)

// e2eVault implements port.CredentialVault minimal untuk E2E import
// (password disimpan sebagai ciphertext base64-palsu "enc:").
type e2eVault struct{}

func (e2eVault) EncryptString(_ context.Context, p string) (string, error) {
	return "enc:" + p, nil
}

func (e2eVault) DecryptString(_ context.Context, c string) (string, error) {
	if len(c) > 4 && c[:4] == "enc:" {
		return c[4:], nil
	}
	return "", errors.New("not encrypted")
}

func (e2eVault) Get(_ context.Context, _ string) (domainDevice.Credentials, error) {
	return domainDevice.Credentials{}, domainDevice.ErrNotFound
}

func (e2eVault) Save(_ context.Context, _ string, _ domainDevice.Credentials) error {
	return nil
}
