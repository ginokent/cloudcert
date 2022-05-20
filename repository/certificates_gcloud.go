package repository

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/cockroachdb/errors"
	"github.com/googleapis/gax-go"
	"github.com/newtstat/cloudacme/config"
	"github.com/newtstat/cloudacme/trace"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var _ googleSecretManagerClient = (*secretmanager.Client)(nil)

type googleSecretManagerClient interface {
	GetSecret(ctx context.Context, req *secretmanagerpb.GetSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error)
	CreateSecret(ctx context.Context, req *secretmanagerpb.CreateSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error)
	UpdateSecret(ctx context.Context, req *secretmanagerpb.UpdateSecretRequest, opts ...gax.CallOption) (*secretmanagerpb.Secret, error)
	GetSecretVersion(context.Context, *secretmanagerpb.GetSecretVersionRequest, ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	AddSecretVersion(ctx context.Context, req *secretmanagerpb.AddSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.SecretVersion, error)
}

var (
	ErrSingletonGoogleSecretManagerClientNotInitialized = errors.New("singleton google secret manager client not initialized")
	ErrFailedToAcquireLock                              = errors.New("failed to acquire lock")
)

type vaultGoogleSecretManagerRepository struct {
	client googleSecretManagerClient
}

var (
	singletonGoogleSecretManagerClient       googleSecretManagerClient // nolint: gochecknoglobals
	singletonGoogleSecretManagerClientAtomic uint32                    // nolint: gochecknoglobals
	singletonGoogleSecretManagerClientMu     = &sync.Mutex{}           // nolint: gochecknoglobals
)

func initSingletonGoogleSecretManagerClient(newGoogleSecretManagerClient func(ctx context.Context, opts ...option.ClientOption) (*secretmanager.Client, error), opts ...option.ClientOption) (err error) {
	if atomic.LoadUint32(&singletonGoogleSecretManagerClientAtomic) == 0 {
		singletonGoogleSecretManagerClientMu.Lock()
		defer singletonGoogleSecretManagerClientMu.Unlock()
		if singletonGoogleSecretManagerClientAtomic == 0 {
			client, err := newGoogleSecretManagerClient(context.Background())
			if err != nil {
				return errors.Errorf("secretmanager.NewClient: %w", err)
			}
			singletonGoogleSecretManagerClient = client
			atomic.StoreUint32(&singletonGoogleSecretManagerClientAtomic, 1)
		}
	}

	return nil
}

func NewVaultGoogleSecretManagerRepository(ctx context.Context) (VaultRepository, error) {
	_, span := trace.Start(ctx, "repository.NewVaultGoogleSecretManagerRepository")
	defer span.End()

	return newVaultGoogleSecretManagerRepository(ctx, secretmanager.NewClient)
}

func newVaultGoogleSecretManagerRepository(ctx context.Context, newGoogleSecretManagerClient func(ctx context.Context, opts ...option.ClientOption) (*secretmanager.Client, error)) (VaultRepository, error) {
	if err := initSingletonGoogleSecretManagerClient(newGoogleSecretManagerClient); err != nil {
		return nil, errors.Errorf("onceInitVaultGoogleSecretManagerRepository: %w", err)
	}

	return &vaultGoogleSecretManagerRepository{
		client: singletonGoogleSecretManagerClient,
	}, nil
}

func (repo *vaultGoogleSecretManagerRepository) GetVaultIfExists(ctx context.Context, targetVaultResource string) (exists bool, vaultResource string, err error) {
	ctx, span := trace.Start(ctx, "(*repository.vaultGoogleSecretManagerRepository).ExistsVault")
	defer span.End()

	req := &secretmanagerpb.GetSecretRequest{
		Name: targetVaultResource,
	}

	if err := trace.StartFunc(ctx, "(*repository.vaultGoogleSecretManagerRepository).client.GetSecret")(func(child context.Context) (err error) {
		var v *secretmanagerpb.Secret
		v, err = repo.client.GetSecret(child, req)
		if err == nil {
			exists = true
			vaultResource = v.GetName()
			return nil
		}
		if stat, _ := status.FromError(err); stat.Code() == codes.NotFound {
			exists = false
			vaultResource = ""
			return nil
		}
		return
	}); err != nil {
		return false, "", errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.GetSecret: %w", err)
	}

	span.SetAttributes(attribute.Bool("exists", exists))
	span.SetAttributes(attribute.String("vaultResource", targetVaultResource))

	return exists, vaultResource, nil
}

func (repo *vaultGoogleSecretManagerRepository) CreateVaultIfNotExists(ctx context.Context, vaultResource string) (err error) {
	ctx, span := trace.Start(ctx, "(*repository.vaultGoogleSecretManagerRepository).CreateVaultIfNotExists")
	defer span.End()

	exists, _, err := repo.GetVaultIfExists(ctx, vaultResource)
	if err != nil {
		return errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).ExistsVault: %w", err)
	}

	if exists {
		return nil
	}

	secretParent, vaultID, found := strings.Cut(vaultResource, "/secrets/")
	if !found {
		return errors.Errorf("secretResource=%s: %w", vaultResource, ErrInvalidVaultResource)
	}

	req := &secretmanagerpb.CreateSecretRequest{
		Parent:   secretParent,
		SecretId: vaultID,
		Secret: &secretmanagerpb.Secret{
			Name: vaultResource,
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{},
			},
		},
	}

	if err := trace.StartFunc(ctx, "(*repository.vaultGoogleSecretManagerRepository).client.CreateSecret")(func(child context.Context) (err error) {
		_, err = repo.client.CreateSecret(child, req)
		return
	}); err != nil {
		return errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.CreateSecret: %w", err)
	}

	span.SetAttributes(attribute.String("vaultResource", vaultResource))

	return nil
}

func (repo *vaultGoogleSecretManagerRepository) LockVault(ctx context.Context, targetVaultResource string) (err error) {
	ctx, span := trace.Start(ctx, "(*repository.vaultGoogleSecretManagerRepository).LockVault")
	defer span.End()

	const (
		labelKey   = config.AppName + "-lock"
		labelValue = "true"
	)

	if err := repo.CreateVaultIfNotExists(ctx, targetVaultResource); err != nil {
		return errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).CreateVaultIfNotExists: %w", err)
	}

	var secret *secretmanagerpb.Secret
	if err := trace.StartFunc(ctx, "(*repository.vaultGoogleSecretManagerRepository).client.GetSecret")(func(child context.Context) (err error) {
		secret, err = repo.client.GetSecret(child, &secretmanagerpb.GetSecretRequest{
			Name: targetVaultResource,
		})
		return
	}); err != nil {
		return errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.GetSecret: %w", err)
	}

	if secret.Labels == nil {
		secret.Labels = map[string]string{}
	}

	for k, v := range secret.Labels {
		if k == labelKey {
			if v == labelValue {
				// nolint: wrapcheck
				return ErrFailedToAcquireLock
			}
		}
	}

	secret.Labels[labelKey] = labelValue

	if err := trace.StartFunc(ctx, "(*repository.vaultGoogleSecretManagerRepository).client.UpdateSecret")(func(child context.Context) (err error) {
		_, err = repo.client.UpdateSecret(child, &secretmanagerpb.UpdateSecretRequest{
			Secret: secret,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"labels"},
			},
		})
		return
	}); err != nil {
		return errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.UpdateSecret: %w", err)
	}

	return nil
}

func (repo *vaultGoogleSecretManagerRepository) UnlockVault(ctx context.Context, targetVaultResource string) (err error) {
	ctx, span := trace.Start(ctx, "(*repository.vaultGoogleSecretManagerRepository).UnlockVault")
	defer span.End()

	const (
		labelKey   = config.AppName + "-lock"
		labelValue = "false"
	)

	if err := repo.CreateVaultIfNotExists(ctx, targetVaultResource); err != nil {
		return errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).CreateVaultIfNotExists: %w", err)
	}

	var secret *secretmanagerpb.Secret
	if err := trace.StartFunc(ctx, "(*repository.vaultGoogleSecretManagerRepository).client.GetSecret")(func(child context.Context) (err error) {
		secret, err = repo.client.GetSecret(child, &secretmanagerpb.GetSecretRequest{
			Name: targetVaultResource,
		})
		return
	}); err != nil {
		return errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.GetSecret: %w", err)
	}

	if secret.Labels == nil {
		secret.Labels = map[string]string{}
	}

	secret.Labels[labelKey] = labelValue

	if err := trace.StartFunc(ctx, "(*repository.vaultGoogleSecretManagerRepository).client.UpdateSecret")(func(child context.Context) (err error) {
		_, err = repo.client.UpdateSecret(child, &secretmanagerpb.UpdateSecretRequest{
			Secret: secret,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"labels"},
			},
		})
		return
	}); err != nil {
		return errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.UpdateSecret: %w", err)
	}

	return nil
}

func (repo *vaultGoogleSecretManagerRepository) GetVaultVersionIfExists(ctx context.Context, vaultVersionResource string) (exists bool, version string, err error) {
	ctx, span := trace.Start(ctx, "(*repository.vaultGoogleSecretManagerRepository).GetVaultVersionIfExists")
	defer span.End()

	req := &secretmanagerpb.GetSecretVersionRequest{
		Name: vaultVersionResource,
	}

	if err := trace.StartFunc(ctx, "(*repository.vaultGoogleSecretManagerRepository).client.GetSecretVersion")(func(child context.Context) (err error) {
		var v *secretmanagerpb.SecretVersion
		v, err = repo.client.GetSecretVersion(child, req)
		if err == nil {
			exists = true
			version = v.Name
			return nil
		}
		if stat, _ := status.FromError(err); stat.Code() == codes.NotFound {
			exists = false
			version = ""
			return nil
		}
		return
	}); err != nil {
		return false, "", errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.GetSecretVersion: %w", err)
	}

	return exists, version, nil
}

func (repo *vaultGoogleSecretManagerRepository) GetVaultVersionDataIfExists(ctx context.Context, targetVaultVersionResource string) (exists bool, vaultVersionResource string, data []byte, err error) {
	ctx, span := trace.Start(ctx, "(*repository.vaultGoogleSecretManagerRepository).GetVaultVersionDataIfExists")
	defer span.End()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: targetVaultVersionResource,
	}

	if err := trace.StartFunc(ctx, "(*repository.vaultGoogleSecretManagerRepository).client.AccessSecretVersion")(func(child context.Context) (err error) {
		var secretVersion *secretmanagerpb.AccessSecretVersionResponse
		secretVersion, err = repo.client.AccessSecretVersion(child, req)
		if err == nil {
			exists = true
			vaultVersionResource = secretVersion.Name
			data = secretVersion.Payload.GetData()
			return nil
		}
		if stat, _ := status.FromError(err); stat.Code() == codes.NotFound {
			exists = false
			vaultVersionResource = ""
			data = nil
			return nil // NOTE: if NotFound, deal as not error.
		}
		return
	}); err != nil {
		return false, "", nil, errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.AccessSecretVersion: %w", err)
	}

	span.SetAttributes(attribute.String("vaultVersionResource", vaultVersionResource))

	return exists, vaultVersionResource, data, nil
}

func (repo *vaultGoogleSecretManagerRepository) AddVaultVersion(ctx context.Context, vaultResource string, data []byte) (vaultVersionResource string, err error) {
	ctx, span := trace.Start(ctx, "(*repository.vaultGoogleSecretManagerRepository).AddVaultVersion")
	defer span.End()
	span.SetAttributes(attribute.String("vaultResource", vaultResource))

	req := &secretmanagerpb.AddSecretVersionRequest{
		Parent: vaultResource,
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	}

	var secretVersion *secretmanagerpb.SecretVersion
	if err := trace.StartFunc(ctx, "(*repository.vaultGoogleSecretManagerRepository).client.AddSecretVersion")(func(child context.Context) (err error) {
		secretVersion, err = repo.client.AddSecretVersion(child, req)
		if stat, _ := status.FromError(err); stat.Code() == codes.NotFound {
			return errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.AddSecretVersion: %w", err)
		}
		return
	}); err != nil {
		return "", errors.Errorf("(*repository.vaultGoogleSecretManagerRepository).client.AddSecretVersion: %w", err)
	}

	span.SetAttributes(attribute.String("vaultVersionResource", secretVersion.Name))

	return secretVersion.Name, nil
}
