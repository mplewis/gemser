package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"text/template"

	"github.com/a-h/gemini"
	"github.com/mplewis/gemser/router"
	"github.com/mplewis/gemser/user"
)

func name(u *user.User) string {
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

func renderFunc(templateName string, handler func(router.Request) interface{}) router.RouteFunction {
	return func(w gemini.ResponseWriter, r router.Request) {
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

	router := router.NewRouter(
		router.NewRoute("/", renderFunc("home", func(r router.Request) interface{} {
			return map[string]string{"Name": name(r.User)}
		})),

		router.NewRoute("/foo/:bar", renderFunc("foo", func(r router.Request) interface{} {
			return map[string]string{"Name": name(r.User), "Bar": r.PathParams["bar"]}
		})),

		router.NewRoute("/greet", func(w gemini.ResponseWriter, r router.Request) {

		}),

		router.NewRoute("/input/:next", func(w gemini.ResponseWriter, r router.Request) {
			if r.QueryString != "" {
				w.SetHeader(gemini.CodeRedirect, r.PathParams["next"])
				return
			}
			w.SetHeader(gemini.CodeInput, "input requested")
		}),
	)

	domain := gemini.NewDomainHandler("localhost", cert, router)
	err = gemini.ListenAndServe(context.Background(), ":1965", domain)
	if err != nil {
		log.Fatal("error:", err)
	}
}
