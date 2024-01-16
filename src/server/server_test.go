package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

func TestNewHandlerDynamicPaths(t *testing.T) {
	testdata := []struct {
		path            string
		expIndex        uint8
		expDynamicPaths []string
		msg             string
		//don't think there's actually a possibility of errors here...
		isErr bool
	}{
		{
			"/",
			0,
			nil,
			"simple root, non-dynamic",
			false,
		},
		{
			"/{blarg}",
			0,
			[]string{"blarg"},
			"root pattern, 1 dynamic",
			false,
		},
		{
			"/foo/bar/baz",
			0,
			nil,
			"multi path, non-dynamic",
			false,
		},
		{
			"/foo/bar/{baz}",
			2,
			[]string{"baz"},
			"multi path, 1 dynamic",
			false,
		},
		{
			"/foo/{baz}/{biff}",
			1,
			[]string{"baz", "biff"},
			"multi path, 2 dynamic",
			false,
		},
	}
	dummyRoute := &route{}
	for i, td := range testdata {
		handler, err := newHandler(td.path, dummyRoute, nil)
		if td.isErr {
			if err == nil {
				t.Errorf(getTestMessage(i, td.msg, "expected error"))
			}
			continue
		}
		if len(td.expDynamicPaths) != len(handler.dynamicPaths) {
			t.Errorf(
				getTestMessage(
					i, td.msg,
					"dynamic path lengths don't match, exp: %d, got: %d",
					len(td.expDynamicPaths), len(handler.dynamicPaths),
				),
			)
		}
		if len(td.expDynamicPaths) == 0 {
			continue
		}
		for i, subPath := range td.expDynamicPaths {
			if subPath != handler.dynamicPaths[i] {
				t.Errorf(
					getTestMessage(
						i, td.msg, "dynamic subpaths don't match, exp: '%s', got: '%s'",
						subPath, handler.dynamicPaths[i],
					),
				)
			}
		}
		if td.expIndex != handler.dynamicPathIndex {
			t.Errorf(
				getTestMessage(
					i, td.msg,
					"dynamic path index doesn't match, exp: %d, got: %d",
					td.expIndex, handler.dynamicPathIndex,
				),
			)
		}
	}
}

func getTestMessage(i int, msg, fmtMsg string, args ...interface{}) string {
	return fmt.Sprintf(
		"index: %d, %s, msg: '%s'",
		i, fmt.Sprintf(fmtMsg, args...), msg,
	)
}

func TestBuildDynamicParameters(t *testing.T) {
	testdata := []struct {
		handler *myHandler
		path    string
		exp     map[string]string
		isError bool
		msg     string
	}{
		//0
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
		//1
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
		//2
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
		//3
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
		//4
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
		//5
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
	for i, td := range testdata {
		params, err := td.handler.buildDynamicParameters(td.path)
		if td.isError {
			if err == nil {
				t.Errorf(getTestMessage(i, td.msg, "should have recieved an error"))
			}
			continue
		}
		if err != nil {
			t.Errorf(
				getTestMessage(
					i, td.msg, "should not have received an error, got: '%s'", err,
				),
			)
			continue
		}
		if len(td.exp) != len(params) {
			t.Errorf(
				getTestMessage(
					i, td.msg, "parameter maps don't match, exp: %d, got: %d",
					len(td.exp), len(params),
				),
			)
			continue
		}
		var found string
		var ok bool
		for k, v := range td.exp {
			if found, ok = params[k]; !ok {
				t.Errorf(
					getTestMessage(
						i, td.msg, "returned parameters did not contain: '%s'", k,
					),
				)
				continue
			}
			if v != found {
				t.Errorf(
					getTestMessage(
						i, td.msg,
						"found parameters don't match, key: '%s', exp: '%s', got: '%s'",
						k, v, found,
					),
				)
			}
		}
	}
}

type callbackData struct {
	res bool
	err error
}

var callbackDataMap map[string]callbackData = map[string]callbackData{
	"cb1": callbackData{
		true,
		nil,
	},
	"cb2": callbackData{
		false,
		fmt.Errorf("cb2 error"),
	},
	"cb3": callbackData{
		true,
		fmt.Errorf("cb3 error"),
	},
	"cb4": callbackData{
		false,
		nil,
	},
	"cb5": callbackData{
		true,
		nil,
	},
}

func makeCallback(calllist **[]string, name string, res bool, err error) Callback {
	//use the headers to verify params are correct
	return func(params map[string]string, w http.ResponseWriter, r *http.Request) (bool, error) {
		for k, v := range params {
			w.Header().Add(k, v)
		}
		tmp := *(*calllist)
		tmp = append(tmp, name)
		*calllist = &tmp
		return res, err
	}
}

type testPath struct {
	body   *string
	target string
	method string
}

// TODO here,
// just want to call it and see what happens
// parameters SHOULD be encoded in the headers for checking
func TestServer(t *testing.T) {
	var callbacks map[string]Callback
	var calllist *[]string
	for name, cb := range callbackDataMap {
		callbacks[name] = makeCallback(&calllist, name, cb.res, cb.err)
	}
	testData := []struct {
		serverYaml string
		msg        string
		expErr     bool
		testPaths  testPath
	}{
		"testdata/basic.yaml",
		"simplest test",
		false,
		testPath{
			nil,
			"http://example.com&foo=no_idea",
			"GET",
		},
	}
	for i, td := range testData {
		serverYaml, err := os.Open(td.serverYaml)
		if err != nil {
			t.Fatalf(
				getTestMessage(
					i, td.msg, "failed opening test yaml for: '%s' with error: '%s'",
					td.serverYaml, err,
				),
			)
		}
		server, err := NewServer(serverYaml, callbacks)
		serverYaml.Close()
		if td.expErr != nil {
			if err == nil {
				t.Errorf(
					getTestMessage(i, td.msg, "expected error, received none"),
				)
			}
			continue
		}
		for _, testPath := range td.testPaths {
			w := httptest.NewRecorder()
			var body io.Reader = nil
			if testPath.body != nil {
				body = strings.NewReader(*testPath.body)
			}
			r := httptest.NewRequest(testPath.method, testPath.target, body)
			tmpList := make([]string, 0)
			calllist = &tmpList
			server[p].ServeHTTP(w, r)
			//something about params....
			//TODO verify header/params
			//if testPath.params !=
		}
	}
}

/*
type Callback func(
	map[string]string,
	http.ResponseWriter,
	*http.Request,
) (bool, error)
*/

func makeCallback(calllist **[]string) Callback {
	return func(
		params map[string]string, w http.ResponseWriter, r *http.Request,
	) (bool, error) {

	}
}

func foo() {
	var calllist *[]string
	for _, cbName := range callbackNames {

	}
}
