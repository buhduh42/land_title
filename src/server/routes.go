package server

import (
	"encoding/json"
	"fmt"
	"io"
	"landtitle/util"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Route interface{}

const defRequiredParameter bool = true

type sourceType byte

/*
sourceURL would be /foo/{bar}
sourceForm is embedded in the POST form
sourceQuery is the GET parameter
precedence in increasing order, url -> form  -> query
*/
const (
	sourceURL sourceType = 1 << iota
	sourceForm
	sourceQuery
	defSource = sourceQuery
)

type sourceTypeName string

const (
	sourceURLName   sourceTypeName = "url"
	sourceFormName                 = "form"
	sourceQueryName                = "query"
	defSourceName                  = sourceQueryName
)

type ParamYaml struct {
	Type       string `yaml:"type,omitempty"`
	Regex      string `yaml:"regex,omitempty"`
	Required   *bool  `yaml:"required,omitempty"`
	SourceType string `yaml:"source,omitempty"`
}

func (p *ParamYaml) String() string {
	toPrint, _ := json.MarshalIndent(p, "", "  ")
	return string(toPrint)
}

const dynamicPathPattern string = `^\{[a-z]\w*\}$`

type httpParameterType string

const (
	numberParameterType  httpParameterType = "number"
	stringParameterType                    = "string"
	booleanParameterType                   = "boolean"
	defParameterType                       = stringParameterType
)

func newHttpParameterType(p string) (*httpParameterType, error) {
	toRet := httpParameterType(p)
	switch toRet {
	case numberParameterType:
		fallthrough
	case stringParameterType:
		fallthrough
	case booleanParameterType:
		return util.Ptr(toRet), nil
	}
	return nil, fmt.Errorf("unrecognized httpParameterType: '%s'", p)
}

var (
	dynamicPathRegex         *regexp.Regexp = regexp.MustCompile(dynamicPathPattern)
	defStringParameterRegex                 = regexp.MustCompile(`\w+`)
	defNumberParameterRegex                 = regexp.MustCompile(`[+-]?([0-9]*[.])?[0-9]+`)
	defBooleanParameterRegex                = regexp.MustCompile(`(?i)((true)|(false))`)
)

var parameterRegexpMap map[httpParameterType]*regexp.Regexp = map[httpParameterType]*regexp.Regexp{
	numberParameterType:  defNumberParameterRegex,
	stringParameterType:  defStringParameterRegex,
	booleanParameterType: defBooleanParameterRegex,
}

type MethodsYaml []string

type CallbacksYaml []string

type RouteYaml struct {
	Methods   []string              `yaml:"methods,omitempty,flow"`
	Params    map[string]*ParamYaml `yaml:"params,omitempty,flow"`
	Callbacks []string              `yaml:"callbacks,flow"`
}

func (r *RouteYaml) String() string {
	toPrint, _ := json.MarshalIndent(r, "", "  ")
	return string(toPrint)
}

type route struct {
	methods   []httpMethod
	callbacks []string
	params    routeParameterMap
}

func (r *route) String() string {
	toPrint, _ := json.MarshalIndent(r, "", "  ")
	return string(toPrint)
}

type routeParameter struct {
	pType    httpParameterType `json:"type"`
	regex    *regexp.Regexp
	required bool
	source   sourceType `json:"source"`
}

func (r *routeParameter) isValid(check string) bool {
	if !r.required {
		if check == "" {
			return true
		}
	}
	if !parameterRegexpMap[r.pType].MatchString(check) {
		return false
	}
	if r.regex != nil {
		if !r.regex.MatchString(check) {
			return false
		}
	}
	return true
}

type routeParameterMap map[string]*routeParameter

func (params routeParameterMap) validate(values map[string]string) error {
	var value string
	var ok bool
	for pName, param := range params {
		if value, ok = values[pName]; param.required && !ok {
			return fmt.Errorf("required parameter '%s' missing", pName)
		}
		if !param.isValid(value) {
			return fmt.Errorf(
				"parameter '%s' is not valid, value: '%s'",
				pName, value,
			)
		}
	}
	return nil
}

func newMethods(ms []string) ([]httpMethod, error) {
	//get is the default method when not defined
	if len(ms) == 0 {
		return []httpMethod{getMethod}, nil
	}
	toRet := make([]httpMethod, len(ms))
	for i, m := range ms {
		tmp, err := newHttpMethod(m)
		if err != nil {
			return nil, err
		}
		toRet[i] = *tmp
	}
	return toRet, nil
}

func newParam(p *ParamYaml) (*routeParameter, error) {
	if p == nil {
		p = &ParamYaml{
			Type: stringParameterType,
		}
	}
	if p.Required == nil {
		p.Required = util.Ptr(true)
	}
	pType, err := newHttpParameterType(p.Type)
	if err != nil {
		return nil, err
	}
	if pType == nil {
		return nil, fmt.Errorf("parameter type was set to nil for '%s'", p.Type)
	}
	var regex *regexp.Regexp = nil
	if p.Regex != "" {
		if regex, err = regexp.Compile(p.Regex); err != nil {
			return nil, err
		}
	}
	reqSourceType, err := getSourceMask(p.SourceType)
	if err != nil {
		return nil, err
	}
	return &routeParameter{
		pType:    *pType,
		regex:    regex,
		required: *p.Required,
		source:   reqSourceType,
	}, nil
}

func getSourceMask(yamlSourceName string) (sourceType, error) {
	if yamlSourceName == "" {
		yamlSourceName = string(defSourceName)
	}
	sources := strings.Split(yamlSourceName, "|")
	toRet := sourceType(0)
	for _, s := range sources {
		myLogger.Tracef("processing source: '%s'", s)
		switch s {
		case "form":
			toRet |= sourceForm
		case "query":
			toRet |= sourceQuery
		case "url":
			toRet |= sourceURL
		default:
			return 0, fmt.Errorf("unrecognized parameter source: '%s'", s)
		}
	}
	return toRet, nil
}

func newRoute(r *RouteYaml) (*route, error) {
	if len(r.Callbacks) == 0 {
		return nil, fmt.Errorf("at least one callback is required")
	}
	methods, err := newMethods(r.Methods)
	if err != nil {
		return nil, err
	}
	params := make(map[string]*routeParameter)
	for pKey, param := range r.Params {
		tmp, err := newParam(param)
		if err != nil {
			return nil, err
		}
		params[pKey] = tmp
	}
	return &route{
		methods:   methods,
		callbacks: r.Callbacks,
		params:    params,
	}, nil
}

func loadRouteYaml(
	r io.Reader,
) (map[string]*RouteYaml, error) {
	rawBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	yamlData := make(map[string]*RouteYaml)
	if err = yaml.Unmarshal(rawBytes, &yamlData); err != nil {
		myLogger.Errorf("failed unmarshaling yaml with error: '%s'", err)
		return nil, err
	}
	for p, rte := range yamlData {
		if err = verifyPath(p); err != nil {
			myLogger.Errorf(
				"could not verify path: '%s' for route yaml with error: '%s'",
				err,
			)
			return nil, err
		}
		//can't loop through map values, as they may be nil, the
		//Required check will blow it up
		for k, _ := range rte.Params {
			if rte.Params[k] == nil {
				rte.Params[k] = &ParamYaml{
					Type: defParameterType,
				}
			}
			if rte.Params[k].Type == "" {
				rte.Params[k].Type = defParameterType
			}
			if rte.Params[k].Required == nil {
				rte.Params[k].Required = util.Ptr(defRequiredParameter)
			}
			if rte.Params[k].SourceType == "" {
				rte.Params[k].SourceType = string(defSourceName)
			}
			myLogger.Tracef(
				"loading route params for param name: '%s' and param:\n%s",
				k, rte.Params[k],
			)
		}
	}
	myLogger.Tracef("loaded yaml data:\n%s", yamlData)
	return yamlData, nil
}

func verifyPath(path string) error {
	dynamic := false
	for _, p := range strings.Split(path, "/") {
		if len(p) == 0 {
			continue
		}
		if string(p[0]) == "{" && string(p[len(p)-1]) == "}" {
			if !dynamicPathRegex.MatchString(p) {
				return fmt.Errorf(
					"dynamic paths must match pattern: '%s'",
					dynamicPathPattern,
				)
			}
			dynamic = true
		} else if dynamic {
			return fmt.Errorf(
				"no non dynamic path patterns can be after dynamic patterns, eg /{foo}/foo",
			)
		}
	}
	return nil
}

func loadRoutes(r io.Reader) (map[string]*route, error) {
	routeYaml, err := loadRouteYaml(r)
	if err != nil {
		return nil, err
	}
	toRet := make(map[string]*route)
	for k, v := range routeYaml {
		rte, err := newRoute(v)
		if err != nil {
			return nil, err
		}
		toRet[k] = rte
	}
	return toRet, nil
}
