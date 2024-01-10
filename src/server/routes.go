package server

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

type Route interface{}

type ParamYaml struct {
	Type     string `yaml:"type,omitempty"`
	Regex    string `yaml:"regex,omitempty"`
	Required bool   `yaml:"required,omitempty"`
}

type MethodsYaml []string

type CallbacksYaml []string

type RouteYaml struct {
	Methods   []string              `yaml:"methods,omitempty,flow"`
	Params    map[string]*ParamYaml `yaml:"params,omitempty,flow"`
	Callbacks []string              `yaml:"callbacks,flow"`
}

type route struct {
	methods   []httpMethod
	callbacks []string
	params    []*routeParameter
}

type routeParameter struct {
	pType    httpParameterType
	regex    string
	required bool
}

func newMethods(ms []string) ([]httpMethod, error) {
	//get is the default method when not defined
	if len(ms) == 0 {
		return []httpMethod{getMethod}
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
	pType, err := newHttpParameterType(p.Type)
	if err != nil {
		return nil, err
	}
	return &routeParameter{
		pType:    pType,
		regex:    p.Regex,
		required: p.Required,
	}
}

func newRoute(r *RouteYaml) (*route, error) {
	if len(r.callbacks) == 0 {
		return nil, fmt.Errorf("at least one callback is required")
	}
	methods, err := newMethods(r.Methods)
	if err != nil {
		return nil, err
	}
	params := make([]*routeParameter, len(r.Params))
	for i, p := range r.Params {
		tmp, err := newParam(p)
		if err != nil {
			return nil, err
		}
		params[i] = tmpe
	}
	return &route{
		methods:   methods,
		callbacks: callbacks,
		params:    params,
	}, nil
}

func loadRouteYaml(r io.Reader) (map[string]*RouteYaml, error) {
	rawBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	yamlData := make(map[string]*RouteYaml)
	if err = yaml.Unmarshal(rawBytes, &yamlData); err != nil {
		return nil, err
	}
	//fmt.Printf("yamlData: %+v\n", yamlData["/"])
	return yamlData, nil
}

func LoadRoutes(r io.Reader) (map[string]Route, error) {
	return nil, nil
}
