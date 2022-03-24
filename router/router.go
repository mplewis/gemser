package router

import (
	"log"
	"net/url"

	"github.com/a-h/gemini"
	"github.com/mplewis/gemser/user"
)

type Router struct {
	routes []Route
}

type RouterMatch struct {
	Router Router
	Params Params
}

type Request struct {
	Raw         *gemini.Request
	PathParams  Params
	QueryParams Params
	QueryString string
	User        *user.User
}

type Params = map[string]string

func NewRouter(routes ...Route) Router {
	return Router{routes}
}

func (r Router) ServeGemini(w gemini.ResponseWriter, rq *gemini.Request) {
	for _, route := range r.routes {
		pathParams, match := route.Match(rq.URL.Path)
		if !match {
			continue
		}

		queryParams := flattenQueryParams(rq.URL.Query())
		user, err := user.Get(rq.Certificate)
		if err != nil {
			log.Println(err)
			w.SetHeader(gemini.CodeTemporaryFailure, "internal server error")
			return
		}
		route.handler(w, Request{
			User:        user,
			PathParams:  pathParams,
			QueryParams: queryParams,
			QueryString: rq.URL.RawQuery,
			Raw:         rq,
		})
		return
	}

	w.SetHeader(gemini.CodeNotFound, "path not found")
}

func flattenQueryParams(raw url.Values) Params {
	params := Params{}
	for k, v := range raw {
		params[k] = v[0]
	}
	return params
}
