package controller

import (
	"context"
	"crypto"
	"io"
	"os"

	"github.com/newtstat/cloudacme/contexts"
	"github.com/newtstat/cloudacme/proto/generated/go/v1/cloudacme"
	"github.com/newtstat/cloudacme/repository"
	"github.com/newtstat/cloudacme/trace"
	"github.com/newtstat/cloudacme/usecase"
	"golang.org/x/xerrors"
)

type CertificatesController struct {
	cloudacme.UnimplementedCertificatesServer
}

func (c *CertificatesController) Issue(ctx context.Context, req *cloudacme.IssueCertificateRequest) (resp *cloudacme.IssueCertificateResponse, err error) {
	return c.issue(ctx, req, repository.NewVaultGoogleSecretManagerRepository, repository.NewLetsEncryptGoogleCloudRepository)
}

// nolint: cyclop, funlen
func (*CertificatesController) issue(
	ctx context.Context,
	req *cloudacme.IssueCertificateRequest,
	newVaultGoogleSecretManagerRepository func(ctx context.Context) (repository.VaultRepository, error),
	newLetsEncryptGoogleCloudRepository func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error),
) (
	resp *cloudacme.IssueCertificateResponse,
	err error,
) {
	ctx, span := trace.Start(ctx, "(*controller.CertificatesController).Issue")
	defer span.End()

	l := contexts.GetLogger(ctx)

	var vaultRepo repository.VaultRepository
	switch req.GetVaultProvider() {
	case cloudacme.IssueCertificateRequest_gcloud.String():
		vaultRepo, err = newVaultGoogleSecretManagerRepository(ctx)
		if err != nil {
			return nil, xerrors.Errorf("repository.NewVaultGoogleSecretManagerRepository: %w", err)
		}
	default:
		return nil, xerrors.Errorf("provider=\"%s\": %w", req.GetVaultProvider(), err)
	}

	privateKeyUsecase := usecase.NewPrivateKeyUsecase(vaultRepo)
	if err := privateKeyUsecase.Lock(ctx, req.PrivateKeyVaultResource, req.CertificateVaultResource); err != nil {
		return nil, xerrors.Errorf("(usecase.PrivateKeyUseCase).Lock: %w", err)
	}
	defer func() {
		if err := privateKeyUsecase.Unlock(ctx, req.PrivateKeyVaultResource, req.CertificateVaultResource); err != nil {
			l.E().Error(xerrors.Errorf("(usecase.PrivateKeyUseCase).Lock: %w", err))
		}
	}()
	privateKeyRenewed, privateKey, err := privateKeyUsecase.GetPrivateKey(ctx, req.GetPrivateKeyVaultResource(), req.GetRenewPrivateKey(), req.GetKeyAlgorithm())
	if err != nil {
		return nil, xerrors.Errorf("(usecase.PrivateKeyUseCase).GetPrivateKey: %w", err)
	}

	var letsencryptRepo repository.LetsEncryptRepository
	switch req.GetDnsProvider() {
	case cloudacme.IssueCertificateRequest_gcloud.String():
		letsencryptRepo, err = newLetsEncryptGoogleCloudRepository(ctx, req.GetTermsOfServiceAgreed(), req.GetEmail(), privateKey, req.GetDnsProviderID(), req.GetStaging(), os.Stdout)
		if err != nil {
			return nil, xerrors.Errorf("repository.NewLetsEncryptGoogleCloudRepository: %w", err)
		}
	default:
		return nil, xerrors.Errorf("provider=\"%s\": %w", req.GetVaultProvider(), err)
	}

	certUseCase := usecase.NewCertificatesUseCase(vaultRepo, letsencryptRepo)

	privateKeyVaultVersionResource, certificateVaultVersionResource, err := certUseCase.IssueCertificate(ctx, privateKeyRenewed, req.GetPrivateKeyVaultResource(), req.GetCertificateVaultResource(), req.GetThresholdOfDaysToExpire(), req.GetDomains())
	if err != nil {
		return nil, xerrors.Errorf("(usecase.CertificatesUseCase).IssueCertificate: %w", err)
	}

	resp = &cloudacme.IssueCertificateResponse{
		PrivateKeyVaultVersionResource:  privateKeyVaultVersionResource,
		CertificateVaultVersionResource: certificateVaultVersionResource,
	}

	return resp, nil
}