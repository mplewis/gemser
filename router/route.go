package router

import (
	"strings"

	"github.com/a-h/gemini"
)

type RouteFunction func(w gemini.ResponseWriter, r Request)

type Route struct {
	parts   []RoutePart
	handler RouteFunction
}

func NewRoute(pattern string, fn RouteFunction) Route {
	raws := strings.Split(strings.Trim(pattern, "/"), "/")
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
	raws := strings.Split(strings.TrimLeft(path, "/"), "/")
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
