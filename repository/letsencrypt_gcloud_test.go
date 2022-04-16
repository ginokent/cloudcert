// nolint: testpackage
package repository

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"io"
	"net/http"
	"testing"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/gcloud"
	"github.com/go-acme/lego/v4/registration"
	"github.com/newtstat/cloudacme/test/fixture"
	"github.com/newtstat/nits.go"
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

var testPrivateKey = nits.Crypto.MustGenerateKey(ecdsa.GenerateKey(elliptic.P256(), rand.Reader))

func TestNewLetsEncryptGoogleCloudRepository(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                  context.Context
		termsOfServiceAgreed bool
		email                string
		privateKey           crypto.PrivateKey
		googleCloudProject   string
		staging              bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"success()", args{context.TODO(), true, testEmail, testPrivateKey, "", false}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewLetsEncryptGoogleCloudRepository(tt.args.ctx, tt.args.termsOfServiceAgreed, tt.args.email, tt.args.privateKey, tt.args.googleCloudProject, tt.args.staging, io.Discard)
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
	successLegoNewClient := func(config *lego.Config) (*lego.Client, error) { return &lego.Client{}, nil }
	failureLegoNewClient := func(config *lego.Config) (*lego.Client, error) { return nil, fixture.ErrTestError }
	type args struct {
		ctx                         context.Context
		termsOfServiceAgreed        bool
		email                       string
		privateKey                  crypto.PrivateKey
		googleCloudProject          string
		staging                     bool
		google_DefaultClient        func(ctx context.Context, scope ...string) (*http.Client, error) // nolint: revive
		gcloud_NewDNSProviderConfig func(config *gcloud.Config) (*gcloud.DNSProvider, error)         // nolint: revive
		lego_NewClient              func(config *lego.Config) (*lego.Client, error)                  // nolint: revive
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"success()" /************************/, args{context.TODO(), true /***/, testEmail, testPrivateKey, "", true, successGoogleDefaultClient, successGcloudNewDNSProviderConfig, successLegoNewClient}, false},
		{"failure(google.DefaultClient)" /****/, args{context.TODO(), true /***/, testEmail, testPrivateKey, "", true, failureGoogleDefaultClient, successGcloudNewDNSProviderConfig, successLegoNewClient}, true},
		{"failure(gcloud.NewDNSProviderConfig)", args{context.TODO(), true /***/, testEmail, testPrivateKey, "", true, successGoogleDefaultClient, failureGcloudNewDNSProviderConfig, successLegoNewClient}, true},
		{"failure(lego.NewClient)" /**********/, args{context.TODO(), true /***/, testEmail, testPrivateKey, "", true, successGoogleDefaultClient, successGcloudNewDNSProviderConfig, failureLegoNewClient}, true},
		{"failure(termsOfServiceAgreed)" /****/, args{context.TODO(), false /**/, testEmail, testPrivateKey, "", true, successGoogleDefaultClient, successGcloudNewDNSProviderConfig, successLegoNewClient}, true},
		{"failure(email)" /*******************/, args{context.TODO(), true /***/, "" /****/, testPrivateKey, "", true, successGoogleDefaultClient, successGcloudNewDNSProviderConfig, successLegoNewClient}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := newLetsEncryptGoogleCloudRepository(tt.args.ctx, tt.args.termsOfServiceAgreed, tt.args.email, tt.args.privateKey, tt.args.googleCloudProject, tt.args.staging, io.Discard, tt.args.google_DefaultClient, tt.args.gcloud_NewDNSProviderConfig, tt.args.lego_NewClient)
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
		domains []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"failure(domains-is-nil)", args{nil}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			successGoogleDefaultClient := func(ctx context.Context, scope ...string) (*http.Client, error) { return http.DefaultClient, nil }
			successGcloudNewDNSProviderConfig := func(config *gcloud.Config) (*gcloud.DNSProvider, error) { return &gcloud.DNSProvider{}, nil }
			successLegoNewClient := func(config *lego.Config) (*lego.Client, error) { return &lego.Client{}, nil }
			repo, repoErr := newLetsEncryptGoogleCloudRepository(context.TODO(), true, testEmail, testPrivateKey, "", true, io.Discard, successGoogleDefaultClient, successGcloudNewDNSProviderConfig, successLegoNewClient)
			if repoErr != nil {
				t.Errorf("newLetsEncryptGoogleCloudRepository: %v", repoErr)
			}

			if repo.user.GetEmail() != testEmail {
				t.Errorf("repo.user.GetEmail() != testEmail")
			}

			_, _, _, _, err := repo.IssueCertificate(context.TODO(), tt.args.domains)
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
		challenge            challenger
		reg                  registerer
		termsOfServiceAgreed bool
		cert                 certificateObtainer
		domains              []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"success()", args{&mockChallenger{} /****************************************/, &mockRegisterer{} /********************************/, true, &mockObtainer{ObtainReturn: &certificate.Resource{}} /**/, []string{"localhost"}}, false},
		{"failure()", args{&mockChallenger{SetDNS01ProviderError: fixture.ErrTestError}, &mockRegisterer{} /********************************/, true, &mockObtainer{} /***************************************/, []string{"localhost"}}, true},
		{"failure()", args{&mockChallenger{} /****************************************/, &mockRegisterer{RegisterError: fixture.ErrTestError}, true, &mockObtainer{} /***************************************/, []string{"localhost"}}, true},
		{"failure()", args{&mockChallenger{} /****************************************/, &mockRegisterer{} /********************************/, true, &mockObtainer{ObtainError: fixture.ErrTestError} /******/, []string{"localhost"}}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			successGoogleDefaultClient := func(ctx context.Context, scope ...string) (*http.Client, error) { return http.DefaultClient, nil }
			successGcloudNewDNSProviderConfig := func(config *gcloud.Config) (*gcloud.DNSProvider, error) { return &gcloud.DNSProvider{}, nil }
			successLegoNewClient := func(config *lego.Config) (*lego.Client, error) { return &lego.Client{}, nil }
			repo, repoErr := newLetsEncryptGoogleCloudRepository(context.TODO(), true, testEmail, testPrivateKey, "", true, io.Discard, successGoogleDefaultClient, successGcloudNewDNSProviderConfig, successLegoNewClient)
			if repoErr != nil {
				t.Errorf("newLetsEncryptGoogleCloudRepository: %v", repoErr)
			}

			_, _, _, _, err := repo.issueCertificate(context.TODO(), repo.user, tt.args.challenge, tt.args.reg, tt.args.termsOfServiceAgreed, tt.args.cert, tt.args.domains)
			if (err != nil) != tt.wantErr {
				t.Errorf("letsEncryptGoogleCloudDNSRepository.issueCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
