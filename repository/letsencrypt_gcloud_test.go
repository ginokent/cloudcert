// nolint: testpackage
package repository

import (
	"context"
	"crypto"
	"io"
	"net/http"
	"testing"

	"github.com/ginokent/cloudacme/test/fixture"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/providers/dns/gcloud"
	"github.com/go-acme/lego/v4/registration"
	"github.com/nitpickers/nits.go"
)

var (
	_ challenger          = (*mockChallenger)(nil)
	_ registerer          = (*mockRegisterer)(nil)
	_ certificateObtainer = (*mockObtainer)(nil)
)

type mockChallenger struct {
	SetDNS01ProviderError error
}

func (t *mockChallenger) SetDNS01Provider(challenge.Provider, ...dns01.ChallengeOption) error {
	return t.SetDNS01ProviderError
}

type mockRegisterer struct {
	RegisterReturn *registration.Resource
	RegisterError  error
}

func (t *mockRegisterer) Register(registration.RegisterOptions) (*registration.Resource, error) {
	return t.RegisterReturn, t.RegisterError
}

type mockObtainer struct {
	ObtainReturn *certificate.Resource
	ObtainError  error
}

func (t *mockObtainer) Obtain(certificate.ObtainRequest) (*certificate.Resource, error) {
	return t.ObtainReturn, t.ObtainError
}

const (
	testEmail = "root@localhost"
)

func TestNewLetsEncryptGoogleCloudRepository(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                  context.Context
		termsOfServiceAgreed bool
		email                string
		googleCloudProject   string
		staging              bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"success()", args{context.TODO(), true, testEmail, "", false}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewLetsEncryptGoogleCloudRepository(tt.args.ctx, tt.args.termsOfServiceAgreed, tt.args.email, tt.args.googleCloudProject, tt.args.staging, io.Discard)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLetsEncryptGoogleCloudRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_newLetsEncryptGoogleCloudRepository(t *testing.T) {
	t.Parallel()
	successGoogleDefaultClient := func(ctx context.Context, scope ...string) (*http.Client, error) { return http.DefaultClient, nil }
	failureGoogleDefaultClient := func(ctx context.Context, scope ...string) (*http.Client, error) { return nil, fixture.ErrTestError }
	successGcloudNewDNSProviderConfig := func(config *gcloud.Config) (*gcloud.DNSProvider, error) { return &gcloud.DNSProvider{}, nil }
	failureGcloudNewDNSProviderConfig := func(config *gcloud.Config) (*gcloud.DNSProvider, error) { return nil, fixture.ErrTestError }
	type args struct {
		ctx                         context.Context
		termsOfServiceAgreed        bool
		email                       string
		googleCloudProject          string
		staging                     bool
		google_DefaultClient        func(ctx context.Context, scope ...string) (*http.Client, error) // nolint: revive,stylecheck
		gcloud_NewDNSProviderConfig func(config *gcloud.Config) (*gcloud.DNSProvider, error)         // nolint: revive,stylecheck
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"success()" /************************/, args{context.TODO(), true /***/, testEmail, "", true, successGoogleDefaultClient, successGcloudNewDNSProviderConfig}, false},
		{"failure(google.DefaultClient)" /****/, args{context.TODO(), true /***/, testEmail, "", true, failureGoogleDefaultClient, successGcloudNewDNSProviderConfig}, true},
		{"failure(gcloud.NewDNSProviderConfig)", args{context.TODO(), true /***/, testEmail, "", true, successGoogleDefaultClient, failureGcloudNewDNSProviderConfig}, true},
		{"failure(termsOfServiceAgreed)" /****/, args{context.TODO(), false /**/, testEmail, "", true, successGoogleDefaultClient, successGcloudNewDNSProviderConfig}, true},
		{"failure(email)" /*******************/, args{context.TODO(), true /***/, "" /****/, "", true, successGoogleDefaultClient, successGcloudNewDNSProviderConfig}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := newLetsEncryptGoogleCloudRepository(tt.args.ctx, tt.args.termsOfServiceAgreed, tt.args.email, tt.args.googleCloudProject, tt.args.staging, io.Discard, tt.args.google_DefaultClient, tt.args.gcloud_NewDNSProviderConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("newLetsEncryptGoogleCloudRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_letsEncryptGoogleCloudDNSRepository_IssueCertificate(t *testing.T) {
	t.Parallel()

	type args struct {
		acmeAccountKey crypto.PrivateKey
		privateKey     crypto.PrivateKey
		domains        []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"failure(privateKey-is-nil)", args{nil, nil, nil}, true},
		{"failure(domains-is-nil)", args{nits.Crypto.MustGenerateKey(nits.Crypto.GenerateKey("rsa512")), nits.Crypto.MustGenerateKey(nits.Crypto.GenerateKey("rsa512")), nil}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			successGoogleDefaultClient := func(ctx context.Context, scope ...string) (*http.Client, error) { return http.DefaultClient, nil }
			successGcloudNewDNSProviderConfig := func(config *gcloud.Config) (*gcloud.DNSProvider, error) { return &gcloud.DNSProvider{}, nil }
			repo, repoErr := newLetsEncryptGoogleCloudRepository(context.TODO(), true, testEmail, "", true, io.Discard, successGoogleDefaultClient, successGcloudNewDNSProviderConfig)
			if repoErr != nil {
				t.Errorf("newLetsEncryptGoogleCloudRepository: %v", repoErr)
			}

			_, _, _, _, err := repo.IssueCertificate(context.TODO(), tt.args.acmeAccountKey, tt.args.privateKey, tt.args.domains)
			if (err != nil) != tt.wantErr {
				t.Errorf("letsEncryptGoogleCloudDNSRepository.IssueCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_letsEncryptGoogleCloudDNSRepository_issueCertificate(t *testing.T) {
	t.Parallel()

	type args struct {
		user                 *User
		challenge            challenger
		reg                  registerer
		termsOfServiceAgreed bool
		cert                 certificateObtainer
		privateKey           crypto.PrivateKey
		domains              []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"success()", args{&User{}, &mockChallenger{} /****************************************/, &mockRegisterer{} /********************************/, true, &mockObtainer{ObtainReturn: &certificate.Resource{}} /**/, nits.Crypto.MustGenerateKey(nits.Crypto.GenerateKey("rsa512")), []string{"localhost"}}, false},
		{"failure()", args{&User{}, &mockChallenger{SetDNS01ProviderError: fixture.ErrTestError}, &mockRegisterer{} /********************************/, true, &mockObtainer{} /***************************************/, nits.Crypto.MustGenerateKey(nits.Crypto.GenerateKey("rsa512")), []string{"localhost"}}, true},
		{"failure()", args{&User{}, &mockChallenger{} /****************************************/, &mockRegisterer{RegisterError: fixture.ErrTestError}, true, &mockObtainer{} /***************************************/, nits.Crypto.MustGenerateKey(nits.Crypto.GenerateKey("rsa512")), []string{"localhost"}}, true},
		{"failure()", args{&User{}, &mockChallenger{} /****************************************/, &mockRegisterer{} /********************************/, true, &mockObtainer{ObtainError: fixture.ErrTestError} /******/, nits.Crypto.MustGenerateKey(nits.Crypto.GenerateKey("rsa512")), []string{"localhost"}}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			successGoogleDefaultClient := func(ctx context.Context, scope ...string) (*http.Client, error) { return http.DefaultClient, nil }
			successGcloudNewDNSProviderConfig := func(config *gcloud.Config) (*gcloud.DNSProvider, error) { return &gcloud.DNSProvider{}, nil }
			repo, repoErr := newLetsEncryptGoogleCloudRepository(context.TODO(), true, testEmail, "", true, io.Discard, successGoogleDefaultClient, successGcloudNewDNSProviderConfig)
			if repoErr != nil {
				t.Errorf("newLetsEncryptGoogleCloudRepository: %v", repoErr)
			}

			_, _, _, _, err := repo.issueCertificate(context.TODO(), tt.args.user, tt.args.challenge, tt.args.reg, tt.args.termsOfServiceAgreed, tt.args.cert, tt.args.privateKey, tt.args.domains)
			if (err != nil) != tt.wantErr {
				t.Errorf("letsEncryptGoogleCloudDNSRepository.issueCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
