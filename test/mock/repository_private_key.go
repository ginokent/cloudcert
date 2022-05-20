package mock

import (
	"context"

	"github.com/ginokent/cloudacme/repository"
)

var _ repository.VaultRepository = (*VaultRepository)(nil)

type VaultRepository struct {
	GetVaultIfExistsBool                bool
	GetVaultIfExistsVaultResource       string
	GetVaultIfExistsErr                 error
	CreateVaultIfNotExistsErr           error
	LockVaultErr                        error
	UnlockVaultErr                      error
	GetVaultVersionIfExistsBool         bool
	GetVaultVersionIfExistsVersion      string
	GetVaultVersionIfExistsErr          error
	GetVaultVersionDataIfExistsBool     bool
	GetVaultVersionDataIfExistsResource string
	GetVaultVersionDataIfExistsData     []byte
	GetVaultVersionDataIfExistsErr      error
	AddVaultVersionResource             string
	AddVaultVersionErr                  error
}

func (m *VaultRepository) GetVaultIfExists(ctx context.Context, targetVaultResource string) (exists bool, vaultResource string, err error) {
	return m.GetVaultIfExistsBool, m.GetVaultIfExistsVaultResource, m.GetVaultIfExistsErr
}

func (m *VaultRepository) CreateVaultIfNotExists(ctx context.Context, vaultResource string) (err error) {
	return m.CreateVaultIfNotExistsErr
}

func (m *VaultRepository) LockVault(ctx context.Context, targetVaultResource string) (err error) {
	return m.LockVaultErr
}

func (m *VaultRepository) UnlockVault(ctx context.Context, targetVaultResource string) (err error) {
	return m.UnlockVaultErr
}

func (m *VaultRepository) GetVaultVersionIfExists(ctx context.Context, vaultVersionResource string) (exists bool, version string, err error) {
	return m.GetVaultVersionIfExistsBool, m.GetVaultVersionIfExistsVersion, m.GetVaultVersionIfExistsErr
}

func (m *VaultRepository) GetVaultVersionDataIfExists(ctx context.Context, targetVaultVersionResource string) (exists bool, vaultVersionResource string, data []byte, err error) {
	return m.GetVaultVersionDataIfExistsBool, m.GetVaultVersionDataIfExistsResource, m.GetVaultVersionDataIfExistsData, m.GetVaultVersionDataIfExistsErr
}

func (m *VaultRepository) AddVaultVersion(ctx context.Context, vaultResource string, data []byte) (vaultVersionResource string, err error) {
	return m.AddVaultVersionResource, m.AddVaultVersionErr
}
