package mock

import (
	"context"

	"github.com/newtstat/cloudacme/repository"
)

var _ repository.LetsEncryptRepository = (*LetsEncryptRepository)(nil)

type LetsEncryptRepository struct {
	IssueCertificatePrivateKey        []byte
	IssueCertificateCertificate       []byte
	IssueCertificateIssuerCertificate []byte
	IssueCertificateCSR               []byte
	IssueCertificateErr               error
}

func (m *LetsEncryptRepository) IssueCertificate(ctx context.Context, domains []string) (privateKey, certificate, issuerCertificate, csr []byte, err error) {
	return m.IssueCertificatePrivateKey, m.IssueCertificateCertificate, m.IssueCertificateIssuerCertificate, m.IssueCertificateCSR, m.IssueCertificateErr
}
