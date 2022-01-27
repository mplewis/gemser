package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/a-h/gemini"
	"github.com/a-h/gemini/mux"
)

type User struct {
	Certificate     *x509.Certificate
	CommonName      string
	CertificateHash string
}

func hash(data string) string {
	sum := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func getUser(c gemini.Certificate) (*User, error) {
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

func main() {
	router := mux.NewMux()
	router.AddRoute("/", gemini.HandlerFunc(func(w gemini.ResponseWriter, r *gemini.Request) {
		user, err := getUser(r.Certificate)
		if err != nil {
			log.Panic(err)
		}

		if user == nil {
			w.Write([]byte("Not logged in"))
			return
		}

		msg := fmt.Sprintf("%s => %s", user.CommonName, user.CertificateHash)
		w.Write([]byte(msg))
	}))

	cert, err := tls.LoadX509KeyPair("localhost.crt", "localhost.key")
	if err != nil {
		log.Fatal(err)
	}

	domain := gemini.NewDomainHandler("localhost", cert, router)
	err = gemini.ListenAndServe(context.Background(), ":1965", domain)
	if err != nil {
		log.Fatal("error:", err)
	}
}
