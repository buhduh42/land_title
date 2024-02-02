/*
Should probably generalize the sub slice test data conventions
from util.GetTestMetaData, eg write a testing library
*/
package server

import (
	"fmt"
	"io"
	"landtitle/util"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	logger "github.com/buhduh42/go-logger"
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
		{
			"/{biff}",
			"/",
		},
		{
			"/{foo}/{bar}",
			"/",
		},
	}
	//myLogger = newTestLogger(t, util.Ptr(logger.TRACE))
	for i, td := range testData {
		tested := getHandlePath(td.test)
		if tested != td.exp {
			t.Errorf(
				"index: %d, handle path mismatch, exp: '%s', got: '%s', for: '%s'",
				i, td.exp, tested, td.test,
			)
		}
	}
	myLogger = newTestLogger(t, util.Ptr(logger.SILENT))
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
						"foo": &routeParameter{
							source: sourceURL,
						},
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
						"foo": &routeParameter{
							source: sourceURL,
						},
						"bar": &routeParameter{
							source: sourceURL,
						},
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
						"foo": &routeParameter{
							source: sourceURL,
						},
						"bar": &routeParameter{
							source: sourceURL,
						},
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
						"foo": &routeParameter{
							source: sourceURL,
						},
						"bar": &routeParameter{
							source: sourceURL,
						},
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
						"foo": &routeParameter{
							source: sourceURL,
						},
						"baz": &routeParameter{
							source: sourceURL,
						},
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
	res  bool
	err  error
	code *int
}

var callbackDataMap map[string]callbackData = map[string]callbackData{
	"cb1": callbackData{
		true,
		nil,
		nil,
	},
	"cb2": callbackData{
		false,
		fmt.Errorf("cb2 error"),
		util.Ptr(http.StatusBadRequest),
	},
	"cb3": callbackData{
		true,
		fmt.Errorf("cb3 error"),
		util.Ptr(http.StatusBadRequest),
	},
	"cb4": callbackData{
		false,
		nil,
		nil,
	},
	"cb5": callbackData{
		true,
		nil,
		nil,
	},
}

type testLogger struct {
	t *testing.T
}

var logLevel logger.LogLevel = logger.DEBUG

func newTestLogger(t *testing.T, pLevel *logger.LogLevel) logger.Logger {
	level := logLevel
	if pLevel != nil {
		level = *pLevel
	}
	return logger.NewLogger(level, "test logger", &testLogger{t})
}

func (t *testLogger) Write(data []byte) (int, error) {
	t.t.Logf(string(data))
	return len(data), nil
}

func makeCallback(
	tLogger logger.Logger, name string, res bool, err error, code *int,
) Callback {
	//use the headers to verify params are correct
	return func(params map[string]string, w http.ResponseWriter, r *http.Request) (bool, error) {
		w.Header().Add("Call_list", name)
		for k, v := range params {
			w.Header().Add(k, v)
		}
		if code != nil && *code != http.StatusOK {
			http.Error(w, err.Error(), *code)
		}
		return res, err
	}
}

type testPath struct {
	serverPath   string
	target       string
	method       string
	body         *string
	expParams    map[string]string
	expCallbacks []string
	expCode      int
}

const StatusGoodRequest int = http.StatusOK

func TestServer(t *testing.T) {
	callbacks := make(map[string]Callback)
	myLogger = newTestLogger(t, nil)
	for name, cb := range callbackDataMap {
		callbacks[name] = makeCallback(myLogger, name, cb.res, cb.err, cb.code)
	}
	testData := []struct {
		serverYaml    string
		msg           string
		serverCreated bool
		testPaths     []testPath
	}{
		{
			//0
			"testdata/basic.yaml",
			"simplest test",
			true,
			[]testPath{
				testPath{
					"/",
					"http://example.com?foo=blarg",
					"GET",
					nil,
					map[string]string{
						"Foo": "blarg",
					},
					[]string{"cb1", "cb5"},
					http.StatusOK,
				},
			},
		},
		{
			//1
			"testdata/basic.yaml",
			"single dynamic parameter",
			true,
			[]testPath{
				testPath{
					"/foo/",
					"http://example.com/foo/123",
					"GET",
					nil,
					map[string]string{
						"Blarg": "123",
					},
					[]string{"cb1", "cb5"},
					http.StatusOK,
				},
			},
		},
		{
			//2
			"testdata/basic.yaml",
			"incorrect parameter type",
			true,
			[]testPath{
				testPath{
					"/bar/",
					"http://example.com/bar?blarg=123",
					"GET",
					nil,
					nil,
					nil,
					http.StatusBadRequest,
				},
			},
		},
		{
			//3
			"testdata/basic.yaml",
			"incorrect parameter type",
			true,
			[]testPath{
				//0
				testPath{
					"/bar/",
					"http://example.com/bar/yolo",
					"GET",
					nil,
					nil,
					nil,
					http.StatusBadRequest,
				},
				//1
				testPath{
					"/bar/",
					"http://example.com/bar/true",
					"GET",
					nil,
					map[string]string{
						"Blarg": "true",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//2
				testPath{
					"/foo/",
					"http://example.com/foo/123.43",
					"GET",
					nil,
					map[string]string{
						"Blarg": "123.43",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//3
				testPath{
					"/baz",
					"http://example.com/baz",
					"GET",
					nil,
					nil,
					[]string{"cb2"},
					http.StatusBadRequest,
				},
			},
		},
		{
			//4
			"testdata/advanced_get.yaml",
			"advanced get",
			true,
			[]testPath{
				//0
				testPath{
					"/bar/",
					"http://example.com/bar/true",
					"GET",
					nil,
					map[string]string{
						"Blarg": "true",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//1
				testPath{
					"/bar/",
					"http://example.com/bar/FalSe",
					"GET",
					nil,
					map[string]string{
						"Blarg": "FalSe",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//2
				testPath{
					"/bar/",
					"http://example.com/bar?blarg=false",
					"GET",
					nil,
					map[string]string{
						"Blarg": "false",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//3
				testPath{
					"/bar/",
					"http://example.com/bar/true?blarg=false",
					"GET",
					nil,
					map[string]string{
						"Blarg": "false",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//4
				testPath{
					"/bar/",
					"http://example.com/bar/true?blarg=yolo",
					"GET",
					nil,
					nil,
					nil,
					http.StatusBadRequest,
				},
				//5
				testPath{
					"/foo/",
					"http://example.com/foo?biff=yolo",
					"GET",
					nil,
					map[string]string{
						"Biff": "yolo",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//6
				testPath{
					"/foo/",
					"http://example.com/foo/123.43?biff=yolo",
					"GET",
					nil,
					map[string]string{
						"Biff":  "yolo",
						"Blarg": "123.43",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//7
				testPath{
					"/foo/",
					"http://example.com/foo/123.43",
					"GET",
					nil,
					nil,
					nil,
					http.StatusBadRequest,
				},
				//8
				testPath{
					"/foo/",
					"http://example.com/foo/123.43?biff=",
					"GET",
					nil,
					nil,
					nil,
					http.StatusBadRequest,
				},
				//9
				testPath{
					"/baz",
					"http://example.com/baz?foo=123-12-1234",
					"GET",
					nil,
					map[string]string{
						"Foo": "123-12-1234",
					},
					[]string{"cb2"},
					http.StatusBadRequest,
				},
				//10
				testPath{
					"/biff",
					"http://example.com/biff?foo=123-12-1234",
					"GET",
					nil,
					map[string]string{
						"Foo": "123-12-1234",
					},
					[]string{"cb5", "cb1"},
					StatusGoodRequest,
				},
				//11
				testPath{
					"/biff",
					"http://example.com/biff?foo=blarg",
					"GET",
					nil,
					nil,
					nil,
					http.StatusBadRequest,
				},
				//12
				testPath{
					"/",
					"http://example.com/?bar=blarg",
					"GET",
					nil,
					nil,
					nil,
					http.StatusBadRequest,
				},
				//13
				testPath{
					"/",
					"http://example.com/?foo=blarg",
					"GET",
					nil,
					map[string]string{
						"Foo": "blarg",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
			},
		},
		{
			//5
			"testdata/post.yaml",
			"post forms",
			true,
			[]testPath{
				//0
				testPath{
					"/",
					"http://example.com",
					"GET",
					nil,
					nil,
					nil,
					http.StatusBadRequest,
				},
				//1
				testPath{
					"/",
					"http://example.com?foo=blarg",
					"GET",
					nil,
					map[string]string{
						"Foo": "blarg",
					},
					[]string{"cb1"},
					StatusGoodRequest,
				},
				//2
				testPath{
					"/",
					"http://example.com?foo=blarg&bar=baz",
					"GET",
					nil,
					nil,
					nil,
					http.StatusBadRequest,
				},
				//3
				testPath{
					"/foo",
					"http://example.com?foo=blarg",
					"POST",
					nil,
					map[string]string{
						"Foo": "blarg",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//4
				testPath{
					"/baz/",
					"http://example.com/baz/123.3",
					"POST",
					nil,
					map[string]string{
						"Foo": "123.3",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				//5
				testPath{
					"/baz/",
					"http://example.com/baz/123.3/5?bar=6",
					"POST",
					nil,
					map[string]string{
						"Foo": "123.3",
						"Bar": "6",
					},
					[]string{"cb1", "cb5"},
					StatusGoodRequest,
				},
				/*
					Don't really know how to manually build the body of
					a post request, don't really care, im sure it'll come back up
					if it don't work
					//6
					testPath{
						"/baz/",
						"http://example.com/baz/123.3",
						"POST",
						util.Ptr("bar=0"),
						map[string]string{
							"Foo": "123.3",
							"Bar": "0",
						},
						[]string{"cb1", "cb5"},
						StatusGoodRequest,
					},
					//7
					testPath{
						"/baz/",
						"http://example.com/baz/123.3/1",
						"POST",
						util.Ptr("bar=0"),
						map[string]string{
							"Foo": "123.3",
							"Bar": "1",
						},
						[]string{"cb1", "cb5"},
						StatusGoodRequest,
					},
					//7
					testPath{
						"/baz/",
						"http://example.com/baz/123.3/1?bar=2",
						"POST",
						util.Ptr("bar=0"),
						map[string]string{
							"Foo": "123.3",
							"Bar": "2",
						},
						[]string{"cb1", "cb5"},
						StatusGoodRequest,
					},
				*/
			},
		},
	}

	tMetaData, err := util.GetTestMetaData()
	if err != nil {
		t.Fatalf("failed retreiving test meta data with error: '%s'", err)
		return
	}
	toTest := testData[:]
	if tMetaData.Start < 0 {
		if tMetaData.Stop > 0 {
			toTest = testData[:tMetaData.Stop]
		}
	} else {
		if tMetaData.Stop < 0 {
			toTest = testData[tMetaData.Start:]
		} else {
			toTest = testData[tMetaData.Start:tMetaData.Stop]
		}
	}
	for i, td := range toTest {
		serverYaml, err := os.Open(td.serverYaml)
		if err != nil {
			t.Fatalf(
				getTestMessage(
					i, td.msg, "failed opening test yaml for: '%s' with error: '%s'",
					td.serverYaml, err,
				),
			)
		}
		tmpServer, err := NewServer(serverYaml, callbacks)
		serverYaml.Close()
		if !td.serverCreated && tmpServer != nil {
			t.Errorf(
				getTestMessage(i, td.msg,
					"server should not have been created and thrown error",
				),
			)
			continue
		}
		if td.serverCreated && tmpServer == nil {
			t.Errorf(
				getTestMessage(
					i, td.msg, "server was not created with error: '%s'", err,
				),
			)
			continue
		}
		testServer := tmpServer.(*server)
		testPaths := td.testPaths[:]
		if len(tMetaData.Rest) > 0 {
			if len(tMetaData.Rest) == 1 {
				testPaths = td.testPaths[tMetaData.Rest[0]:]
			} else {
				testPaths = td.testPaths[tMetaData.Rest[0]:tMetaData.Rest[1]]
			}
		}
		//t.Logf("tMetaData: %v", tMetaData)
		//t.Logf("testPaths: %v", testPaths)
		//return
		for _, testPath := range testPaths {
			w := httptest.NewRecorder()
			var body io.Reader = nil
			if testPath.body != nil {
				body = strings.NewReader(*testPath.body)
			}
			r := httptest.NewRequest(testPath.method, testPath.target, body)
			testServer.pathHandlers[testPath.serverPath].ServeHTTP(w, r)
			if len(testPath.expCallbacks) != len(w.HeaderMap["Call_list"]) {
				t.Errorf(
					getTestMessage(
						i, td.msg,
						"incorrect callback count, exp: %d, got: %d",
						len(testPath.expCallbacks), len(w.HeaderMap["Call_list"]),
					),
				)
			} else {
				for cbIndex, cb := range w.HeaderMap["Call_list"] {
					if testPath.expCallbacks[cbIndex] != cb {
						t.Errorf(
							getTestMessage(
								i, td.msg,
								"incorrect callback sequence at index %d, exp: %s, got: %s",
								cbIndex, testPath.expCallbacks[cbIndex], cb,
							),
						)
					}
				}
			}
			if testPath.expCode != w.Code {
				t.Errorf(
					getTestMessage(
						i, td.msg, "unexpected response code, exp: %d, res: %d",
						testPath.expCode, w.Code,
					),
				)
			}
			if testPath.expCode != StatusGoodRequest {
				continue
			}
			if len(testPath.expParams)+1 != len(w.HeaderMap) {
				t.Errorf(
					getTestMessage(
						i, td.msg,
						"parameter map incorrect length, exp: %d, got: %d",
						len(testPath.expParams), len(w.HeaderMap)-1,
					),
				)
			}
			for hName, hValues := range w.HeaderMap {
				if hName == "Call_list" {
					continue
				}
				if _, ok := testPath.expParams[hName]; !ok {
					t.Errorf(
						getTestMessage(
							i, td.msg, "parameter map missing value(s) for key: '%s'",
							hName,
						),
					)
					continue
				}
				if testPath.expParams[hName] != hValues[0] {
					t.Errorf(
						getTestMessage(
							i, td.msg,
							"parameter value incorrect, exp: %s, got: %s",
							testPath.expParams[hName], hValues[0],
						),
					)
				}
			}
		}
	}
}
