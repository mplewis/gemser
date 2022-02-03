package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"text/template"

	"github.com/a-h/gemini"
)

type Params = map[string]string
type RouteFunction func(w gemini.ResponseWriter, r RequestParams)

type User struct {
	Certificate     *x509.Certificate
	CommonName      string
	CertificateHash string
}

type Router struct {
	routes []Route
}

type RequestParams struct {
	Req    *gemini.Request
	Params Params
	User   *User
}

type RouterMatch struct {
	Router Router
	Params Params
}

type RoutePart struct {
	Key     string
	Matcher bool
}

type Route struct {
	parts   []RoutePart
	handler RouteFunction
}

func NewRouter(routes ...Route) Router {
	return Router{routes}
}

func NewRoute(path string, fn RouteFunction) Route {
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
	return Route{parts, fn}
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

func (r Router) ServeGemini(w gemini.ResponseWriter, rq *gemini.Request) {
	for _, route := range r.routes {
		params, match := route.Match(rq.URL.Path)
		if !match {
			continue
		}

		user, err := getUser(rq.Certificate)
		if err != nil {
			log.Println(err)
			w.SetHeader(gemini.CodeTemporaryFailure, "internal server error")
			return
		}
		route.handler(w, RequestParams{
			User:   user,
			Params: params,
			Req:    rq,
		})
		return
	}

	w.SetHeader(gemini.CodeNotFound, "path not found")
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

func name(u *User) string {
	if u == nil {
		return "anonymous"
	}
	return u.CommonName
}

func render(w io.Writer, templateName string, data interface{}) error {
	path := fmt.Sprintf("templates/%s.md", templateName)
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	parsed, err := template.New(templateName).Parse(string(raw))
	if err != nil {
		return err
	}
	err = parsed.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}

func renderFunc(templateName string, handler func(RequestParams) interface{}) RouteFunction {
	return func(w gemini.ResponseWriter, r RequestParams) {
		data := handler(r)
		err := render(w, templateName, data)
		if err != nil {
			log.Println(err)
			w.SetHeader(gemini.CodeTemporaryFailure, "internal server error")
			return
		}
	}
}

func main() {
	cert, err := tls.LoadX509KeyPair("localhost.crt", "localhost.key")
	if err != nil {
		log.Fatal(err)
	}

	router := NewRouter(
		NewRoute("/", renderFunc("home", func(r RequestParams) interface{} {
			return map[string]string{"Name": name(r.User)}
		})),

		NewRoute("/foo/:bar", renderFunc("foo", func(r RequestParams) interface{} {
			return map[string]string{"Name": name(r.User), "Bar": r.Params["bar"]}
		})),
	)

	domain := gemini.NewDomainHandler("localhost", cert, router)
	err = gemini.ListenAndServe(context.Background(), ":1965", domain)
	if err != nil {
		log.Fatal("error:", err)
	}
}
