package repository

import (
	"context"

	"github.com/cockroachdb/errors"
)

var ErrInvalidVaultResource = errors.New("invalid vault resource")

type VaultRepository interface {
	GetVaultIfExists(ctx context.Context, targetVaultResource string) (exists bool, vaultResource string, err error)
	CreateVaultIfNotExists(ctx context.Context, targetVaultResource string) (err error)
	LockVault(ctx context.Context, targetVaultResource string) (err error)
	UnlockVault(ctx context.Context, targetVaultResource string) (err error)
	GetVaultVersionIfExists(ctx context.Context, targetVaultVersionResource string) (exists bool, vaultVersionResource string, err error)
	GetVaultVersionDataIfExists(ctx context.Context, targetVaultVersionResource string) (exists bool, vaultVersionResource string, data []byte, err error)
	AddVaultVersion(ctx context.Context, vaultResource string, data []byte) (vaultVersionResource string, err error)
}

var _ VaultRepository = (*vaultGoogleSecretManagerRepository)(nil)
