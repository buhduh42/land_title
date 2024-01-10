package server

import (
	"encoding/json"
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
	params    map[string]*routeParameter
}

func (r *route) String() string {
	toPrint, _ := json.MarshalIndent(r, "", "  ")
	return string(toPrint)
}

type routeParameter struct {
	pType    httpParameterType `json:"type"`
	regex    string
	required bool
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
	pType, err := newHttpParameterType(p.Type)
	if err != nil {
		return nil, err
	}
	if pType == nil {
		return nil, fmt.Errorf("parameter type was set to nil for '%s'", p.Type)
	}
	return &routeParameter{
		pType:    *pType,
		regex:    p.Regex,
		required: p.Required,
	}, nil
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

func loadRouteYaml(r io.Reader) (map[string]*RouteYaml, error) {
	rawBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	yamlData := make(map[string]*RouteYaml)
	if err = yaml.Unmarshal(rawBytes, &yamlData); err != nil {
		return nil, err
	}
	return yamlData, nil
}

func LoadRoutes(r io.Reader) (map[string]*Route, error) {
	routeYaml, err := loadRouteYaml(r)
	if err != nil {
		return nil, err
	}
	toRet := make(map[string]*Route)
	for k, v := range routeYaml {
		tmp, err := newRoute(v)
		if err != nil {
			return nil, err
		}
		rte := Route(*tmp)
		toRet[k] = &rte
	}
	return toRet, nil
}
