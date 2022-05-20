package repository

import (
	"context"
	"crypto"
	"io"
	"net/http"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/ginokent/cloudacme/trace"
	legocert "github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/gcloud"
	"github.com/go-acme/lego/v4/registration"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
)

var _ LetsEncryptRepository = (*letsEncryptGoogleCloudDNSRepository)(nil)

type letsEncryptGoogleCloudDNSRepository struct {
	termsOfServiceAgreed bool
	email                string
	staging              bool
	provider             challenge.Provider
}

func NewLetsEncryptGoogleCloudRepository(
	ctx context.Context,
	termsOfServiceAgreed bool,
	email string,
	googleCloudProject string,
	staging bool,
	logWriter io.Writer,
) (LetsEncryptRepository, error) {
	return newLetsEncryptGoogleCloudRepository(
		ctx,
		termsOfServiceAgreed,
		email,
		googleCloudProject,
		staging,
		logWriter,
		google.DefaultClient,
		gcloud.NewDNSProviderConfig,
	)
}

var legoLoggerOnce = &sync.Once{} // nolint: gochecknoglobals

func newLetsEncryptGoogleCloudRepository(
	ctx context.Context,
	termsOfServiceAgreed bool,
	email string,
	googleCloudProject string,
	staging bool,
	logWriter io.Writer,
	google_DefaultClient func(ctx context.Context, scope ...string) (*http.Client, error), // nolint: revive,stylecheck
	gcloud_NewDNSProviderConfig func(config *gcloud.Config) (*gcloud.DNSProvider, error), // nolint: revive,stylecheck
) (*letsEncryptGoogleCloudDNSRepository, error) {
	ctx, span := trace.Start(ctx, "repository.NewLetsEncryptGoogleCloudRepository")
	defer span.End()

	legoLoggerOnce.Do(func() {
		log.Logger = newLegoStdLogger(logWriter)
	})

	if !termsOfServiceAgreed {
		return nil, ErrTermsOfServiceNotAgreed
	}

	if email == "" {
		return nil, errors.Errorf("email: %w", ErrEmailIsEmpty)
	}

	var client *http.Client
	if err := trace.StartFunc(ctx, "google.DefaultClient")(func(child context.Context) (err error) {
		client, err = google_DefaultClient(ctx, dns.NdevClouddnsReadwriteScope)
		return
	}); err != nil {
		return nil, errors.Errorf("google.DefaultClient: googlecloud: unable to get Google Cloud client: %w", err)
	}

	config := gcloud.NewDefaultConfig()
	config.Project = googleCloudProject
	config.HTTPClient = client

	var dnsProvider *gcloud.DNSProvider
	if err := trace.StartFunc(ctx, "gcloud.NewDNSProviderConfig")(func(child context.Context) (err error) {
		dnsProvider, err = gcloud_NewDNSProviderConfig(config)
		return
	}); err != nil {
		return nil, errors.Errorf("gcloud.NewDNSProviderConfig: %w", err)
	}

	return &letsEncryptGoogleCloudDNSRepository{
		termsOfServiceAgreed: termsOfServiceAgreed,
		staging:              staging,
		email:                email,
		provider:             dnsProvider,
	}, nil
}

func (repo *letsEncryptGoogleCloudDNSRepository) newClient(ctx context.Context, acmeAccountKey crypto.PrivateKey) (user *User, legoClient *lego.Client, err error) {
	user = &User{
		email: repo.email,
		key:   acmeAccountKey,
	}

	legoConfig := lego.NewConfig(user)

	if repo.staging {
		legoConfig.CADirURL = lego.LEDirectoryStaging
	}

	if err := trace.StartFunc(ctx, "lego.NewClient")(func(child context.Context) (err error) {
		legoClient, err = lego.NewClient(legoConfig)
		return
	}); err != nil {
		return nil, nil, errors.Errorf("lego.NewClient: %w", err)
	}

	return user, legoClient, nil
}

// IssueCertificate issues a Let's Encrypt certificate.
func (repo *letsEncryptGoogleCloudDNSRepository) IssueCertificate(ctx context.Context, acmeAccountKey crypto.PrivateKey, privateKey crypto.PrivateKey, domains []string) (privateKeyPEM, certificatePEM, issuerCertificate, csr []byte, err error) {
	user, legoClient, err := repo.newClient(ctx, acmeAccountKey)
	if err != nil {
		return nil, nil, nil, nil, errors.Errorf("lego.NewClient: %w", err)
	}

	return repo.issueCertificate(
		ctx,
		user,
		legoClient.Challenge,
		legoClient.Registration,
		repo.termsOfServiceAgreed,
		legoClient.Certificate,
		privateKey,
		domains,
	)
}

type challenger interface {
	SetDNS01Provider(challenge.Provider, ...dns01.ChallengeOption) error
}

type registerer interface {
	Register(registration.RegisterOptions) (*registration.Resource, error)
}

type certificateObtainer interface {
	Obtain(legocert.ObtainRequest) (*legocert.Resource, error)
}

func (repo *letsEncryptGoogleCloudDNSRepository) issueCertificate(
	ctx context.Context,
	user *User,
	challenge challenger,
	reg registerer,
	termsOfServiceAgreed bool,
	cert certificateObtainer,
	privateKey crypto.PrivateKey,
	domains []string,
) (privateKeyPEM, certificatePEM, issuerCertificatePEM, csr []byte, err error) {
	ctx, span := trace.Start(ctx, "(*repository.vaultGoogleSecretManagerRepository).IssueCertificate")
	defer span.End()

	if len(domains) == 0 {
		return nil, nil, nil, nil, ErrDomainsIsNil
	}

	if err := trace.StartFunc(ctx, "(*resolver.SolverManager).SetDNS01Provider")(func(child context.Context) (err error) {
		err = challenge.SetDNS01Provider(repo.provider)
		return
	}); err != nil {
		return nil, nil, nil, nil, errors.Errorf("(*resolver.SolverManager).SetDNS01Provider: %w", err)
	}

	var registrationResource *registration.Resource
	if err := trace.StartFunc(ctx, "(*registration.Registrar).Register")(func(child context.Context) (err error) {
		registrationResource, err = reg.Register(registration.RegisterOptions{TermsOfServiceAgreed: termsOfServiceAgreed})
		return
	}); err != nil {
		return nil, nil, nil, nil, errors.Errorf("(*registration.Registrar).Register: %w", err)
	}

	user.registration = registrationResource

	request := legocert.ObtainRequest{
		Domains:    domains,
		PrivateKey: privateKey,
		Bundle:     true,
	}

	var certificateResource *legocert.Resource
	if err := trace.StartFunc(ctx, "(*legocert.Certifier).Obtain")(func(child context.Context) (err error) {
		certificateResource, err = cert.Obtain(request)
		return
	}); err != nil {
		return nil, nil, nil, nil, errors.Errorf("(*legocert.Certifier).Obtain: %w", err)
	}

	return certificateResource.PrivateKey, certificateResource.Certificate, certificateResource.IssuerCertificate, certificateResource.CSR, nil
}
