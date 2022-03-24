package user

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"

	"github.com/a-h/gemini"
)

type User struct {
	Certificate     *x509.Certificate
	CommonName      string
	CertificateHash string
}

func Get(c gemini.Certificate) (*User, error) {
	if len(c.Key) == 0 {
		return nil, nil
	}

	cert, err := x509.ParseCertificate([]byte(c.Key))
	if err != nil {
		return nil, err
	}

	return &User{
		Certificate:     cert,
		CommonName:      cert.Subject.CommonName,
		CertificateHash: hash(c.ID),
	}, nil
}

func hash(data string) string {
	sum := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(sum[:])
}
