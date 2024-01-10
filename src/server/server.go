package server

import (
	"fmt"
	"io"
	"landtitle/util"
	"net/http"
)

type httpMethod string

func newHttpMethod(m string) (*httpMethod, error) {
	methodMap := map[string]httpMethod{
		"get":  getMethod,
		"post": postMethod,
		"head": headMethod,
		"put":  putMethod,
	}
	tmp, ok := methodMap[m]
	if !ok {
		return util.Ptr(httpMethod("")),
			fmt.Errorf("unrecognized method: '%s'", m)
	}
	return &tmp, nil
}

const (
	getMethod  httpMethod = "get"
	postMethod            = "post"
	headMethod            = "head"
	putMethod             = "put"
)

type httpParameterType string

func newHttpParameterType(p string) (*httpParameterType, error) {
	parameterMap := map[string]httpParameterType{
		"number":  numberParameterType,
		"string":  stringParameterType,
		"boolean": booleanParameterType,
	}
	tmp, ok := parameterMap[p]
	if !ok {
		return util.Ptr(httpParameterType("")),
			fmt.Errorf("unrecognized parameter type: '%s'", p)
	}
	return &tmp, nil
}

const (
	numberParameterType  httpParameterType = "number"
	stringParameterType                    = "string"
	booleanParameterType                   = "boolean"
)

type Callback func(map[string]string, http.ResponseWriter) (bool, error)

type Server interface{}

func NewServer(
	routes io.Reader, callbacks map[string][]Callback,
) (Server, error) {
	return nil, nil
}
