package repository

import (
	"context"
	"crypto"
	"errors"

	"github.com/go-acme/lego/v4/registration"
)

type LetsEncryptRepository interface {
	IssueCertificate(ctx context.Context, domains []string) (privateKey, certificate, issuerCertificate, csr []byte, err error)
}

var (
	// ErrEmailIsEmpty email is empty.
	ErrEmailIsEmpty = errors.New("email is empty")

	// ErrTermsOfServiceNotAgreed to issue a certificate, you need to agree to the current Let's Encrypt terms of service.
	ErrTermsOfServiceNotAgreed = errors.New("to issue a certificate, you need to agree to the current Let's Encrypt terms of service")

	// ErrDomainsIsNil domains is nil.
	ErrDomainsIsNil = errors.New("domains is nil")
)

// User implements registration.User interface.
type User struct {
	email        string
	registration *registration.Resource
	key          crypto.PrivateKey
}

// GetEmail returns (*User).email.
func (u *User) GetEmail() string {
	return u.email
}

// GetRegistration returns (*User).registration.
func (u User) GetRegistration() *registration.Resource {
	return u.registration
}

// GetPrivateKey returns (*User).key.
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
