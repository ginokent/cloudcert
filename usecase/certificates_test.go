// nolint: testpackage
package usecase

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/tls"
	"io"
	"testing"

	"github.com/newtstat/cloudacme/contexts"
	"github.com/newtstat/cloudacme/repository"
	"github.com/newtstat/cloudacme/test/fixture"
	"github.com/newtstat/cloudacme/test/mock"
	"github.com/newtstat/nits.go"
	"github.com/rec-logger/rec.go"
)

var testPrivateKey = nits.Crypto.MustGenerateKey(nits.Crypto.GenerateKey("rsa512"))

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
		{"success()", args{&mock.VaultRepository{}, &mock.LetsEncryptRepository{}}},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_ = NewCertificatesUseCase(tt.args.certificatesRepo, tt.args.letsencryptRepo)
		})
	}
}

func Test_certificatesUseCase_IssueCertificate(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                      context.Context
		privateKey               crypto.PrivateKey
		renewPrivateKey          bool
		privateKeyVaultResource  string
		certificateVaultResource string
		thresholdOfDaysToExpire  int64
		domains                  []string
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
		{"failure(GetVaultVersionDataIfExists)", &mock.VaultRepository{GetVaultVersionDataIfExistsErr: fixture.ErrTestError}, nil, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, true, "test", "test", 30, []string{"localhost"}}, "", "", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &certificatesUseCase{
				vaultRepo:       tt.vaultRepo,
				letsencryptRepo: tt.letsencryptRepo,
			}
			gotPrivateKeyVaultVersionResource, gotCertificateVaultVersionResource, err := s.IssueCertificate(tt.args.ctx, nil, tt.args.renewPrivateKey, tt.args.privateKeyVaultResource, tt.args.certificateVaultResource, tt.args.thresholdOfDaysToExpire, tt.args.domains)
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

func Test_certificatesUseCase_issueCertificate(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                           context.Context
		privateKey                    crypto.PrivateKey
		renewPrivateKey               bool
		privateKeyVaultResource       string
		certificateVaultResource      string
		thresholdOfDaysToExpire       int64
		domains                       []string
		checkCertificatePEMFunc       func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error)
		parsePKCSXPrivateKeyPEMFunc   func(pemData []byte) (crypto.PrivateKey, error)
		marshalPKCSXPrivateKeyPEMFunc func(privateKey crypto.PrivateKey) (pemData []byte, err error)
		tls_X509KeyPair               func(certPEMBlock []byte, keyPEMBlock []byte) (tls.Certificate, error) // nolint: revive
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
		{"success(NotExpired)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, -1, false, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "test", "test", false},
		{"success(Expired)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "test", "test", false},
		{"success(RenewPrivateKey)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, true, "test", "test", 30, []string{"localhost"}, nil, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "test", "test", false},
		{"success(parsePKCSXPrivateKeyPEMFunc)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, -1, false, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return nil, fixture.ErrTestError }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "test", "test", false},
		{"success(marshalPKCSXPrivateKeyPEMFunc)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, -1, false, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, fixture.ErrTestError }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
			return tls.Certificate{}, fixture.ErrTestError
		}}, "test", "test", false},
		{"success(tls_X509KeyPair)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, -1, false, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
			return tls.Certificate{}, fixture.ErrTestError
		}}, "test", "test", false},
		{"success(keyPairIsBroken,x509.MarshalPKCS8PrivateKey)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), []byte("test"), false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", false},
		{"success(keyPairIsBroken,tls_X509KeyPair)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, 0, false, 0, fixture.ErrTestError
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "test", "test", false},
		{"failure(GetVaultVersionDataIfExists)", &mock.VaultRepository{GetVaultVersionDataIfExistsErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, -1, false, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
		{"failure(CreateVaultIfNotExists)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", CreateVaultIfNotExistsErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
		{"failure(IssueCertificate)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test"}, &mock.LetsEncryptRepository{IssueCertificateErr: fixture.ErrTestError}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
		{"failure(RenewPrivateKey,AddVaultVersion)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, true, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
		{"failure(AddVaultVersion)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), testPrivateKey, false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey("test"), nil }, func(privateKey crypto.PrivateKey) (pemData []byte, err error) { return nil, nil }, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &certificatesUseCase{
				vaultRepo:       tt.vaultRepo,
				letsencryptRepo: tt.letsencryptRepo,
			}
			// gotPrivateKeyVaultVersionResource, gotCertificateVaultVersionResource, err := s.issueCertificate(tt.args.ctx, tt.args.privateKey, tt.args.renewPrivateKey, tt.args.privateKeyVaultResource, tt.args.certificateVaultResource, tt.args.thresholdOfDaysToExpire, tt.args.domains, tt.args.checkCertificatePEMFunc, tt.args.parsePKCSXPrivateKeyPEMFunc, tt.args.tls_X509KeyPair)
			_, gotCertificateVaultVersionResource, err := s.issueCertificate(tt.args.ctx, tt.args.privateKey, tt.args.renewPrivateKey, tt.args.privateKeyVaultResource, tt.args.certificateVaultResource, tt.args.thresholdOfDaysToExpire, tt.args.domains, tt.args.checkCertificatePEMFunc, tt.args.parsePKCSXPrivateKeyPEMFunc, tt.args.marshalPKCSXPrivateKeyPEMFunc, tt.args.tls_X509KeyPair)
			if (err != nil) != tt.wantErr {
				t.Errorf("certificatesUseCase.issueCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if gotPrivateKeyVaultVersionResource != tt.wantPrivateKeyVaultVersionResource {
			// 	t.Errorf("certificatesUseCase.issueCertificate() gotPrivateKeyVaultVersionResource = %v, want %v", gotPrivateKeyVaultVersionResource, tt.wantPrivateKeyVaultVersionResource)
			// }
			if gotCertificateVaultVersionResource != tt.wantCertificateVaultVersionResource {
				t.Errorf("certificatesUseCase.issueCertificate() gotCertificateVaultVersionResource = %v, want %v", gotCertificateVaultVersionResource, tt.wantCertificateVaultVersionResource)
			}
		})
	}
}
