// nolint: testpackage
package usecase

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"io"
	"testing"

	"github.com/newtstat/cloudacme/contexts"
	"github.com/newtstat/cloudacme/repository"
	"github.com/newtstat/cloudacme/test/fixture"
	"github.com/newtstat/cloudacme/test/mock"
	"github.com/newtstat/nits.go"
	"github.com/rec-logger/rec.go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewPrivateKeyUseCase(t *testing.T) {
	t.Parallel()
	t.Run("success()", func(t *testing.T) {
		t.Parallel()
		_ = NewPrivateKeyUsecase(nil)
	})
}

func Test_privateKeyUseCase_GetPrivateKey(t *testing.T) {
	t.Parallel()
	t.Run("failure()", func(t *testing.T) {
		t.Parallel()
		s := NewPrivateKeyUsecase(&mock.VaultRepository{GetVaultVersionDataIfExistsErr: status.Error(codes.Internal, "")})
		renewed, _, err := s.GetPrivateKey(context.TODO(), "test", true, nits.CryptoEd25519)
		if err == nil {
			t.Errorf("err == nil")
		}
		if renewed {
			t.Errorf("renewed")
		}
	})
}

func Test_privateKeyUseCase_getPrivateKey(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                         context.Context
		privateKeyVaultResource     string
		renewPrivateKey             bool
		keyAlgorithm                string
		parsePKCSXPrivateKeyPEMFunc func(pemData []byte) (crypto.PrivateKey, error)
		generateKeyFunc             func(algorithm string) (crypto.PrivateKey, error)
	}
	tests := []struct {
		name        string
		vaultRepo   repository.VaultRepository
		args        args
		wantErr     bool
		wantRenewed bool
	}{
		{"success(renewPrivateKey)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionDataIfExistsErr: nil, CreateVaultIfNotExistsErr: nil}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), "test", true, "", func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey([]byte("test")), nil }, func(algorithm string) (crypto.PrivateKey, error) { return ed25519.PrivateKey([]byte("test")), nil }}, false, true},
		{"failure(GetVaultVersionDataIfExists)", &mock.VaultRepository{GetVaultVersionDataIfExistsErr: status.Error(codes.Internal, "")}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), "test", true, "rsa2048", nil, nil}, true, false},
		{"success(parsePKCSXPrivateKeyPEMFunc)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionDataIfExistsErr: nil, CreateVaultIfNotExistsErr: nil}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), "test", false, "rsa2048", func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey([]byte("test")), nil }, func(algorithm string) (crypto.PrivateKey, error) { return ed25519.PrivateKey([]byte("test")), nil }}, false, false},
		{"success(parsePKCSXPrivateKeyPEMFuncErr)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionDataIfExistsErr: nil, CreateVaultIfNotExistsErr: nil}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), "test", false, "rsa2048", func(pemData []byte) (crypto.PrivateKey, error) { return nil, fixture.ErrTestError }, func(algorithm string) (crypto.PrivateKey, error) { return ed25519.PrivateKey([]byte("test")), nil }}, false, true},
		{"failure(generateKeyFunc)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionDataIfExistsErr: nil}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), "test", true, "rsa2048", func(pemData []byte) (crypto.PrivateKey, error) { return nil, fixture.ErrTestError }, func(algorithm string) (crypto.PrivateKey, error) { return nil, fixture.ErrTestError }}, true, false},
		{"failure(CreateVaultIfNotExists)", &mock.VaultRepository{GetVaultVersionDataIfExistsBool: true, GetVaultVersionDataIfExistsResource: "test", GetVaultVersionDataIfExistsData: []byte("test"), GetVaultVersionDataIfExistsErr: nil, CreateVaultIfNotExistsErr: fixture.ErrTestError}, args{contexts.WithLogger(context.TODO(), rec.Must(rec.New(io.Discard))), "test", true, "rsa2048", func(pemData []byte) (crypto.PrivateKey, error) { return ed25519.PrivateKey([]byte("test")), nil }, func(algorithm string) (crypto.PrivateKey, error) { return ed25519.PrivateKey([]byte("test")), nil }}, true, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := &privateKeyUseCase{
				vaultRepo: tt.vaultRepo,
			}
			renewed, _, err := s.getPrivateKey(tt.args.ctx, tt.args.privateKeyVaultResource, tt.args.renewPrivateKey, tt.args.keyAlgorithm, tt.args.parsePKCSXPrivateKeyPEMFunc, tt.args.generateKeyFunc)
			if (err != nil) != tt.wantErr {
				t.Errorf("privateKeyUseCase.getPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if renewed != tt.wantRenewed {
				t.Errorf("privateKeyUseCase.getPrivateKey() renewed = %v, wantRenewed %v", renewed, tt.wantRenewed)
			}
		})
	}
}

func Test_privateKeyUseCase_Lock(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                      context.Context
		privateKeyVaultResource  string
		certificateVaultResource string
	}
	tests := []struct {
		name      string
		vaultRepo repository.VaultRepository
		args      args
		wantErr   bool
	}{
		{"success()", &mock.VaultRepository{}, args{context.Background(), "test", "test"}, false},
		{"failure()", &mock.VaultRepository{LockVaultErr: fixture.ErrTestError}, args{context.Background(), "test", "test"}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := &privateKeyUseCase{
				vaultRepo: tt.vaultRepo,
			}
			if err := uc.Lock(tt.args.ctx, tt.args.privateKeyVaultResource, tt.args.certificateVaultResource); (err != nil) != tt.wantErr {
				t.Errorf("privateKeyUseCase.Lock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_privateKeyUseCase_Unlock(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                      context.Context
		privateKeyVaultResource  string
		certificateVaultResource string
	}
	tests := []struct {
		name      string
		vaultRepo repository.VaultRepository
		args      args
		wantErr   bool
	}{
		{"success()", &mock.VaultRepository{}, args{context.Background(), "test", "test"}, false},
		{"failure()", &mock.VaultRepository{UnlockVaultErr: fixture.ErrTestError}, args{context.Background(), "test", "test"}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := &privateKeyUseCase{
				vaultRepo: tt.vaultRepo,
			}
			if err := uc.Unlock(tt.args.ctx, tt.args.privateKeyVaultResource, tt.args.certificateVaultResource); (err != nil) != tt.wantErr {
				t.Errorf("privateKeyUseCase.Unlock() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
