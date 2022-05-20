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
		acmeAccountKeyVaultResource string,
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

func (uc *certificatesUseCase) lock(ctx context.Context, resources ...string) (unlock func(), err error) {
	l := contexts.GetLogger(ctx)

	var deferFuncs []func()
	deferFunc := func() {
		for _, f := range deferFuncs {
			f()
		}
	}
	defer func() {
		if err != nil {
			deferFunc()
		}
	}()

	for _, resource := range resources {
		resource := resource

		if err := uc.vaultRepo.CreateVaultIfNotExists(ctx, resource); err != nil {
			return func() {}, errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.CreateVaultIfNotExists: %s: %w", resource, err)
		}

		if err := uc.vaultRepo.LockVault(ctx, resource); err != nil {
			return func() {}, errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.LockVault: %s: %w", resource, err)
		}

		deferFuncs = append(deferFuncs, func() {
			if err := uc.vaultRepo.UnlockVault(ctx, resource); err != nil {
				l.E().Error(errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.UnlockVault: %s: %w", resource, err))
			}
		})
	}

	return deferFunc, nil
}

func (uc *certificatesUseCase) IssueCertificate(
	ctx context.Context,
	acmeAccountKeyVaultResource string,
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
		acmeAccountKeyVaultResource,
		privateKeyVaultResource,
		certificateVaultResource,
		renewPrivateKey,
		keyAlgorithm,
		thresholdOfDaysToExpire,
		domains,
		certcrypto.ParsePEMPrivateKey,
		tls.X509KeyPair,
		nits.X509.CheckCertificatePEM,
		nits.Crypto.GenerateKey,
	)
}

func (uc *certificatesUseCase) _issueCertificate(
	ctx context.Context,
	acmeAccountKeyVaultResource string,
	privateKeyVaultResource string,
	certificateVaultResource string,
	renewPrivateKey bool,
	keyAlgorithm string,
	thresholdOfDaysToExpire int64,
	domains []string,
	_certcrypto_ParsePEMPrivateKey func(key []byte) (crypto.PrivateKey, error),
	_tls_X509KeyPair func(certPEMBlock []byte, keyPEMBlock []byte) (tls.Certificate, error), // nolint: revive,stylecheck
	_nits_X509_CheckCertificatePEM func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error), // nolint: revive,stylecheck
	_nits_Crypto_GenerateKey func(algorithm string) (crypto.PrivateKey, error), // nolint: revive,stylecheck
) (
	privateKeyVaultVersionResource string,
	certificateVaultVersionResource string,
	err error,
) {
	l := contexts.GetLogger(ctx)

	unlock, err := uc.lock(ctx, acmeAccountKeyVaultResource, privateKeyVaultResource, certificateVaultResource)
	if err != nil {
		return "", "", errors.Errorf("(*usecase.certificatesUseCase).lock: %w", err)
	}
	defer unlock()

	acmeAccountKey, err := uc.getAcmeAccountKey(ctx, acmeAccountKeyVaultResource, certcrypto.ParsePEMPrivateKey, nits.Crypto.GenerateKey)
	if err != nil {
		return "", "", errors.Errorf("(*usecase.certificatesUseCase).getAcmeAccountKey: %w", err)
	}

	privateKeyExists, privateKeyVaultVersionResource, privateKeyPEM, privateKeyErr := uc.vaultRepo.GetVaultVersionDataIfExists(ctx, privateKeyVaultResource+"/versions/latest")
	certificateExists, certificateVaultVersionResource, certificatePEM, certificateErr := uc.vaultRepo.GetVaultVersionDataIfExists(ctx, certificateVaultResource+"/versions/latest")
	if privateKeyErr != nil || certificateErr != nil {
		return "", "", errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.GetVaultVersionDataIfExists: privateKeyErr=%v, certificateErr=%v", privateKeyErr, certificateErr)
	}

	var privateKey crypto.PrivateKey

	// NOTE: If renewPrivateKey flag is false, and private key exists, and certificate exists, checking key pair and certificate.
	if !renewPrivateKey && privateKeyExists && certificateExists {
		var keyPairIsBroken bool

		if err := trace.StartFunc(ctx, "certcrypto.ParsePEMPrivateKey")(func(child context.Context) (err error) {
			privateKey, err = _certcrypto_ParsePEMPrivateKey(privateKeyPEM)
			return
		}); err != nil {
			l.E().Error(errors.Errorf("🚨 private key is broken: certcrypto.ParsePEMPrivateKey: %v", err))
			renewPrivateKey = true
			keyPairIsBroken = true
		}

		if err := trace.StartFunc(ctx, "tls.X509KeyPair")(func(child context.Context) (err error) {
			_, err = _tls_X509KeyPair(certificatePEM, privateKeyPEM)
			return
		}); err != nil && !keyPairIsBroken {
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
				// NOTE: If certificate is not expired, early return.
				l.F().Infof("✅ there is still time (%d days) for current certificate to expire. It will not be renewed", daysToExpire)
				return privateKeyVaultVersionResource, certificateVaultVersionResource, nil
			}

			l.F().Infof("❗️ current certificate has expired or is due to expire in less than %d days. Renew the certificate", thresholdOfDaysToExpire)
		}
	}

	// NOTE: If renewPrivateKey flag is true, or private key does not exist, generating private key.
	if renewPrivateKey || !privateKeyExists {
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

		if err := uc.vaultRepo.CreateVaultIfNotExists(ctx, privateKeyVaultResource); err != nil {
			return "", "", errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.CreateVaultIfNotExists: %w", err)
		}

		privateKeyVaultVersionResource, err = uc.vaultRepo.AddVaultVersion(ctx, privateKeyVaultResource, certcrypto.PEMEncode(privateKey))
		if err != nil {
			return "", "", errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.AddVaultVersion: %w", err)
		}

		l.F().Infof("🔐 generated %s private key", keyAlgorithm)
	}

	l.Info("🪪 generate certificate...")

	privateKeyPEM, certificatePEM, _, _, err = uc.letsencryptRepo.IssueCertificate(ctx, acmeAccountKey, privateKey, domains)
	if err != nil {
		return "", "", errors.Errorf("(*usecase.certificatesUseCase).letsencryptRepo.IssueCertificate: %w", err)
	}

	l.Info("🪪 generated certificate")

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

func (uc *certificatesUseCase) getAcmeAccountKey(ctx context.Context, acmeAccountKeyVaultResource string, _certcrypto_ParsePEMPrivateKey func(key []byte) (crypto.PrivateKey, error), _nits_Crypto_GenerateKey func(algorithm string) (crypto.PrivateKey, error)) (acmeAccountKey crypto.PrivateKey, err error) {
	ctx, span := trace.Start(ctx, "(*usecase.certificatesUseCase).getAcmeAccountKey")
	defer span.End()

	l := contexts.GetLogger(ctx)

	acmeAccountKeyExists, _, acmeAccountKeyPEM, err := uc.vaultRepo.GetVaultVersionDataIfExists(ctx, acmeAccountKeyVaultResource+"/versions/latest")
	if err != nil {
		return nil, errors.Errorf("uc.vaultRepo.GetVaultVersionDataIfExists: %w", err)
	}

	if acmeAccountKeyExists {
		acmeAccountKey, err = _certcrypto_ParsePEMPrivateKey(acmeAccountKeyPEM)
		if err != nil {
			l.E().Error(errors.Errorf("🚨 private key is broken: certcrypto.ParsePEMPrivateKey: %v", err))
			acmeAccountKeyExists = false
		}
	}

	if !acmeAccountKeyExists {
		if err := trace.StartFunc(ctx, "nits.Crypto.GenerateKey")(func(child context.Context) (err error) {
			acmeAccountKey, err = _nits_Crypto_GenerateKey(nits.CryptoRSA4096)
			return
		}); err != nil {
			return nil, errors.Errorf("nits.Crypto.GenerateKey: %w", err)
		}
	}

	if _, err = uc.vaultRepo.AddVaultVersion(ctx, acmeAccountKeyVaultResource, certcrypto.PEMEncode(acmeAccountKey)); err != nil {
		return nil, errors.Errorf("(*usecase.certificatesUseCase).vaultRepo.AddVaultVersion: %w", err)
	}

	return acmeAccountKey, nil
}
