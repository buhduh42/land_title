package server

import (
	"fmt"
	"strings"
	"testing"
)

type routeTestData struct {
	yamlString   string
	expPath      string
	expError     bool
	expParams    paramsTestData
	expCallbacks []string
	expMethods   []string
	msg          string
}

type paramsTestData map[string]*paramTestData

type paramTestData struct {
	expType     string
	expRequired bool
	expRegex    string
}

func renderMsg(index int, run string, msg string, args ...interface{}) string {
	return fmt.Sprintf(
		"index: %d, run_id: '%s', %s", index, run, fmt.Sprintf(msg, args...),
	)
}

func (r *routeTestData) compare(t *testing.T, index int, runString string) {
	routes, err := loadRouteYaml(strings.NewReader(r.yamlString))
	if r.expError {
		if err == nil {
			t.Errorf(renderMsg(index, runString, "expected error, msg: '%s'", r.msg))
		}
		return
	}
	if err != nil {
		t.Errorf(
			renderMsg(
				index, runString, "unexpected error: '%s', msg: '%s'",
				err, r.msg,
			),
		)
	}
	if len(routes) != 1 {
		t.Logf("%+v\n", routes)
		t.Fatalf(
			renderMsg(
				index, runString, "only testing single route maps, '%s'", r.msg,
			),
		)
		return
	}
	routeData, ok := routes[r.expPath]
	if !ok {
		t.Errorf(
			renderMsg(
				index, runString, "expected path: '%s', route(s): '%v', msg: '%s'",
				r.expPath, routeData, r.msg,
			),
		)
		return
	}
	//NOTE: forcing order, too lazy to check order independence
	if len(routeData.Callbacks) != len(r.expCallbacks) {
		t.Errorf(
			renderMsg(
				index, runString,
				"callback length mismatch, exp: %d,  got: %d, msg: '%s'",
				len(r.expCallbacks), len(routeData.Callbacks), r.msg,
			),
		)
		return
	}
	for j, cb := range routeData.Callbacks {
		if r.expCallbacks[j] != cb {
			t.Errorf(
				renderMsg(
					index, runString,
					"callback mismatch, exp: '%s', got: '%s', msg: '%s'",
					r.expCallbacks[j], cb, r.msg,
				),
			)
			//might need to break if forces variable instantiation
			return
		}
	}
	if len(routeData.Methods) != len(r.expMethods) {
		t.Errorf(
			renderMsg(
				index, runString, "method length mismatch, exp: %d,  got: %d, msg: '%s'",
				len(r.expMethods), len(routeData.Methods), r.msg,
			),
		)
		return
	}
	for j, m := range routeData.Methods {
		if r.expMethods[j] != m {
			t.Errorf(
				renderMsg(
					index, runString, "method mismatch, exp: '%s', got: '%s', msg: '%s'",
					r.expMethods[j], m, r.msg,
				),
			)
			//might need to break if forces variable instantiation
			return
		}
	}
	if len(r.expParams) != len(routeData.Params) {
		t.Errorf(
			renderMsg(
				index, runString, "params length mismatch, exp: %d, got: %d, msg: '%s'",
				len(r.expParams), len(routeData.Params), r.msg,
			),
		)
		return
	}
	for expPKey, expParam := range r.expParams {
		param, ok := routeData.Params[expPKey]
		if !ok {
			t.Errorf(
				renderMsg(
					index, runString, "expected param key: '%s', msg: '%s'",
					expPKey, r.msg,
				),
			)
			return
		}
		//it's found
		if param == nil && ok {
			t.Logf(
				renderMsg(
					index, runString,
					"param: '%s', warning, yaml set a param to nil, may be OK",
					expPKey,
				),
			)
			continue
		}
		if param.Type != expParam.expType {
			t.Errorf(
				renderMsg(
					index, runString,
					"parameter type for '%s' mismatch, exp: '%s', got: '%s', msg: '%s'",
					expPKey, expParam.expType, param.Type, r.msg,
				),
			)
		}
		if param.Regex != expParam.expRegex {
			t.Errorf(
				renderMsg(
					index, runString,
					"parameter regex for '%s' mismatch, exp: '%s', got: '%s', msg: '%s'",
					expPKey, expParam.expRegex, param.Regex, r.msg,
				),
			)
		}
		if param.Required != expParam.expRequired {
			t.Errorf(
				renderMsg(
					index, runString,
					"parameter required for '%s' mismatch, exp: %t, got: %t, msg: '%s'",
					expPKey, expParam.expRequired, param.Required, r.msg,
				),
			)
		}
	}
}

//NOTE, too lazy to test for unordered callbacks
//TODO, here? need to lookup methods
func (r *RouteYaml) compare(toCmp *route) bool {
	if len(r.Callbacks) != len(toCmp.callbacks) {
		return false	
	}
	for i, m := range r.Methods {
		if m != toCmp[i] {
			return false
		}
	}

}

func TestRouteYamlToRoute(t *testing.T) {
	testData := []*struct{
		routeYaml *RouteYaml
		expRoute *route
		expError bool
		msg string
	}{
		&struct{
			&RouteYaml{
				Callbacks: []string{"foo"},
			},
			&route{
				callbacks: []string{"foo"},
				methods: []httpMethod{getMethod},
			},
			false,
			"simplest test case",
		},
	}
	for i, td := range testData {

	}
}

func TestRouteYaml(t *testing.T) {
	testData := []*routeTestData{
		//0
		&routeTestData{
			expPath: "/",
			yamlString: `
/:
  callbacks:
    - single
`,
			expError:  false,
			expParams: paramsTestData{},
			expCallbacks: []string{
				"single",
			},
			expMethods: []string{},
			msg:        "root path, single callback",
		},
		//1
		&routeTestData{
			expPath: "/",
			yamlString: `
/:
  params:
    foo:
      type: string
      required: true
    bar:
      type: number
      regex: \d{1,5}
      required: false
  callbacks:
    - first
    - second
`,
			expError: false,
			expParams: paramsTestData{
				"foo": &paramTestData{
					expType:     "string",
					expRequired: true,
					expRegex:    "",
				},
				"bar": &paramTestData{
					expType:     "number",
					expRequired: false,
					expRegex:    `\d{1,5}`,
				},
			},
			expCallbacks: []string{
				"first",
				"second",
			},
			expMethods: []string{},
			msg:        "root path, multiple params",
		},
		//2
		&routeTestData{
			expPath: "/foo/bar/{yolo}/biff",
			yamlString: `
/foo/bar/{yolo}/biff:
  methods:
    - methods_arent_validated_here
  params:
    bar:
      type: number
      regex: '\d{1,9}-\w?-\\'
      required: false
  callbacks:
    - blarg
`,
			expError: false,
			expParams: paramsTestData{
				"bar": &paramTestData{
					expType:     "number",
					expRequired: false,
					expRegex:    `\d{1,9}-\w?-\\`,
				},
			},
			expCallbacks: []string{
				"blarg",
			},
			expMethods: []string{
				"methods_arent_validated_here",
			},
			msg: "regex with backslashes",
		},
		//3
		&routeTestData{
			expPath: "/foo/bar/{yolo}/biff",
			yamlString: `
/foo/bar/{yolo}/biff:
  methods:
    - post 
    - get
    - methods_arent_validated_here
  params:
    bar:
      type: number
      regex: '[4-9]{4,9}'
      required: false
  callbacks:
    - blarg
`,
			expError: false,
			expParams: paramsTestData{
				"bar": &paramTestData{
					expType:     "number",
					expRequired: false,
					expRegex:    "[4-9]{4,9}",
				},
			},
			expCallbacks: []string{
				"blarg",
			},
			expMethods: []string{
				"post",
				"get",
				"methods_arent_validated_here",
			},
			msg: "multiple methods",
		},
		//4
		&routeTestData{
			expPath: "/foo/bar/{yolo}/biff",
			yamlString: `
/foo/bar/{yolo}/biff:
  methods:
    - post 
  params:
    bar:
      type: number
      regex: '[4-9]{4,9}'
      required: false
  callbacks:
    - blarg
`,
			expError: false,
			expParams: paramsTestData{
				"bar": &paramTestData{
					expType:     "number",
					expRequired: false,
					expRegex:    "[4-9]{4,9}",
				},
			},
			expCallbacks: []string{
				"blarg",
			},
			expMethods: []string{
				"post",
			},
			msg: "complicated path",
		},
		//5
		&routeTestData{
			expPath: "/bazz/biff",
			yamlString: `
/bazz/biff:
  params:
    bar:
  callbacks:
    - yolo
`,
			expError: false,
			expParams: paramsTestData{
				"bar": &paramTestData{},
			},
			expCallbacks: []string{
				"yolo",
			},
			expMethods: []string{},
			msg:        "empty parameters are acceptable",
		},
	}
	for i, td := range testData {
		td.compare(t, i, "route yaml")
	}
}
