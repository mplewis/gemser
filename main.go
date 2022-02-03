package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/a-h/gemini"
)

type User struct {
	Certificate     *x509.Certificate
	CommonName      string
	CertificateHash string
}

type Handler struct {
	routes []Route
}

type RequestParams struct {
	Path string
	User *User
}

type HandlerMatch struct {
	Handler Handler
	Params  map[string]string
}

type RoutePart struct {
	Key     string
	Matcher bool
}

type Route struct {
	parts []RoutePart
}

func NewHandler(routes ...Route) Handler {
	return Handler{routes}
}

func NewRoute(path string) Route {
	raws := strings.Split(strings.Trim(path, "/"), "/")
	parts := []RoutePart{}
	for _, raw := range raws {
		if strings.HasPrefix(raw, ":") {
			parts = append(parts, RoutePart{
				Key:     raw[1:],
				Matcher: true,
			})
		} else {
			parts = append(parts, RoutePart{
				Key:     raw,
				Matcher: false,
			})
		}
	}
	return Route{parts}
}

func (r Route) Match(path string) (map[string]string, bool) {
	raws := strings.Split(strings.Trim(path, "/"), "/")
	if len(raws) != len(r.parts) {
		return nil, false
	}
	params := map[string]string{}
	for i, a := range raws {
		b := r.parts[i]
		if b.Matcher {
			params[b.Key] = a
		} else if a != b.Key {
			return nil, false
		}
	}
	return params, true
}

func (h Handler) ServeGemini(w gemini.ResponseWriter, r *gemini.Request) {
	log.Println(r.URL.Path)
	log.Println(h.routes)
	for _, route := range h.routes {
		params, ok := route.Match(r.URL.Path)
		if !ok {
			continue
		}
		log.Println(params)
		w.Write([]byte("hello"))
		return
	}

	w.SetHeader(gemini.CodeNotFound, "not found")
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
	cert, err := tls.LoadX509KeyPair("localhost.crt", "localhost.key")
	if err != nil {
		log.Fatal(err)
	}

	h := NewHandler(
		NewRoute("/"),
		NewRoute("/foo/:bar"),
	)

	r := NewRoute("/foo/:bar")
	fmt.Println(r)
	fmt.Println(r.Match("/foo/quux"))

	domain := gemini.NewDomainHandler("localhost", cert, h)
	err = gemini.ListenAndServe(context.Background(), ":1965", domain)
	if err != nil {
		log.Fatal("error:", err)
	}
}
