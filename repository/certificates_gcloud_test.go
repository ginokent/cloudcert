// nolint: testpackage
package repository

import (
	"context"
	"reflect"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/googleapis/gax-go"
	"github.com/newtstat/cloudacme/config"
	"github.com/newtstat/cloudacme/test/fixture"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// nolint: paralleltest
func TestNewGoogleSecretManagerRepository(t *testing.T) {
	t.Run("failure(initSingletonGoogleSecretManagerClient)", func(t *testing.T) {
		singletonGoogleSecretManagerClientAtomic = 0
		backup := singletonGoogleSecretManagerClient
		singletonGoogleSecretManagerClient = nil
		t.Cleanup(func() { singletonGoogleSecretManagerClient = backup })
		if err := initSingletonGoogleSecretManagerClient(func(ctx context.Context, opts ...option.ClientOption) (*secretmanager.Client, error) {
			return nil, fixture.ErrTestError
		}); err == nil {
			t.Errorf("initSingletonGoogleSecretManagerClient(): err == nil")
		}
	})

	t.Run("failure(newVaultGoogleSecretManagerRepository)", func(t *testing.T) {
		singletonGoogleSecretManagerClientAtomic = 0
		backup := singletonGoogleSecretManagerClient
		singletonGoogleSecretManagerClient = nil
		t.Cleanup(func() { singletonGoogleSecretManagerClient = backup })

		if _, err := newVaultGoogleSecretManagerRepository(context.Background(), func(ctx context.Context, opts ...option.ClientOption) (*secretmanager.Client, error) {
			return nil, fixture.ErrTestError
		}); err == nil {
			t.Errorf("newVaultGoogleSecretManagerRepository(): err == nil")
		}
	})

	t.Run("success(newVaultGoogleSecretManagerRepository)", func(t *testing.T) {
		singletonGoogleSecretManagerClientAtomic = 0
		backup := singletonGoogleSecretManagerClient
		singletonGoogleSecretManagerClient = nil
		t.Cleanup(func() { singletonGoogleSecretManagerClient = backup })

		if _, err := newVaultGoogleSecretManagerRepository(context.Background(), func(ctx context.Context, opts ...option.ClientOption) (*secretmanager.Client, error) {
			return &secretmanager.Client{}, nil
		}); err != nil {
			t.Errorf("newVaultGoogleSecretManagerRepository(): err != nil: %v", err)
		}
	})

	t.Run("success(NewVaultGoogleSecretManagerRepository)", func(t *testing.T) {
		singletonGoogleSecretManagerClientAtomic = 1
		if _, err := NewVaultGoogleSecretManagerRepository(context.Background()); err != nil {
			t.Errorf("err != nil: %v", err)
		}
	})
}

var _ googleSecretManagerClient = (*mockGSMClient)(nil)

type mockGSMClient struct {
	GetSecretSecret               *secretmanagerpb.Secret
	GetSecretErr                  error
	GetSecretVersionSecretVersion *secretmanagerpb.SecretVersion
	GetSecretVersionErr           error
	CreateSecretSecret            *secretmanagerpb.Secret
	CreateSecretErr               error
	UpdateSecretResource          *secretmanagerpb.Secret
	UpdateSecretErr               error
	AccessSecretVersionName       string
	AccessSecretVersionData       []byte
	AccessSecretVersionErr        error
	AddSecretVersionSecretVersion *secretmanagerpb.SecretVersion
	AddSecretVersionErr           error
}

func (m *mockGSMClient) GetSecret(ctx context.Context, req *secretmanagerpb.GetSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error) {
	return m.GetSecretSecret, m.GetSecretErr
}

func (m *mockGSMClient) GetSecretVersion(ctx context.Context, req *secretmanagerpb.GetSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	return m.GetSecretVersionSecretVersion, m.GetSecretVersionErr
}

func (m *mockGSMClient) CreateSecret(ctx context.Context, req *secretmanagerpb.CreateSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error) {
	return m.CreateSecretSecret, m.CreateSecretErr
}

func (m *mockGSMClient) UpdateSecret(ctx context.Context, req *secretmanagerpb.UpdateSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error) {
	return m.UpdateSecretResource, m.UpdateSecretErr
}

func (m *mockGSMClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return &secretmanagerpb.AccessSecretVersionResponse{
		Name: m.AccessSecretVersionName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: m.AccessSecretVersionData,
		},
	}, m.AccessSecretVersionErr
}

func (m *mockGSMClient) AddSecretVersion(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error) {
	return m.AddSecretVersionSecretVersion, m.AddSecretVersionErr
}

func Test_vaultGoogleSecretManagerRepository_ExistsVault(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx           context.Context
		vaultResource string
	}
	tests := []struct {
		name       string
		client     googleSecretManagerClient
		args       args
		wantExists bool
		wantErr    bool
	}{
		{"success(exists)", &mockGSMClient{GetSecretSecret: &secretmanagerpb.Secret{Name: "test"}}, args{context.Background(), "projects/test/secrets/test"}, true, false},
		{"success(!exists)", &mockGSMClient{GetSecretErr: status.Error(codes.NotFound, "")}, args{context.Background(), "projects/test/secrets/test"}, false, false},
		{"failure()", &mockGSMClient{GetSecretErr: fixture.ErrTestError}, args{context.Background(), "test"}, false, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &vaultGoogleSecretManagerRepository{
				client: tt.client,
			}
			gotExists, _, err := repo.GetVaultIfExists(tt.args.ctx, tt.args.vaultResource)
			if (err != nil) != tt.wantErr {
				t.Errorf("vaultGoogleSecretManagerRepository.ExistsVault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExists != tt.wantExists {
				t.Errorf("vaultGoogleSecretManagerRepository.ExistsVault() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

func Test_vaultGoogleSecretManagerRepository_CreateVaultIfNotExists(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx           context.Context
		vaultResource string
	}
	tests := []struct {
		name    string
		client  googleSecretManagerClient
		args    args
		wantErr bool
	}{
		{"success(exists)", &mockGSMClient{GetSecretSecret: &secretmanagerpb.Secret{Name: "test"}}, args{context.Background(), "projects/test/secrets/test"}, false},
		{"success(!exists)", &mockGSMClient{GetSecretErr: status.Error(codes.NotFound, "")}, args{context.Background(), "projects/test/secrets/test"}, false},
		{"failure(ErrSecretInvalidSecretResource)", &mockGSMClient{GetSecretErr: status.Error(codes.NotFound, "")}, args{context.Background(), "invalidSecretName"}, true},
		{"failure(repo.client.CreateSecret)", &mockGSMClient{GetSecretErr: status.Error(codes.NotFound, ""), CreateSecretErr: fixture.ErrTestError}, args{context.Background(), "projects/test/secrets/test"}, true},
		{"failure(repo.ExistsVault)", &mockGSMClient{GetSecretErr: fixture.ErrTestError}, args{context.Background(), "projects/test/secrets/test"}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &vaultGoogleSecretManagerRepository{
				client: tt.client,
			}
			if err := repo.CreateVaultIfNotExists(tt.args.ctx, tt.args.vaultResource); (err != nil) != tt.wantErr {
				t.Errorf("vaultGoogleSecretManagerRepository.CreateVaultIfNotExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_vaultGoogleSecretManagerRepository_LockVault(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                 context.Context
		targetVaultResource string
	}
	tests := []struct {
		name    string
		client  googleSecretManagerClient
		args    args
		wantErr bool
	}{
		{"success()", &mockGSMClient{GetSecretSecret: &secretmanagerpb.Secret{Labels: nil}, UpdateSecretResource: &secretmanagerpb.Secret{}}, args{context.Background(), "test"}, false},
		{"failure(GetSecret)", &mockGSMClient{GetSecretErr: fixture.ErrTestError}, args{context.Background(), "test"}, true},
		{"failure(ErrFailedToAcquireLock)", &mockGSMClient{GetSecretSecret: &secretmanagerpb.Secret{Labels: map[string]string{config.AppName + "-lock": "true"}}}, args{context.Background(), "test"}, true},
		{"failure(UpdateSecret)", &mockGSMClient{GetSecretSecret: &secretmanagerpb.Secret{Labels: map[string]string{}}, UpdateSecretErr: fixture.ErrTestError}, args{context.Background(), "test"}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &vaultGoogleSecretManagerRepository{
				client: tt.client,
			}
			if err := repo.LockVault(tt.args.ctx, tt.args.targetVaultResource); (err != nil) != tt.wantErr {
				t.Errorf("vaultGoogleSecretManagerRepository.LockVault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_vaultGoogleSecretManagerRepository_UnlockVault(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                 context.Context
		targetVaultResource string
	}
	tests := []struct {
		name    string
		client  googleSecretManagerClient
		args    args
		wantErr bool
	}{
		{"success()", &mockGSMClient{GetSecretSecret: &secretmanagerpb.Secret{Labels: nil}, UpdateSecretResource: &secretmanagerpb.Secret{}}, args{context.Background(), "test"}, false},
		{"failure(GetSecret)", &mockGSMClient{GetSecretErr: fixture.ErrTestError}, args{context.Background(), "test"}, true},
		{"failure(UpdateSecret)", &mockGSMClient{GetSecretSecret: &secretmanagerpb.Secret{Labels: map[string]string{}}, UpdateSecretErr: fixture.ErrTestError}, args{context.Background(), "test"}, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &vaultGoogleSecretManagerRepository{
				client: tt.client,
			}
			if err := repo.UnlockVault(tt.args.ctx, tt.args.targetVaultResource); (err != nil) != tt.wantErr {
				t.Errorf("vaultGoogleSecretManagerRepository.LockVault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_vaultGoogleSecretManagerRepository_GetVaultVersionIfExists(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx                  context.Context
		vaultVersionResource string
	}
	tests := []struct {
		name                     string
		client                   googleSecretManagerClient
		args                     args
		wantExists               bool
		wantVaultVersionResource string
		wantErr                  bool
	}{
		{"success()", &mockGSMClient{GetSecretVersionSecretVersion: &secretmanagerpb.SecretVersion{Name: "test"}}, args{context.TODO(), "test"}, true, "test", false},
		{"failure()", &mockGSMClient{GetSecretVersionErr: status.Error(codes.NotFound, "")}, args{context.TODO(), "test"}, false, "", false},
		{"failure()", &mockGSMClient{GetSecretVersionErr: status.Error(codes.Internal, "")}, args{context.TODO(), "test"}, false, "", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &vaultGoogleSecretManagerRepository{
				client: tt.client,
			}
			gotExists, gotVaultVersionResource, err := repo.GetVaultVersionIfExists(tt.args.ctx, tt.args.vaultVersionResource)
			if (err != nil) != tt.wantErr {
				t.Errorf("vaultGoogleSecretManagerRepository.GetVaultVersionResourceIfExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExists != tt.wantExists {
				t.Errorf("vaultGoogleSecretManagerRepository.GetVaultVersionResourceIfExists() gotExists = %v, want %v", gotExists, tt.wantExists)
			}
			if gotVaultVersionResource != tt.wantVaultVersionResource {
				t.Errorf("vaultGoogleSecretManagerRepository.GetVaultVersionResourceIfExists() gotVaultVersionResource = %v, want %v", gotVaultVersionResource, tt.wantVaultVersionResource)
			}
		})
	}
}

func Test_vaultGoogleSecretManagerRepository_GetVaultVersionDataIfExists(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx          context.Context
		resourceName string
	}
	tests := []struct {
		name                     string
		client                   googleSecretManagerClient
		args                     args
		wantExists               bool
		wantVaultVersionResource string
		wantData                 []byte
		wantErr                  bool
	}{
		{"success(exists)", &mockGSMClient{AccessSecretVersionName: "test", AccessSecretVersionData: []byte("ok")}, args{context.Background(), "test"}, true, "test", []byte("ok"), false},
		{"success(!exists)", &mockGSMClient{AccessSecretVersionErr: status.Error(codes.NotFound, codes.NotFound.String())}, args{context.Background(), "test"}, false, "", nil, false},
		{"failure()", &mockGSMClient{AccessSecretVersionErr: status.Error(codes.Internal, codes.Internal.String())}, args{context.Background(), "test"}, false, "", nil, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := vaultGoogleSecretManagerRepository{tt.client}
			exists, gotVaultVersionResource, gotData, err := repo.GetVaultVersionDataIfExists(tt.args.ctx, tt.args.resourceName)
			if exists != tt.wantExists {
				t.Errorf("VaultGoogleSecretManagerRepository.GetVaultVersionDataIfExists() exists = %v, wantExists %v", exists, tt.wantExists)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("VaultGoogleSecretManagerRepository.GetVaultVersionDataIfExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("VaultGoogleSecretManagerRepository.GetVaultVersionDataIfExists() gotData = %v, want %v", string(gotData), string(tt.wantData))
			}
			if !reflect.DeepEqual(gotVaultVersionResource, tt.wantVaultVersionResource) {
				t.Errorf("VaultGoogleSecretManagerRepository.GetVaultVersionDataIfExists() gotVaultVersionResource = %v, want %v", gotVaultVersionResource, tt.wantVaultVersionResource)
			}
		})
	}
}

func Test_vaultGoogleSecretManagerRepository_AddVaultVersion(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx           context.Context
		vaultResource string
		data          []byte
	}
	tests := []struct {
		name              string
		client            googleSecretManagerClient
		args              args
		wantSecretVersion string
		wantErr           bool
	}{
		{"success()", &mockGSMClient{AddSecretVersionSecretVersion: &secretmanagerpb.SecretVersion{Name: "test"}}, args{context.Background(), "test", []byte("test")}, "test", false},
		{"failure()", &mockGSMClient{AddSecretVersionErr: fixture.ErrTestError}, args{context.Background(), "test", []byte("test")}, "test", true},
		{"failure()", &mockGSMClient{AddSecretVersionErr: status.Error(codes.NotFound, codes.NotFound.String())}, args{context.Background(), "test", []byte("test")}, "test", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := &vaultGoogleSecretManagerRepository{
				client: tt.client,
			}
			_, err := repo.AddVaultVersion(tt.args.ctx, tt.args.vaultResource, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("vaultGoogleSecretManagerRepository.CreateVaultVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
