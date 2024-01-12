package server

import (
	"fmt"
	"testing"
)

func TestGetHandlePath(t *testing.T) {
	testData := []struct {
		test string
		exp  string
	}{
		{
			"/foo/bar",
			"/foo/bar",
		},
		{
			"/foo/bar/",
			"/foo/bar",
		},
		{
			"/foo/bar///",
			"/foo/bar",
		},
		{
			"/foo/{bar}",
			"/foo/",
		},
		{
			"/foo/{bar}/",
			"/foo/",
		},
		{
			"/",
			"/",
		},
		{
			"/foo/bar/{yolo}/{biff}/",
			"/foo/bar/",
		},
	}
	for i, td := range testData {
		tested := getHandlePath(td.test)
		if tested != td.exp {
			t.Errorf(
				"index: %d, handle path mismatch, exp: '%s', got: '%s', for: '%s'",
				i, td.exp, tested, td.test,
			)
		}
	}
}

func TestBuildDynamicParameters(t *testing.T) {
	testdata := []struct {
		handler *myHandler
		path    string
		exp     map[string]string
		isError bool
		msg     string
	}{
		{
			handler: &myHandler{
				dynamicPaths:     []string{},
				dynamicPathIndex: 0,
				route:            &route{},
			},
			path:    "/baz",
			exp:     nil,
			isError: false,
			msg:     "no dynamic parameters",
		},
		{
			handler: &myHandler{
				dynamicPaths: []string{
					"foo",
				},
				dynamicPathIndex: 0,
				route: &route{
					params: routeParameterMap{
						//none of this stuff really matters for this particular test
						"foo": &routeParameter{},
					},
				},
			},
			path: "/baz",
			exp: map[string]string{
				"foo": "baz",
			},
			isError: false,
			msg:     "simple dynamic pattern",
		},
		{
			handler: &myHandler{
				dynamicPaths: []string{
					"foo", "bar",
				},
				dynamicPathIndex: 1,
				route: &route{
					params: routeParameterMap{
						//none of this stuff really matters for this particular test
						"foo": &routeParameter{},
						"bar": &routeParameter{},
					},
				},
			},
			path: "/static/dynamicFoo/dynamicBar",
			exp: map[string]string{
				"foo": "dynamicFoo",
				"bar": "dynamicBar",
			},
			isError: false,
			msg:     "2 dynamic vars",
		},
		{
			handler: &myHandler{
				dynamicPaths: []string{
					"foo", "bar",
				},
				dynamicPathIndex: 1,
				route: &route{
					params: routeParameterMap{
						//none of this stuff really matters for this particular test
						"foo": &routeParameter{},
						"bar": &routeParameter{},
					},
				},
			},
			path: "/static/dynamicFoo",
			exp: map[string]string{
				"foo": "dynamicFoo",
			},
			isError: false,
			msg:     "first dynamic var set only",
		},
		{
			handler: &myHandler{
				dynamicPaths: []string{
					"foo", "bar",
				},
				dynamicPathIndex: 5,
				route: &route{
					params: routeParameterMap{
						//none of this stuff really matters for this particular test
						"foo": &routeParameter{},
						"bar": &routeParameter{},
					},
				},
			},
			path:    "/static/dynamicFoo",
			exp:     nil,
			isError: true,
			msg:     "dynamic path index greater than path length",
		},
		{
			handler: &myHandler{
				dynamicPaths: []string{
					"foo", "bar",
				},
				dynamicPathIndex: 1,
				route: &route{
					params: routeParameterMap{
						//none of this stuff really matters for this particular test
						"foo": &routeParameter{},
						"baz": &routeParameter{},
					},
				},
			},
			path:    "/static/dynamicFoo/dynamicBaz",
			exp:     nil,
			isError: true,
			msg:     "dynamic paths don't match params",
		},
	}
	doError := func(i int, msg, fmtMsg string, args ...interface{}) {
		t.Errorf("index: %d, %s, msg: '%s'", i, fmt.Sprintf(fmtMsg, args...), msg)
	}
	for i, td := range testdata {
		params, err := td.handler.buildDynamicParameters(td.path)
		if td.isError {
			if err == nil {
				doError(i, td.msg, "should have recieved an error")
			}
			continue
		}
		if err != nil {
			doError(i, td.msg, "should not have received an error, got: '%s'", err)
			continue
		}
		if len(td.exp) != len(params) {
			doError(
				i, td.msg, "parameter maps don't match, exp: %d, got: %d",
				len(td.exp), len(params),
			)
			continue
		}
		var found string
		var ok bool
		for k, v := range td.exp {
			if found, ok = params[k]; !ok {
				doError(i, td.msg, "returned parameters did not contain: '%s'", k)
				continue
			}
			if v != found {
				doError(
					i, td.msg,
					"found parameters don't match, key: '%s', exp: '%s', got: '%s'",
					k, v, found,
				)
			}
		}
	}
}
