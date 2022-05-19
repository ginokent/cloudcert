package usecase

import (
	"context"
	"crypto"
	"crypto/tls"

	"github.com/cockroachdb/errors"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/newtstat/cloudacme/contexts"
	"github.com/newtstat/cloudacme/repository"
	"github.com/newtstat/cloudacme/trace"
	"github.com/newtstat/nits.go"
)

type CertificatesUseCase interface {
	IssueCertificate(
		ctx context.Context,
		privateKeyVaultResource string,
		certificateVaultResource string,
		renewPrivateKey bool,
		keyAlgorithm string,
		thresholdOfDaysToExpire int64,
		domains []string,
	) (
		privateKeyVaultVersionResource string,
		certificateVaultVersionResource string,
		err error,
	)
}

var _ CertificatesUseCase = (*certificatesUseCase)(nil)

type certificatesUseCase struct {
	vaultRepo       repository.VaultRepository
	letsencryptRepo repository.LetsEncryptRepository
}

func NewCertificatesUseCase(certificatesRepo repository.VaultRepository, letsencryptRepo repository.LetsEncryptRepository) CertificatesUseCase {
	return CertificatesUseCase(&certificatesUseCase{
		vaultRepo:       certificatesRepo,
		letsencryptRepo: letsencryptRepo,
	})
}

func (uc *certificatesUseCase) IssueCertificate(
	ctx context.Context,
	privateKeyVaultResource string,
	certificateVaultResource string,
	renewPrivateKey bool,
	keyAlgorithm string,
	thresholdOfDaysToExpire int64,
	domains []string,
) (
	privateKeyVaultVersionResource string,
	certificateVaultVersionResource string,
	err error,
) {
	ctx, span := trace.Start(ctx, "(*usecase.certificatesUseCase).IssueCertificate")
	defer span.End()

	return uc._issueCertificate(
		ctx,
		privateKeyVaultResource,
		certificateVaultResource,
		renewPrivateKey,
		keyAlgorithm,
		thresholdOfDaysToExpire,
		domains,
		certcrypto.PEMEncode,
		tls.X509KeyPair,
		nits.X509.CheckCertificatePEM,
		nits.Crypto.GenerateKey,
	)
}

func (uc *certificatesUseCase) _issueCertificate(
	ctx context.Context,
	privateKeyVaultResource string,
	certificateVaultResource string,
	renewPrivateKey bool,
	keyAlgorithm string,
	thresholdOfDaysToExpire int64,
	domains []string,
	_certcrypto_PEMEncode func(data interface{}) []byte, // nolint: revive,stylecheck
	_tls_X509KeyPair func(certPEMBlock []byte, keyPEMBlock []byte) (tls.Certificate, error), // nolint: revive,stylecheck
	_nits_X509_CheckCertificatePEM func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error), // nolint: revive,stylecheck
	_nits_Crypto_GenerateKey func(algorithm string) (crypto.PrivateKey, error), // nolint: revive,stylecheck
) (
	privateKeyVaultVersionResource string,
	certificateVaultVersionResource string,
	err error,
) {
	l := contexts.GetLogger(ctx)

	privateKeyExists, privateKeyVaultVersionResource, privateKeyPEM, privateKeyErr := uc.vaultRepo.GetVaultVersionDataIfExists(ctx, privateKeyVaultResource+"/versions/latest")
	certificateExists, certificateVaultVersionResource, certificatePEM, certificateErr := uc.vaultRepo.GetVaultVersionDataIfExists(ctx, certificateVaultResource+"/versions/latest")
	if privateKeyErr != nil || certificateErr != nil {
		return "", "", errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.GetVaultVersionDataIfExists: privateKeyErr=%v, certificateErr=%w", privateKeyErr, certificateErr)
	}
	l.F().Debugf("Vault から取得した秘密鍵: %s", string(privateKeyPEM))
	l.F().Debugf("Vault から取得した証明書: %s", string(certificatePEM))

	var privateKey crypto.PrivateKey

	if !renewPrivateKey && privateKeyExists && certificateExists {
		var keyPairIsBroken bool

		privateKey, err = certcrypto.ParsePEMPrivateKey(privateKeyPEM)
		if err != nil {
			l.E().Error(errors.Errorf("🚨 private key is broken: certcrypto.ParsePEMPrivateKey: %v", err))
			renewPrivateKey = true
			keyPairIsBroken = true
		}

		if _, err := _tls_X509KeyPair(certificatePEM, privateKeyPEM); !keyPairIsBroken && err != nil {
			l.E().Error(errors.Errorf("🚨 a pair of certificate and private key is broken. tls.X509KeyPair: %w", err))
			keyPairIsBroken = true
		}

		if !keyPairIsBroken {
			l.Info("🔬 checking certificate...")

			var notyet bool
			var expired bool
			var daysToExpire int64
			if err := trace.StartFunc(ctx, "nits.X509.CheckCertificatePEM")(func(child context.Context) (err error) {
				notyet, _, expired, daysToExpire, err = _nits_X509_CheckCertificatePEM(certificatePEM)
				return
			}); err != nil {
				l.E().Error(errors.Errorf("🚨 certificate (%s) is broken. nits.X509.CheckCertificatePEM: %w", certificateVaultVersionResource, err))
			} else if !notyet && !expired && daysToExpire > thresholdOfDaysToExpire {
				l.F().Infof("✅ there is still time (%d days) for current certificate to expire. It will not be renewed", daysToExpire)
				return privateKeyVaultVersionResource, certificateVaultVersionResource, nil // early return
			}

			l.F().Infof("❗️ current certificate has expired or is due to expire in less than %d days. Renew the certificate", thresholdOfDaysToExpire)
		}
	}

	if renewPrivateKey {
		if keyAlgorithm == "" {
			keyAlgorithm = nits.CryptoRSA4096
		}

		l.F().Infof("🔐 generate %s private key...", keyAlgorithm)

		if err := trace.StartFunc(ctx, "nits.Crypto.GenerateKey")(func(child context.Context) (err error) {
			privateKey, err = _nits_Crypto_GenerateKey(keyAlgorithm)
			return
		}); err != nil {
			return "", "", errors.Errorf("nits.Crypto.GenerateKey: %w", err)
		}

		l.F().Infof("🔐 generated %s private key", keyAlgorithm)
	}

	l.Info("🪪 generate certificate...")

	privateKeyPEM, certificatePEM, _, _, err = uc.letsencryptRepo.IssueCertificate(ctx, privateKey, domains)
	if err != nil {
		return "", "", errors.Errorf("(*usecase.certificatesUseCase).letsencryptRepo.IssueCertificate: %w", err)
	}
	l.F().Debugf("Let's Encrypt から取得した秘密鍵: %s", string(privateKeyPEM))
	l.F().Debugf("Let's Encrypt から取得した証明書: %s", string(certificatePEM))

	l.Info("🪪 generated certificate")

	if renewPrivateKey {
		if err := uc.vaultRepo.CreateVaultIfNotExists(ctx, privateKeyVaultResource); err != nil {
			return "", "", errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.CreateVaultIfNotExists: %w", err)
		}

		privateKeyVaultVersionResource, err = uc.vaultRepo.AddVaultVersion(ctx, privateKeyVaultResource, privateKeyPEM)
		if err != nil {
			return "", "", errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.AddVaultVersion: %w", err)
		}
	}

	if err := uc.vaultRepo.CreateVaultIfNotExists(ctx, certificateVaultResource); err != nil {
		return "", "", errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.CreateVaultIfNotExists: %w", err)
	}

	certificateVaultVersionResource, err = uc.vaultRepo.AddVaultVersion(ctx, certificateVaultResource, certificatePEM)
	if err != nil {
		return "", "", errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.AddVaultVersion: %w", err)
	}

	l.Info("✅ save certificate")

	return privateKeyVaultVersionResource, certificateVaultVersionResource, nil
}
