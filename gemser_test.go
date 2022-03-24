package main_test

import (
	"fmt"
	"testing"

	"github.com/a-h/gemini"
	"github.com/mplewis/gemser/router"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGemser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gemser Suite")
}

var _ = Describe("gemser", func() {
	Describe("route", func() {
		dummy := func(w gemini.ResponseWriter, r router.Request) {}
		empty := map[string]string{}
		type pattern = string
		type expectation struct {
			path   string
			match  bool
			params router.Params
		}

		specs := map[pattern][]expectation{
			"/": []expectation{
				{path: "/", match: true, params: empty},
				{path: "/foo", match: false},
			},
			"/users/:name": []expectation{
				{path: "/users/mplewis", match: true, params: map[string]string{"name": "mplewis"}},
				{path: "/users/mplewis/", match: false},
				{path: "/", match: false},
			},
		}

		for pattern, expectations := range specs {
			route := router.NewRoute(pattern, dummy)
			for _, expectation := range expectations {
				expectation := expectation
				matchStr := "matches"
				if !expectation.match {
					matchStr = "does not match"
				}
				It(fmt.Sprintf("%s %s -> %s", matchStr, pattern, expectation.path), func() {
					params, success := route.Match(expectation.path)
					Expect(success).To(Equal(expectation.match))
					if expectation.params != nil {
						Expect(params).To(Equal(expectation.params))
					}
				})
			}
		}
	})
})
