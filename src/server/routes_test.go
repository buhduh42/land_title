package server

import (
	"fmt"
	"landtitle/util"
	"os"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
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
			/*
				t.Logf(
					renderMsg(
						index, runString,
						"param: '%s', warning, yaml set a param to nil, may be OK",
						expPKey,
					),
				)
			*/
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
		if *param.Required != expParam.expRequired {
			t.Errorf(
				renderMsg(
					index, runString,
					"parameter required for '%s' mismatch, exp: %t, got: %t, msg: '%s'",
					expPKey, expParam.expRequired, *param.Required, r.msg,
				),
			)
		}
	}
}

func (p *routeParameter) compare(toCmp *routeParameter) bool {
	if p.pType != toCmp.pType {
		return false
	}
	/*
		//will need to test the regex seperately
		if p.regex != toCmp.regex {
			return false
		}
	*/
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
		//0
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
		//1
		{
			&RouteYaml{
				Callbacks: []string{},
			},
			nil,
			true,
			"callbacks can't be empty",
		},
		//2
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
		//3
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
			false,
			"list of accepted methods",
		},
		//4
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
						regex:    nil,
						required: true,
					},
				},
			},
			false,
			"empty parameter, verify defaults",
		},
		//5
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
						regex:    regexp.MustCompile(`\d{1,3}`),
						required: true,
					},
				},
			},
			false,
			"verify various parameter values",
		},
		//6
		{
			&RouteYaml{
				Callbacks: []string{
					"foo",
				},
				Params: map[string]*ParamYaml{
					"foo": &ParamYaml{
						Type:     "blarg",
						Required: util.Ptr(false),
					},
				},
			},
			nil,
			true,
			"unrecognized parameter type",
		},
		//7
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
						Required: util.Ptr(false),
					},
					"bar": &ParamYaml{
						Type:     "string",
						Regex:    `this/is/a/regex`,
						Required: util.Ptr(false),
					},
					"baz": &ParamYaml{
						Type: "boolean",
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
						regex:    regexp.MustCompile(`\d{1,3}`),
						required: false,
					},
					"bar": &routeParameter{
						pType:    stringParameterType,
						regex:    regexp.MustCompile(`this is a regex`),
						required: false,
					},
					"baz": &routeParameter{
						pType:    booleanParameterType,
						required: true,
						regex:    nil,
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
			expPath: "/foo/bar/biff/{yolo}",
			yamlString: `
/foo/bar/biff/{yolo}:
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
			expPath: "/foo/bar/biff/{yolo}",
			yamlString: `
/foo/bar/biff/{yolo}:
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
			expPath: "/foo/bar/biff/{yolo}",
			yamlString: `
/foo/bar/biff/{yolo}:
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
				"bar": &paramTestData{
					expType:     defParameterType,
					expRequired: defRequiredParameter,
				},
			},
			expCallbacks: []string{
				"yolo",
			},
			expMethods: []string{},
			msg:        "empty parameters are acceptable",
		},
		//6
		&routeTestData{
			expPath: "",
			yamlString: `
/bazz/{yolo}/biff:
  params:
    bar:
  callbacks:
    - yolo
`,
			expError:     true,
			expParams:    nil,
			expCallbacks: nil,
			expMethods:   []string{},
			msg:          "dynamic pattern before non-dynamic patterns is an error",
		},
		//7
		&routeTestData{
			expPath: "",
			yamlString: `
/bazz/{9adf}:
  params:
    bar:
  callbacks:
    - yolo
`,
			expError:     true,
			expParams:    nil,
			expCallbacks: nil,
			expMethods:   []string{},
			msg: fmt.Sprintf(
				"dynamic pattern doesn't match regex: '%s'",
				dynamicPathPattern,
			),
		},
		//7
		&routeTestData{
			expPath: "",
			yamlString: `
/bazz/{*&4adf}:
  params:
    bar:
  callbacks:
    - yolo
`,
			expError:     true,
			expParams:    nil,
			expCallbacks: nil,
			expMethods:   []string{},
			msg: fmt.Sprintf(
				"dynamic pattern doesn't match regex: '%s'",
				dynamicPathPattern,
			),
		},
	}
	for i, td := range testData {
		td.compare(t, i, "route yaml")
	}
}

func TestRouteParameterIsValid(t *testing.T) {
	testdata := []struct {
		value     string
		parameter *routeParameter
		exp       bool
		msg       string
	}{
		{
			"foo",
			&routeParameter{
				pType:    stringParameterType,
				regex:    nil,
				required: true,
			},
			true,
			"simple string parameter",
		},
		{
			"123",
			&routeParameter{
				pType:    numberParameterType,
				regex:    nil,
				required: true,
			},
			true,
			"simple number parameter",
		},
		{
			"-123",
			&routeParameter{
				pType:    numberParameterType,
				regex:    nil,
				required: true,
			},
			true,
			"negative integer",
		},
		{
			"-123.42",
			&routeParameter{
				pType:    numberParameterType,
				regex:    nil,
				required: true,
			},
			true,
			"negative float",
		},
		{
			"123.42",
			&routeParameter{
				pType:    numberParameterType,
				regex:    nil,
				required: true,
			},
			true,
			"floating point",
		},
		{
			"false",
			&routeParameter{
				pType:    booleanParameterType,
				regex:    nil,
				required: true,
			},
			true,
			"boolean",
		},
		{
			"true",
			&routeParameter{
				pType:    booleanParameterType,
				regex:    nil,
				required: true,
			},
			true,
			"boolean",
		},
		{
			"tRUe",
			&routeParameter{
				pType:    booleanParameterType,
				regex:    nil,
				required: true,
			},
			true,
			"case insensitive bool",
		},
		{
			"8769",
			&routeParameter{
				pType:    stringParameterType,
				regex:    regexp.MustCompile(`[\D]+`),
				required: true,
			},
			false,
			"string can't have numbers",
		},
		{
			"8769",
			&routeParameter{
				pType:    booleanParameterType,
				regex:    nil,
				required: true,
			},
			false,
			"numbers aren't a boolean",
		},
	}
	for i, td := range testdata {
		if td.parameter.isValid(td.value) != td.exp {
			t.Errorf(
				"index: %d, value: '%s', result: %t, exp: %t, msg: '%s'",
				i, td.value, td.parameter.isValid(td.value), td.exp, td.msg,
			)
		}
	}
}

type DummyTestRouteYaml map[string]*ParamYaml

// was simple string comparison before, will need to actually validate if
// regex is  properly set here
func TestRouteRegex(t *testing.T) {
	testdata := []struct {
		yamlString string
		value      string
		pattern    *regexp.Regexp
		exp        bool
		msg        string
	}{
		{
			`
foo:
  type: string
  required: true
`,
			"a string",
			nil,
			true,
			"simple string test",
		},
		{
			`
foo:
  type: number
  required: true
`,
			"124",
			nil,
			true,
			"simple number test",
		},
		{
			`
foo:
  type: boolean
  required: true
`,
			"true",
			nil,
			true,
			"simple bboolean test",
		},
		{
			`
foo:
  type: string
  required: true
  regex: '^[\D+]$'
`,
			"456",
			nil,
			false,
			"fail the regexp",
		},
	}
	doError := func(i int, msg, fmtMsg string, vals ...interface{}) {
		t.Errorf(
			"index: %d, msg: '%s', %s",
			i, msg, fmt.Sprintf(fmtMsg, vals...),
		)
	}
	for i, td := range testdata {
		toTest := make(DummyTestRouteYaml)
		err := yaml.Unmarshal([]byte(td.yamlString), &toTest)
		if err != nil {
			t.Errorf("could not load test yaml with error: '%s'", err)
			continue
		}
		for _, pValue := range toTest {
			rteParam, err := newParam(pValue)
			if err != nil {
				t.Errorf("could not instantiate route parameter with error: '%s'", err)
				continue
			}
			res := rteParam.isValid(td.value)
			if res != td.exp {
				doError(i, td.msg, "exp: %t, got: %t", td.value, res)
				continue
			}
		}
	}
}

func TestRealRouteYaml(t *testing.T) {
	testData := map[string]*route{
		"/": &route{
			methods: []httpMethod{
				getMethod,
			},
			callbacks: []string{
				"first",
				"second",
			},
			params: map[string]*routeParameter{
				"foo": &routeParameter{
					pType:    stringParameterType,
					regex:    nil,
					required: true,
				},
				"bar": &routeParameter{
					pType:    numberParameterType,
					regex:    regexp.MustCompile(`\d{1,5}`),
					required: false,
				},
			},
		},
		"/first": &route{
			methods: []httpMethod{
				postMethod,
			},
			callbacks: []string{
				"first",
				"second",
				"only_first",
			},
			params: map[string]*routeParameter{
				"baz": &routeParameter{
					pType:    booleanParameterType,
					regex:    nil,
					required: false,
				},
				"biff": &routeParameter{
					pType:    stringParameterType,
					regex:    regexp.MustCompile(`\d{3}-\d{2}-\d{4}`),
					required: true,
				},
			},
		},
		"/first/{second}": &route{
			methods: []httpMethod{
				getMethod,
				postMethod,
			},
			callbacks: []string{
				"first",
				"second",
				"second_third",
				"only_second",
			},
			params: map[string]*routeParameter{
				"another": &routeParameter{
					pType:    booleanParameterType,
					regex:    nil,
					required: false,
				},
			},
		},
		"/first/{second}/{third}": &route{
			methods: []httpMethod{
				getMethod,
			},
			callbacks: []string{
				"first",
				"second",
				"second_third",
			},
			params: map[string]*routeParameter{},
		},
	}
	r, err := os.Open("./testdata/routes.yaml")
	if err != nil {
		t.Fatalf("failed opening test data with error: '%s'", err)
	}
	routes, err := LoadRoutes(r)
	if err != nil {
		t.Fatalf("failed loading route yaml with error: '%s'", err)
	}
	r.Close()
	if len(testData) != len(routes) {
		t.Errorf(
			"route count mismatch, exp: %d', got: '%d'",
			len(testData), len(routes),
		)
	}
	for path, rte := range testData {
		tmp, ok := routes[path]
		if !ok {
			t.Errorf("did not find route for '%s'", path)
			continue
		}
		toCheck := tmp.(route)
		if !rte.compare(&toCheck) {
			t.Errorf(
				"routes did not match for path: '%s'\nexp:\n%s\ngot:\n%s",
				path, rte, &toCheck,
			)
		}
	}
}
