// nolint: testpackage
package controller

import (
	"context"
	"crypto"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/newtstat/cloudacme/contexts"
	"github.com/newtstat/cloudacme/proto/generated/go/v1/cloudacme"
	"github.com/newtstat/cloudacme/repository"
	"github.com/newtstat/cloudacme/test/fixture"
	"github.com/newtstat/cloudacme/test/mock"
	"github.com/newtstat/nits.go"
	"github.com/rec-logger/rec.go"
)

func TestCertificatesController_Issue(t *testing.T) {
	t.Parallel()
	t.Run("failure()", func(t *testing.T) {
		t.Parallel()
		c := &CertificatesController{}
		_, err := c.Issue(context.Background(), &cloudacme.IssueCertificateRequest{})
		if err == nil {
			t.Errorf("*CertificatesController.Issue() error = %v, want not-nil", err)
			return
		}
		const prefix = `provider="": `
		if !strings.HasPrefix(err.Error(), prefix) {
			t.Errorf("*CertificatesController.Issue() error should has prefix: %s", prefix)
		}
	})
}

func TestCertificatesController_issue(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                                   context.Context
		req                                   *cloudacme.IssueCertificateRequest
		newVaultGoogleSecretManagerRepository func(ctx context.Context) (repository.VaultRepository, error)
		newLetsEncryptGoogleCloudRepository   func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error)
	}
	tests := []struct {
		name     string
		args     args
		wantResp *cloudacme.IssueCertificateResponse
		wantErr  bool
	}{
		{
			"success(gcloud)",
			args{
				contexts.WithLogger(context.Background(), rec.Must(rec.New(io.Discard))),
				&cloudacme.IssueCertificateRequest{VaultProvider: "gcloud", KeyAlgorithm: nits.CryptoEd25519, DnsProvider: "gcloud", TermsOfServiceAgreed: true, PrivateKeyVaultResource: "test"},
				func(ctx context.Context) (repository.VaultRepository, error) {
					return &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true}, nil
				},
				func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error) {
					return &mock.LetsEncryptRepository{}, nil
				},
			},
			&cloudacme.IssueCertificateResponse{},
			false,
		},
		{
			"failure(vaultProviderIsEmpty)",
			args{
				contexts.WithLogger(context.Background(), rec.Must(rec.New(io.Discard))),
				&cloudacme.IssueCertificateRequest{VaultProvider: ""},
				func(ctx context.Context) (repository.VaultRepository, error) {
					return &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true}, nil
				},
				func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error) {
					return &mock.LetsEncryptRepository{}, nil
				},
			},
			nil,
			true,
		},
		{
			"failure(gcloud,newVaultGoogleSecretManagerRepository)",
			args{
				contexts.WithLogger(context.Background(), rec.Must(rec.New(io.Discard))),
				&cloudacme.IssueCertificateRequest{VaultProvider: "gcloud"},
				func(ctx context.Context) (repository.VaultRepository, error) { return nil, fixture.ErrTestError },
				func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error) {
					return &mock.LetsEncryptRepository{}, nil
				},
			},
			nil,
			true,
		},
		{
			"failure(gcloud,privateKeyUsecase.Lock)",
			args{
				contexts.WithLogger(context.Background(), rec.Must(rec.New(io.Discard))),
				&cloudacme.IssueCertificateRequest{VaultProvider: "gcloud", KeyAlgorithm: nits.CryptoEd25519, TermsOfServiceAgreed: true},
				func(ctx context.Context) (repository.VaultRepository, error) {
					return &mock.VaultRepository{LockVaultErr: fixture.ErrTestError}, nil
				},
				func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error) {
					return &mock.LetsEncryptRepository{}, nil
				},
			},
			nil,
			true,
		},
		{
			"failure(gcloud,privateKeyUsecase.Unlock)",
			args{
				contexts.WithLogger(context.Background(), rec.Must(rec.New(io.Discard))),
				&cloudacme.IssueCertificateRequest{VaultProvider: "gcloud", KeyAlgorithm: nits.CryptoEd25519, TermsOfServiceAgreed: true},
				func(ctx context.Context) (repository.VaultRepository, error) {
					return &mock.VaultRepository{UnlockVaultErr: fixture.ErrTestError}, nil
				},
				func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error) {
					return &mock.LetsEncryptRepository{}, nil
				},
			},
			nil,
			true,
		},
		{
			"failure(gcloud,privateKeyUsecase.GetPrivateKey)",
			args{
				contexts.WithLogger(context.Background(), rec.Must(rec.New(io.Discard))),
				&cloudacme.IssueCertificateRequest{VaultProvider: "gcloud", KeyAlgorithm: nits.CryptoEd25519, TermsOfServiceAgreed: true},
				func(ctx context.Context) (repository.VaultRepository, error) {
					return &mock.VaultRepository{GetVaultVersionDataIfExistsErr: fixture.ErrTestError}, nil
				},
				func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error) {
					return &mock.LetsEncryptRepository{}, nil
				},
			},
			nil,
			true,
		},
		{
			"failure(gcloud,newLetsEncryptGoogleCloudRepository)",
			args{
				contexts.WithLogger(context.Background(), rec.Must(rec.New(io.Discard))),
				&cloudacme.IssueCertificateRequest{VaultProvider: "gcloud", KeyAlgorithm: nits.CryptoEd25519, DnsProvider: "gcloud", TermsOfServiceAgreed: true},
				func(ctx context.Context) (repository.VaultRepository, error) {
					return &mock.VaultRepository{}, nil
				},
				func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error) {
					return nil, fixture.ErrTestError
				},
			},
			nil,
			true,
		},
		{
			"failure(dnsProviderIsEmpty)",
			args{
				contexts.WithLogger(context.Background(), rec.Must(rec.New(io.Discard))),
				&cloudacme.IssueCertificateRequest{VaultProvider: "gcloud", KeyAlgorithm: nits.CryptoEd25519, TermsOfServiceAgreed: true},
				func(ctx context.Context) (repository.VaultRepository, error) {
					return &mock.VaultRepository{}, nil
				},
				func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error) {
					return nil, fixture.ErrTestError
				},
			},
			nil,
			true,
		},
		{
			"failure(svc.IssueCertificate)",
			args{
				contexts.WithLogger(context.Background(), rec.Must(rec.New(io.Discard))),
				&cloudacme.IssueCertificateRequest{VaultProvider: "gcloud", KeyAlgorithm: nits.CryptoEd25519, DnsProvider: "gcloud", TermsOfServiceAgreed: true},
				func(ctx context.Context) (repository.VaultRepository, error) {
					return &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true}, nil
				},
				func(ctx context.Context, termsOfServiceAgreed bool, email string, privateKey crypto.PrivateKey, googleCloudProject string, staging bool, logWriter io.Writer) (repository.LetsEncryptRepository, error) {
					return &mock.LetsEncryptRepository{
						IssueCertificateErr: fixture.ErrTestError,
					}, nil
				},
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &CertificatesController{}
			gotResp, err := c.issue(tt.args.ctx, tt.args.req, tt.args.newVaultGoogleSecretManagerRepository, tt.args.newLetsEncryptGoogleCloudRepository)
			if (err != nil) != tt.wantErr {
				t.Errorf("CertificatesController.issue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("CertificatesController.issue() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}
