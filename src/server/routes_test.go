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

func (p *routeParameter) compare(toCmp *routeParameter) bool {
	if p.pType != toCmp.pType {
		return false
	}
	if p.regex != toCmp.regex {
		return false
	}
	if p.required != toCmp.required {
		return false
	}
	return true
}

// callbacks ARE ordered, they must be
func (r *route) compare(toCmp *route) bool {
	if len(r.callbacks) != len(toCmp.callbacks) {
		return false
	}
	for i, c := range r.callbacks {
		if c != toCmp.callbacks[i] {
			return false
		}
	}
	if len(r.methods) != len(toCmp.methods) {
		return false
	}
	for i, m := range r.methods {
		if m != toCmp.methods[i] {
			return false
		}
	}
	//NOTE forcing order
	if len(r.params) != len(toCmp.params) {
		return false
	}
	for i, p := range r.params {
		if !p.compare(toCmp.params[i]) {
			return false
		}
	}
	return true
}

func TestRouteYamlToRoute(t *testing.T) {
	testData := []struct {
		routeYaml *RouteYaml
		expRoute  *route
		expError  bool
		msg       string
	}{
		{
			&RouteYaml{
				Callbacks: []string{"foo"},
			},
			&route{
				callbacks: []string{"foo"},
				methods:   []httpMethod{getMethod},
			},
			false,
			"simplest test case",
		},
		{
			&RouteYaml{
				Callbacks: []string{},
			},
			nil,
			true,
			"callbacks can't be empty",
		},
		{
			&RouteYaml{
				Callbacks: []string{
					"foo",
					"bar",
					"baz",
				},
				Methods: []string{"blarg"},
			},
			nil,
			true,
			"method unrecognized",
		},
		{
			&RouteYaml{
				Callbacks: []string{
					"foo",
					"bar",
					"baz",
				},
				Methods: []string{
					"get",
					"post",
				},
			},
			&route{
				callbacks: []string{
					"foo",
					"bar",
					"baz",
				},
				methods: []httpMethod{
					getMethod,
					postMethod,
				},
			},
			true,
			"list of accepted methods",
		},
		{
			&RouteYaml{
				Callbacks: []string{
					"foo",
				},
				Params: map[string]*ParamYaml{
					"foo": nil,
				},
			},
			&route{
				callbacks: []string{
					"foo",
				},
				methods: []httpMethod{
					getMethod,
				},
				params: map[string]*routeParameter{
					"foo": &routeParameter{
						pType:    stringParameterType,
						regex:    "",
						required: true,
					},
				},
			},
			false,
			"empty parameter, verify defaults",
		},
		{
			&RouteYaml{
				Callbacks: []string{
					"foo",
				},
				Params: map[string]*ParamYaml{
					"foo": &ParamYaml{
						Type:  "number",
						Regex: `\d{1,3}`,
					},
				},
			},
			&route{
				callbacks: []string{
					"foo",
				},
				methods: []httpMethod{
					getMethod,
				},
				params: map[string]*routeParameter{
					"foo": &routeParameter{
						pType:    numberParameterType,
						regex:    `\d{1,3}`,
						required: true,
					},
				},
			},
			false,
			"verify various parameter values",
		},
		{
			&RouteYaml{
				Callbacks: []string{
					"foo",
				},
				Params: map[string]*ParamYaml{
					"foo": &ParamYaml{
						Type:     "blarg",
						Required: false,
					},
				},
			},
			nil,
			true,
			"unrecognized parameter type",
		},
		{
			&RouteYaml{
				Callbacks: []string{
					"foo",
					"bar",
				},
				Methods: []string{
					"get",
					"post",
				},
				Params: map[string]*ParamYaml{
					"foo": &ParamYaml{
						Type:     "number",
						Regex:    `\d{1,3}`,
						Required: false,
					},
					"bar": &ParamYaml{
						Type:     "string",
						Regex:    `this/is/a/regex`,
						Required: true,
					},
					"baz": &ParamYaml{
						Type: "bool",
					},
				},
			},
			&route{
				callbacks: []string{
					"foo",
					"bar",
				},
				methods: []httpMethod{
					getMethod,
					postMethod,
				},
				params: map[string]*routeParameter{
					"foo": &routeParameter{
						pType:    numberParameterType,
						regex:    `\d{1,3}`,
						required: false,
					},
					"bar": &routeParameter{
						pType:    stringParameterType,
						regex:    `this/is/a/regex`,
						required: true,
					},
					"baz": &routeParameter{
						pType:    booleanParameterType,
						required: true,
						regex:    "",
					},
				},
			},
			false,
			"more complex good case",
		},
	}
	runString := "test route yaml to route"
	for i, td := range testData {
		toCheck, err := newRoute(td.routeYaml)
		if td.expError {
			if err == nil {
				t.Errorf(
					renderMsg(i, runString, "expected error"),
				)
			}
			continue
		}
		if err != nil {
			t.Errorf(
				renderMsg(
					i, runString, "did not expect error, got '%s'", err,
				),
			)
			continue
		}
		if !td.expRoute.compare(toCheck) {
			t.Errorf(
				renderMsg(
					i, runString,
					"routes did not match, expected:\n %s\n got:\n %s",
					td.expRoute, toCheck,
				),
			)
			continue
		}
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
