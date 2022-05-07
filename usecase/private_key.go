package usecase

import (
	"context"
	"crypto"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/newtstat/cloudacme/contexts"
	"github.com/newtstat/cloudacme/repository"
	"github.com/newtstat/cloudacme/trace"
	"github.com/newtstat/nits.go"
	"golang.org/x/xerrors"
)

type PrivateKeyUseCase interface {
	GetPrivateKey(ctx context.Context, privateKeyVaultResource string, renewPrivateKey bool, keyAlgorithm string) (renewd bool, privateKey crypto.PrivateKey, err error)
	Lock(ctx context.Context, privateKeyVaultResource string, certificateVaultResource string) (err error)
	Unlock(ctx context.Context, privateKeyVaultResource string, certificateVaultResource string) (err error)
}

var _ PrivateKeyUseCase = (*privateKeyUseCase)(nil)

type privateKeyUseCase struct {
	vaultRepo repository.VaultRepository
}

// nolint: ireturn
func NewPrivateKeyUsecase(vaultRepo repository.VaultRepository) PrivateKeyUseCase {
	return PrivateKeyUseCase(&privateKeyUseCase{
		vaultRepo: vaultRepo,
	})
}

func (uc *privateKeyUseCase) GetPrivateKey(
	ctx context.Context,
	privateKeyVaultResource string,
	renewPrivateKey bool,
	keyAlgorithm string,
) (
	renewed bool,
	privateKey crypto.PrivateKey,
	err error,
) {
	return uc.getPrivateKey(
		ctx,
		privateKeyVaultResource,
		renewPrivateKey,
		keyAlgorithm,
		nits.X509.ParsePKCSXPrivateKeyPEM,
		nits.Crypto.GenerateKey,
	)
}

func (uc *privateKeyUseCase) getPrivateKey(
	ctx context.Context,
	privateKeyVaultResource string,
	renewPrivateKey bool,
	keyAlgorithm string,
	parsePKCSXPrivateKeyPEMFunc func(pemData []byte) (crypto.PrivateKey, error),
	generateKeyFunc func(algorithm string) (crypto.PrivateKey, error),
) (
	renewed bool,
	privateKey crypto.PrivateKey,
	err error,
) {
	ctx, span := trace.Start(ctx, "(*usecase.privateKeyUseCase).GetPrivateKey")
	defer span.End()

	l := contexts.GetLogger(ctx)

	var privateKeyExists bool
	var privateKeyPEM []byte
	privateKeyExists, _, privateKeyPEM, err = uc.vaultRepo.GetVaultVersionDataIfExists(ctx, privateKeyVaultResource+"/versions/latest")
	if err != nil {
		return false, nil, xerrors.Errorf("(*usecase.privateKeyUseCase).vaultRepo.GetVaultVersionDataIfExists: %w", err)
	}

	// l.F().Debugf("usecase: uc.vaultRepo.GetVaultVersionDataIfExists: %s", string(privateKeyPEM))

	if privateKeyExists && !renewPrivateKey {
		privateKey, err = certcrypto.ParsePEMPrivateKey(privateKeyPEM)
		if err == nil {
			return false, privateKey, nil
		}

		l.E().Error(xerrors.Errorf("🚨 private key is broken: certcrypto.ParsePEMPrivateKey: %v", err))
	}

	// NOTE: if !privateKeyExists OR renewPrivateKey, always generate private key
	if keyAlgorithm == "" {
		keyAlgorithm = nits.CryptoRSA4096
	}

	l.F().Infof("🔐 generate %s private key...", keyAlgorithm)

	if err := trace.StartFunc(ctx, "nits.Crypto.GenerateKey")(func(child context.Context) (err error) {
		privateKey, err = generateKeyFunc(keyAlgorithm)
		return
	}); err != nil {
		return false, nil, xerrors.Errorf("nits.Crypto.GenerateKey: %w", err)
	}

	if err := uc.vaultRepo.CreateVaultIfNotExists(ctx, privateKeyVaultResource); err != nil {
		return false, nil, xerrors.Errorf("(*usecase.privateKeyUseCase).vaultRepo.CreateVaultIfNotExists: %w", err)
	}

	l.F().Infof("🔐 generated %s private key", keyAlgorithm)

	return true, privateKey, nil
}

func (uc *privateKeyUseCase) Lock(ctx context.Context, privateKeyVaultResource, certificateVaultResource string) (err error) {
	ctx, span := trace.Start(ctx, "(*usecase.privateKeyUseCase).Lock")
	defer span.End()

	privateKeyErr := uc.vaultRepo.LockVault(ctx, privateKeyVaultResource)
	certificateErr := uc.vaultRepo.LockVault(ctx, certificateVaultResource)

	if privateKeyErr != nil || certificateErr != nil {
		return xerrors.Errorf("(*usecase.privateKeyUseCase).vaultRepo.LockVault: %v: privateKey=%v certificate=%v", repository.ErrFailedToAcquireLock, privateKeyErr, certificateErr)
	}

	return nil
}

func (uc *privateKeyUseCase) Unlock(ctx context.Context, privateKeyVaultResource, certificateVaultResource string) (err error) {
	ctx, span := trace.Start(ctx, "(*usecase.privateKeyUseCase).Unlock")
	defer span.End()

	privateKeyErr := uc.vaultRepo.UnlockVault(ctx, privateKeyVaultResource)
	certificateErr := uc.vaultRepo.UnlockVault(ctx, certificateVaultResource)

	if privateKeyErr != nil || certificateErr != nil {
		return xerrors.Errorf("(*usecase.privateKeyUseCase).vaultRepo.LockVault: privateKey=%v certificate=%v", privateKeyErr, certificateErr)
	}

	return nil
}
