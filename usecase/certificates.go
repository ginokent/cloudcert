package usecase

import (
	"context"

	"github.com/newtstat/cloudacme/contexts"
	"github.com/newtstat/cloudacme/repository"
	"github.com/newtstat/cloudacme/trace"
	"github.com/newtstat/nits.go"
	"golang.org/x/xerrors"
)

type CertificatesUseCase interface {
	IssueCertificate(ctx context.Context, privateKeyRenewed bool, privateKeyVaultResource string, certificateVaultResource string, thresholdOfDaysToExpire int64, domains []string) (privateKeyVaultVersionResource, certificateVaultVersionResource string, err error)
}

var _ CertificatesUseCase = (*certificatesUseCase)(nil)

type certificatesUseCase struct {
	vaultRepo       repository.VaultRepository
	letsencryptRepo repository.LetsEncryptRepository
}

// nolint: ireturn
func NewCertificatesUseCase(certificatesRepo repository.VaultRepository, letsencryptRepo repository.LetsEncryptRepository) CertificatesUseCase {
	return CertificatesUseCase(&certificatesUseCase{
		vaultRepo:       certificatesRepo,
		letsencryptRepo: letsencryptRepo,
	})
}

func (uc *certificatesUseCase) IssueCertificate(
	ctx context.Context,
	privateKeyRenewed bool,
	privateKeyVaultResource string,
	certificateVaultResource string,
	thresholdOfDaysToExpire int64,
	domains []string,
) (
	privateKeyVaultVersionResource,
	certificateVaultVersionResource string,
	err error,
) {
	return uc.issueCertificate(ctx,
		privateKeyRenewed,
		privateKeyVaultResource,
		certificateVaultResource,
		thresholdOfDaysToExpire,
		domains,
		nits.X509.CheckCertificatePEM,
	)
}

// nolint: cyclop, funlen
func (uc *certificatesUseCase) issueCertificate(
	ctx context.Context,
	privateKeyRenewed bool,
	privateKeyVaultResource string,
	certificateVaultResource string,
	thresholdOfDaysToExpire int64,
	domains []string,
	checkCertificatePEMFunc func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error),
) (
	privateKeyVaultVersionResource,
	certificateVaultVersionResource string,
	err error,
) {
	ctx, span := trace.Start(ctx, "(*usecase.certificatesUseCase).Issue")
	defer span.End()

	l := contexts.GetLogger(ctx)

	_, privateKeyVaultVersionResource, err = uc.vaultRepo.GetVaultVersionIfExists(ctx, privateKeyVaultResource+"/versions/latest")
	if err != nil {
		return "", "", xerrors.Errorf("(*usecase.certificatesUseCase).vaultRepo.GetVaultVersionIfExists: %w", err)
	}

	var certificateExists bool
	var certificatePEM []byte
	certificateExists, certificateVaultVersionResource, certificatePEM, err = uc.vaultRepo.GetVaultVersionDataIfExists(ctx, certificateVaultResource+"/versions/latest")
	if err != nil {
		return "", "", xerrors.Errorf("(*usecase.certificatesUseCase).vaultRepo.GetVaultVersionDataIfExists: %w", err)
	}

	// nolint: godox
	// TODO: verify private key and certificate

	if !privateKeyRenewed && // NOTE: If renewPrivateKey, skip checking certificate and force to renew certificate
		certificateExists {
		l.Info("🔬 checking certificate...")

		var notyet bool
		var expired bool
		var daysToExpire int64
		if err := trace.StartFunc(ctx, "nits.X509.CheckCertificatePEM")(func(child context.Context) (err error) {
			notyet, _, expired, daysToExpire, err = checkCertificatePEMFunc(certificatePEM)
			return
		}); err != nil {
			l.E().Error(xerrors.Errorf("🚨 certificate (%s) is broken. nits.X509.CheckCertificatePEM: %w", certificateVaultVersionResource, err))
		} else if !notyet && !expired && daysToExpire > thresholdOfDaysToExpire {
			l.F().Infof("✅ there is still time (%d days) for current certificate to expire. It will not be renewed", daysToExpire)
			return privateKeyVaultVersionResource, certificateVaultVersionResource, nil // early return
		}

		l.F().Infof("❗️ current certificate has expired or is due to expire in less than %d days. Renew the certificate", thresholdOfDaysToExpire)
	}

	l.Info("🪪 generate certificate...")

	if err := uc.vaultRepo.CreateVaultIfNotExists(ctx, certificateVaultResource); err != nil {
		return "", "", xerrors.Errorf("(*usecase.certificatesUseCase).vaultRepo.CreateVaultIfNotExists: %w", err)
	}

	privateKeyPEM, certificatePEM, _, _, err := uc.letsencryptRepo.IssueCertificate(ctx, domains)
	if err != nil {
		return "", "", xerrors.Errorf("(*usecase.certificatesUseCase).letsencryptRepo.IssueCertificate: %w", err)
	}

	l.Info("🪪 generated certificate")

	if privateKeyRenewed {
		privateKeyVaultVersionResource, err = uc.vaultRepo.AddVaultVersion(ctx, privateKeyVaultResource, privateKeyPEM)
		if err != nil {
			return "", "", xerrors.Errorf("(*usecase.certificatesUseCase).vaultRepo.AddVaultVersion: %w", err)
		}
	}

	certificateVaultVersionResource, err = uc.vaultRepo.AddVaultVersion(ctx, certificateVaultResource, certificatePEM)
	if err != nil {
		return "", "", xerrors.Errorf("(*usecase.certificatesUseCase).vaultRepo.AddVaultVersion: %w", err)
	}

	l.Info("✅ save certificate")

	return privateKeyVaultVersionResource, certificateVaultVersionResource, nil
}