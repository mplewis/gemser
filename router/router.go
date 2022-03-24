package router

import (
	"log"

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

type RoutePart struct {
	Key     string
	Matcher bool
}

type RequestParams struct {
	Req    *gemini.Request
	Params Params
	User   *user.User
}

type Params = map[string]string

func NewRouter(routes ...Route) Router {
	return Router{routes}
}

func (r Router) ServeGemini(w gemini.ResponseWriter, rq *gemini.Request) {
	for _, route := range r.routes {
		params, match := route.Match(rq.URL.Path)
		if !match {
			continue
		}

		user, err := user.Get(rq.Certificate)
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
