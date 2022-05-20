package controller

import (
	"context"
	"io"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/newtstat/cloudacme/proto/generated/go/v1/cloudacme"
	"github.com/newtstat/cloudacme/repository"
	"github.com/newtstat/cloudacme/trace"
	"github.com/newtstat/cloudacme/usecase"
)

type CertificatesController struct {
	cloudacme.UnimplementedCertificatesServer
}

func (c *CertificatesController) Issue(ctx context.Context, req *cloudacme.IssueCertificateRequest) (resp *cloudacme.IssueCertificateResponse, err error) {
	return c.issue(ctx, req, repository.NewVaultGoogleSecretManagerRepository, repository.NewLetsEncryptGoogleCloudRepository)
}

func (*CertificatesController) issue(
	ctx context.Context,
	req *cloudacme.IssueCertificateRequest,
	newVaultGoogleSecretManagerRepository func(ctx context.Context) (repository.VaultRepository, error),
	newLetsEncryptGoogleCloudRepository func(ctx context.Context, termsOfServiceAgreed bool, email string, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error),
) (
	resp *cloudacme.IssueCertificateResponse,
	err error,
) {
	ctx, span := trace.Start(ctx, "(*controller.CertificatesController).Issue")
	defer span.End()

	var vaultRepo repository.VaultRepository
	switch req.GetVaultProvider() {
	case cloudacme.IssueCertificateRequest_gcloud.String():
		vaultRepo, err = newVaultGoogleSecretManagerRepository(ctx)
		if err != nil {
			return nil, errors.Errorf("repository.NewVaultGoogleSecretManagerRepository: %w", err)
		}
	default:
		return nil, errors.Errorf("provider=\"%s\": %w", req.GetVaultProvider(), err)
	}

	var letsencryptRepo repository.LetsEncryptRepository
	switch req.GetDnsProvider() {
	case cloudacme.IssueCertificateRequest_gcloud.String():
		letsencryptRepo, err = newLetsEncryptGoogleCloudRepository(ctx, req.GetTermsOfServiceAgreed(), req.GetEmail(), req.GetDnsProviderID(), req.GetStaging(), os.Stdout)
		if err != nil {
			return nil, errors.Errorf("repository.NewLetsEncryptGoogleCloudRepository: %w", err)
		}
	default:
		return nil, errors.Errorf("provider=\"%s\": %w", req.GetVaultProvider(), err)
	}

	certUseCase := usecase.NewCertificatesUseCase(vaultRepo, letsencryptRepo)

	privateKeyVaultVersionResource, certificateVaultVersionResource, err := certUseCase.IssueCertificate(ctx, req.GetAcmeAccountKeyVaultResource(), req.GetPrivateKeyVaultResource(), req.GetCertificateVaultResource(), req.GetRenewPrivateKey(), req.GetKeyAlgorithm(), req.GetThresholdOfDaysToExpire(), req.GetDomains())
	if err != nil {
		return nil, errors.Errorf("usecase.CertificatesUseCase.IssueCertificate: %w", err)
	}

	resp = &cloudacme.IssueCertificateResponse{
		PrivateKeyVaultVersionResource:  privateKeyVaultVersionResource,
		CertificateVaultVersionResource: certificateVaultVersionResource,
	}

	return resp, nil
}
