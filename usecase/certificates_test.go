// nolint: testpackage
package usecase

import (
	"context"
	"crypto"
	"crypto/tls"
	"io"
	"testing"

	"github.com/ginokent/cloudacme/contexts"
	"github.com/ginokent/cloudacme/repository"
	"github.com/ginokent/cloudacme/test/fixture"
	"github.com/ginokent/cloudacme/test/mock"
	"github.com/nitpickers/nits.go"
	"github.com/rec-logger/rec.go"
)

func TestNewCertificatesUseCase(t *testing.T) {
	t.Parallel()
	type args struct {
		certificatesRepo repository.VaultRepository
		letsencryptRepo  repository.LetsEncryptRepository
	}
	tests := []struct {
		name string
		args args
	}{
		{"success()", args{nil, nil}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_ = NewCertificatesUseCase(tt.args.certificatesRepo, tt.args.letsencryptRepo)
		})
	}
}

func Test_certificatesUseCase_lock(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx       context.Context
		resources []string
	}
	tests := []struct {
		name            string
		vaultRepo       repository.VaultRepository
		letsencryptRepo repository.LetsEncryptRepository
		args            args
		wantUnlock      func()
		wantErr         bool
	}{
		{"success()", &mock.VaultRepository{}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.L().RenewWriter(io.Discard)), []string{"test"}}, nil, false},
		{"failure(uc.vaultRepo.CreateVaultIfNotExists)", &mock.VaultRepository{CreateVaultIfNotExistsErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.L().RenewWriter(io.Discard)), []string{"test"}}, nil, true},
		{"failure(uc.vaultRepo.LockVault)", &mock.VaultRepository{LockVaultErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.L().RenewWriter(io.Discard)), []string{"test"}}, nil, true},
		{"failure(uc.vaultRepo.UnlockVault)", &mock.VaultRepository{UnlockVaultErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.L().RenewWriter(io.Discard)), []string{"test"}}, nil, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := &certificatesUseCase{
				vaultRepo:       tt.vaultRepo,
				letsencryptRepo: tt.letsencryptRepo,
			}
			unlock, err := uc.lock(tt.args.ctx, tt.args.resources...)
			if (err != nil) != tt.wantErr {
				t.Errorf("certificatesUseCase.lock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			unlock()
		})
	}
}

func Test_certificatesUseCase_IssueCertificate(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                         context.Context
		acmeAccountKeyVaultResource string
		privateKeyVaultResource     string
		certificateVaultResource    string
		renewPrivateKey             bool
		keyAlgorithm                string
		thresholdOfDaysToExpire     int64
		domains                     []string
	}
	tests := []struct {
		name                                string
		vaultRepo                           repository.VaultRepository
		letsencryptRepo                     repository.LetsEncryptRepository
		args                                args
		wantPrivateKeyVaultVersionResource  string
		wantCertificateVaultVersionResource string
		wantErr                             bool
	}{
		{
			"success()",
			&mock.VaultRepository{AddVaultVersionResource: "test"},
			&mock.LetsEncryptRepository{},
			args{
				contexts.WithLogger(context.TODO(), rec.L().RenewWriter(io.Discard)),
				"",
				"",
				"",
				true,
				"rsa128",
				0,
				[]string{""},
			},
			"test",
			"test",
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := &certificatesUseCase{
				vaultRepo:       tt.vaultRepo,
				letsencryptRepo: tt.letsencryptRepo,
			}
			gotPrivateKeyVaultVersionResource, gotCertificateVaultVersionResource, err := uc.IssueCertificate(tt.args.ctx, tt.args.acmeAccountKeyVaultResource, tt.args.privateKeyVaultResource, tt.args.certificateVaultResource, tt.args.renewPrivateKey, tt.args.keyAlgorithm, tt.args.thresholdOfDaysToExpire, tt.args.domains)
			if (err != nil) != tt.wantErr {
				t.Errorf("certificatesUseCase.IssueCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPrivateKeyVaultVersionResource != tt.wantPrivateKeyVaultVersionResource {
				t.Errorf("certificatesUseCase.IssueCertificate() gotPrivateKeyVaultVersionResource = %v, want %v", gotPrivateKeyVaultVersionResource, tt.wantPrivateKeyVaultVersionResource)
			}
			if gotCertificateVaultVersionResource != tt.wantCertificateVaultVersionResource {
				t.Errorf("certificatesUseCase.IssueCertificate() gotCertificateVaultVersionResource = %v, want %v", gotCertificateVaultVersionResource, tt.wantCertificateVaultVersionResource)
			}
		})
	}
}

func Test_certificatesUseCase__issueCertificate(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                            context.Context
		acmeAccountKeyVaultResource    string
		privateKeyVaultResource        string
		certificateVaultResource       string
		renewPrivateKey                bool
		keyAlgorithm                   string
		thresholdOfDaysToExpire        int64
		domains                        []string
		_certcrypto_ParsePEMPrivateKey func(key []byte) (crypto.PrivateKey, error)                                                        // nolint: revive,stylecheck
		_tls_X509KeyPair               func(certPEMBlock []byte, keyPEMBlock []byte) (tls.Certificate, error)                             // nolint: revive,stylecheck
		_nits_X509_CheckCertificatePEM func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) // nolint: revive,stylecheck
		_nits_Crypto_GenerateKey       func(algorithm string) (crypto.PrivateKey, error)                                                  // nolint: revive,stylecheck
	}
	tests := []struct {
		name                                string
		vaultRepo                           repository.VaultRepository
		letsencryptRepo                     repository.LetsEncryptRepository
		args                                args
		wantPrivateKeyVaultVersionResource  string
		wantCertificateVaultVersionResource string
		wantErr                             bool
	}{
		{
			"success()",
			&mock.VaultRepository{AddVaultVersionResource: "test"},
			&mock.LetsEncryptRepository{},
			args{
				contexts.WithLogger(context.TODO(), rec.L().RenewWriter(io.Discard)),
				"",
				"",
				"",
				true,
				"rsa128",
				0,
				[]string{""},
				func(key []byte) (crypto.PrivateKey, error) { return nil, nil },
				func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil },
				func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
					return false, -1, false, 99, nil
				},
				func(algorithm string) (crypto.PrivateKey, error) { return nits.Crypto.GenerateKey("rsa128") }, // nolint: wrapcheck
			},
			"test", "test", false,
		},
		{
			"success()",
			&mock.VaultRepository{AddVaultVersionResource: "test"},
			&mock.LetsEncryptRepository{},
			args{
				contexts.WithLogger(context.TODO(), rec.L().RenewWriter(io.Discard)),
				"",
				"",
				"",
				true,
				"rsa128",
				0,
				[]string{""},
				func(key []byte) (crypto.PrivateKey, error) { return nil, nil },
				func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil },
				func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
					return false, -1, false, 99, nil
				},
				func(algorithm string) (crypto.PrivateKey, error) { return nits.Crypto.GenerateKey("rsa128") }, // nolint: wrapcheck
			},
			"test", "test", false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := &certificatesUseCase{
				vaultRepo:       tt.vaultRepo,
				letsencryptRepo: tt.letsencryptRepo,
			}
			gotPrivateKeyVaultVersionResource, gotCertificateVaultVersionResource, err := uc._issueCertificate(tt.args.ctx, tt.args.acmeAccountKeyVaultResource, tt.args.privateKeyVaultResource, tt.args.certificateVaultResource, tt.args.renewPrivateKey, tt.args.keyAlgorithm, tt.args.thresholdOfDaysToExpire, tt.args.domains, tt.args._certcrypto_ParsePEMPrivateKey, tt.args._tls_X509KeyPair, tt.args._nits_X509_CheckCertificatePEM, tt.args._nits_Crypto_GenerateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("certificatesUseCase._issueCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPrivateKeyVaultVersionResource != tt.wantPrivateKeyVaultVersionResource {
				t.Errorf("certificatesUseCase._issueCertificate() gotPrivateKeyVaultVersionResource = %v, want %v", gotPrivateKeyVaultVersionResource, tt.wantPrivateKeyVaultVersionResource)
			}
			if gotCertificateVaultVersionResource != tt.wantCertificateVaultVersionResource {
				t.Errorf("certificatesUseCase._issueCertificate() gotCertificateVaultVersionResource = %v, want %v", gotCertificateVaultVersionResource, tt.wantCertificateVaultVersionResource)
			}
		})
	}
}
