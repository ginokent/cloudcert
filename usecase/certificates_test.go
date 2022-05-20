// nolint: testpackage
package usecase

import (
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/ginokent/cloudacme/contexts"
	"github.com/ginokent/cloudacme/repository"
	"github.com/ginokent/cloudacme/test/mock"
	"github.com/rec-logger/rec.go"
)

func TestNewCertificatesUseCase(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			_ = NewCertificatesUseCase(tt.args.certificatesRepo, tt.args.letsencryptRepo)
		})
	}
}

func Test_certificatesUseCase_lock(t *testing.T) {
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
		{"success()", &mock.VaultRepository{}, &mock.LetsEncryptRepository{}, args{contexts.WithLogger(context.TODO(), rec.L().RenewWriter(io.Discard)), []string{"test"}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := &certificatesUseCase{
				vaultRepo:       tt.vaultRepo,
				letsencryptRepo: tt.letsencryptRepo,
			}
			gotUnlock, err := uc.lock(tt.args.ctx, tt.args.resources...)
			if (err != nil) != tt.wantErr {
				t.Errorf("certificatesUseCase.lock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotUnlock, tt.wantUnlock) {
				t.Errorf("certificatesUseCase.lock() = %v, want %v", gotUnlock, tt.wantUnlock)
			}
		})
	}
}
