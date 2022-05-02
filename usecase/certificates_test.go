// nolint: testpackage
package usecase

import (
	"context"
	"crypto/tls"
	"io"
	"testing"

	"github.com/newtstat/cloudacme/contexts"
	"github.com/newtstat/cloudacme/repository"
	"github.com/newtstat/cloudacme/test/fixture"
	"github.com/newtstat/cloudacme/test/mock"
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
		privateKeyRenewed        bool
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
		{"failure(GetVaultVersionDataIfExists)", &mock.VaultRepository{GetVaultVersionDataIfExistsErr: fixture.ErrTestError}, nil, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), true, "test", "test", 30, []string{"localhost"}}, "", "", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &certificatesUseCase{
				vaultRepo:       tt.vaultRepo,
				letsencryptRepo: tt.letsencryptRepo,
			}
			gotPrivateKeyVaultVersionResource, gotCertificateVaultVersionResource, err := s.IssueCertificate(tt.args.ctx, tt.args.privateKeyRenewed, tt.args.privateKeyVaultResource, tt.args.certificateVaultResource, tt.args.thresholdOfDaysToExpire, tt.args.domains)
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
		ctx                      context.Context
		privateKeyRenewed        bool
		privateKeyVaultResource  string
		certificateVaultResource string
		thresholdOfDaysToExpire  int64
		domains                  []string
		checkCertificatePEMFunc  func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error)
		tls_X509KeyPair          func(certPEMBlock []byte, keyPEMBlock []byte) (tls.Certificate, error) // nolint: revive
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
		{"success(NotExpired)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, -1, false, 80, nil
		}, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "test", "test", false},
		{"success(Expired)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "test", "test", false},
		{"success(RenewPrivateKey)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), true, "test", "test", 30, []string{"localhost"}, nil, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "test", "test", false},
		{"success(certificateKeyPairIsBroken)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, -1, false, 80, nil
		}, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
			return tls.Certificate{}, fixture.ErrTestError
		}}, "test", "test", false},
		{"success(CertificateIsBroken)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionResource: "test"}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, 0, false, 0, fixture.ErrTestError
		}, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "test", "test", false},
		{"failure(GetVaultVersionDataIfExists)", &mock.VaultRepository{GetVaultVersionDataIfExistsErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return false, -1, false, 80, nil
		}, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
		{"failure(CreateVaultIfNotExists)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", CreateVaultIfNotExistsErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
		{"failure(IssueCertificate)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test"}, &mock.LetsEncryptRepository{IssueCertificateErr: fixture.ErrTestError}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
		{"failure(RenewPrivateKey,AddVaultVersion)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), true, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
		{"failure(AddVaultVersion)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionIfExistsBool: true, GetVaultVersionIfExistsVersion: "test", AddVaultVersionErr: fixture.ErrTestError}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), false, "test", "test", 30, []string{"localhost"}, func(pemData []byte) (notyet bool, daysToStart int64, expired bool, daysToExpire int64, err error) {
			return true, -1, true, 80, nil
		}, func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) { return tls.Certificate{}, nil }}, "", "", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &certificatesUseCase{
				vaultRepo:       tt.vaultRepo,
				letsencryptRepo: tt.letsencryptRepo,
			}
			gotPrivateKeyVaultVersionResource, gotCertificateVaultVersionResource, err := s.issueCertificate(tt.args.ctx, tt.args.privateKeyRenewed, tt.args.privateKeyVaultResource, tt.args.certificateVaultResource, tt.args.thresholdOfDaysToExpire, tt.args.domains, tt.args.checkCertificatePEMFunc, tt.args.tls_X509KeyPair)
			if (err != nil) != tt.wantErr {
				t.Errorf("certificatesUseCase.issueCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPrivateKeyVaultVersionResource != tt.wantPrivateKeyVaultVersionResource {
				t.Errorf("certificatesUseCase.issueCertificate() gotPrivateKeyVaultVersionResource = %v, want %v", gotPrivateKeyVaultVersionResource, tt.wantPrivateKeyVaultVersionResource)
			}
			if gotCertificateVaultVersionResource != tt.wantCertificateVaultVersionResource {
				t.Errorf("certificatesUseCase.issueCertificate() gotCertificateVaultVersionResource = %v, want %v", gotCertificateVaultVersionResource, tt.wantCertificateVaultVersionResource)
			}
		})
	}
}
