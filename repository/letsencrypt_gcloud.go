package repository

import (
	"context"
	"crypto"
	"io"
	"net/http"
	"sync"

	legocert "github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/providers/dns/gcloud"
	"github.com/go-acme/lego/v4/registration"
	"github.com/newtstat/cloudacme/trace"
	"golang.org/x/oauth2/google"
	"golang.org/x/xerrors"
	"google.golang.org/api/dns/v1"
)

var _ LetsEncryptRepository = (*letsEncryptGoogleCloudDNSRepository)(nil)

type letsEncryptGoogleCloudDNSRepository struct {
	user                 *User
	termsOfServiceAgreed bool
	legoClient           *lego.Client
	provider             challenge.Provider
}

func NewLetsEncryptGoogleCloudRepository( // nolint: ireturn
	ctx context.Context,
	termsOfServiceAgreed bool,
	email string,
	privateKey crypto.PrivateKey,
	googleCloudProject string,
	staging bool,
	logWriter io.Writer,
) (LetsEncryptRepository, error) {
	return newLetsEncryptGoogleCloudRepository(
		ctx,
		termsOfServiceAgreed,
		email,
		privateKey,
		googleCloudProject,
		staging,
		logWriter,
		google.DefaultClient,
		gcloud.NewDNSProviderConfig,
		lego.NewClient,
	)
}

var legoLoggerOnce = &sync.Once{} // nolint: gochecknoglobals

// nolint: funlen
func newLetsEncryptGoogleCloudRepository(
	ctx context.Context,
	termsOfServiceAgreed bool,
	email string,
	privateKey crypto.PrivateKey,
	googleCloudProject string,
	staging bool,
	logWriter io.Writer,
	google_DefaultClient func(ctx context.Context, scope ...string) (*http.Client, error), // nolint: revive
	gcloud_NewDNSProviderConfig func(config *gcloud.Config) (*gcloud.DNSProvider, error), // nolint: revive
	lego_NewClient func(config *lego.Config) (*lego.Client, error), // nolint: revive
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
		return nil, xerrors.Errorf("email: %w", ErrEmailIsEmpty)
	}

	var client *http.Client
	if err := trace.StartFunc(ctx, "google.DefaultClient")(func(child context.Context) (err error) {
		client, err = google_DefaultClient(ctx, dns.NdevClouddnsReadwriteScope)
		return
	}); err != nil {
		return nil, xerrors.Errorf("google.DefaultClient: googlecloud: unable to get Google Cloud client: %w", err)
	}

	config := gcloud.NewDefaultConfig()
	config.Project = googleCloudProject
	config.HTTPClient = client

	var dnsProvider *gcloud.DNSProvider
	if err := trace.StartFunc(ctx, "gcloud.NewDNSProviderConfig")(func(child context.Context) (err error) {
		dnsProvider, err = gcloud_NewDNSProviderConfig(config)
		return
	}); err != nil {
		return nil, xerrors.Errorf("gcloud.NewDNSProviderConfig: %w", err)
	}

	user := &User{
		email: email,
		key:   privateKey,
	}

	legoConfig := lego.NewConfig(user)

	if staging {
		legoConfig.CADirURL = lego.LEDirectoryStaging
	}

	var legoClient *lego.Client
	if err := trace.StartFunc(ctx, "lego.NewClient")(func(child context.Context) (err error) {
		legoClient, err = lego_NewClient(legoConfig)
		return
	}); err != nil {
		return nil, xerrors.Errorf("lego.NewClient: %w", err)
	}

	return &letsEncryptGoogleCloudDNSRepository{
		user:                 user,
		termsOfServiceAgreed: termsOfServiceAgreed,
		legoClient:           legoClient,
		provider:             dnsProvider,
	}, nil
}

// IssueCertificate issues a Let's Encrypt certificate.
func (repo *letsEncryptGoogleCloudDNSRepository) IssueCertificate(ctx context.Context, domains []string) (privateKey, certificate, issuerCertificate, csr []byte, err error) {
	return repo.issueCertificate(
		ctx,
		repo.user,
		repo.legoClient.Challenge,
		repo.legoClient.Registration,
		repo.termsOfServiceAgreed,
		repo.legoClient.Certificate,
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
	domains []string,
) (privateKey, certificate, issuerCertificate, csr []byte, err error) {
	ctx, span := trace.Start(ctx, "(*repository.vaultGoogleSecretManagerRepository).IssueCertificate")
	defer span.End()

	if len(domains) == 0 {
		return nil, nil, nil, nil, ErrDomainsIsNil
	}

	if err := trace.StartFunc(ctx, "(*resolver.SolverManager).SetDNS01Provider")(func(child context.Context) (err error) {
		err = challenge.SetDNS01Provider(repo.provider)
		return
	}); err != nil {
		return nil, nil, nil, nil, xerrors.Errorf("(*resolver.SolverManager).SetDNS01Provider: %w", err)
	}

	var registrationResource *registration.Resource
	if err := trace.StartFunc(ctx, "(*registration.Registrar).Register")(func(child context.Context) (err error) {
		registrationResource, err = reg.Register(registration.RegisterOptions{TermsOfServiceAgreed: termsOfServiceAgreed})
		return
	}); err != nil {
		return nil, nil, nil, nil, xerrors.Errorf("(*registration.Registrar).Register: %w", err)
	}

	user.registration = registrationResource

	request := legocert.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	var certificateResource *legocert.Resource
	if err := trace.StartFunc(ctx, "(*legocert.Certifier).Obtain")(func(child context.Context) (err error) {
		certificateResource, err = cert.Obtain(request)
		return
	}); err != nil {
		return nil, nil, nil, nil, xerrors.Errorf("(*legocert.Certifier).Obtain: %w", err)
	}

	return certificateResource.PrivateKey, certificateResource.Certificate, certificateResource.IssuerCertificate, certificateResource.CSR, nil
}
