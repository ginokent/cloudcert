package mock

import (
	"context"
	"crypto"

	"github.com/ginokent/cloudacme/repository"
)

var _ repository.LetsEncryptRepository = (*LetsEncryptRepository)(nil)

type LetsEncryptRepository struct {
	IssueCertificatePrivateKey        []byte
	IssueCertificateCertificate       []byte
	IssueCertificateIssuerCertificate []byte
	IssueCertificateCSR               []byte
	IssueCertificateErr               error
}

func (m *LetsEncryptRepository) IssueCertificate(ctx context.Context, acmeAccountKey crypto.PrivateKey, privateKey crypto.PrivateKey, domains []string) (privateKeyPEM, certificatePEM, issuerCertificate, csr []byte, err error) {
	return m.IssueCertificatePrivateKey, m.IssueCertificateCertificate, m.IssueCertificateIssuerCertificate, m.IssueCertificateCSR, m.IssueCertificateErr
}
