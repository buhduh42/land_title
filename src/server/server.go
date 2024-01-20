// TODO logging
package server

import (
	"fmt"
	"io"
	"landtitle/util"
	"net/http"
	"net/url"
	"strings"

	logger "github.com/buhduh/go-logger"
)

type httpMethod string

const (
	getMethod  httpMethod = "get"
	postMethod            = "post"
	headMethod            = "head"
	putMethod             = "put"
	defMethod             = getMethod
)

func newHttpMethod(m string) (*httpMethod, error) {
	toRet := httpMethod(m)
	switch toRet {
	case getMethod:
		fallthrough
	case postMethod:
		fallthrough
	case headMethod:
		fallthrough
	case putMethod:
		return util.Ptr(toRet), nil
	}
	return nil, fmt.Errorf("unrecognized http method: '%s'", m)
}

type Callback func(
	map[string]string,
	http.ResponseWriter,
	*http.Request,
) (bool, error)

// intended to block thread it's on
type Server interface {
	StartServer(int) error
}

type server struct {
	pathHandlers map[string]*myHandler
}

func (s *server) StartServer(port int) error {
	for path, handler := range s.pathHandlers {
		http.Handle(path, handler)
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

const pathBits int = 255

type myHandler struct {
	callbacks    []Callback
	dynamicPaths []string
	//NOTE we probably got problems if there are more than 255 subpaths
	//is that a security flaw if someone throws 255 slashes at a URL?
	//NOTE pathBits needs to be 2^bitcount of below
	dynamicPathIndex uint8
	route            *route
}

// TODO, rewrite ServerHTTP using this to break ServeHTTP up
func (m *myHandler) buildDynamicParameters(
	path string,
) (map[string]string, error) {
	toRet := make(map[string]string)
	if m.dynamicPaths == nil || len(m.dynamicPaths) == 0 {
		return toRet, nil
	}
	subPaths := strings.Split(strings.Trim(path, "/"), "/")
	if int(m.dynamicPathIndex) > len(subPaths) {
		return nil, fmt.Errorf(
			"dynamic path index out of bounds for current path: '%s'",
			path,
		)
	}
	dynamicPaths := subPaths[m.dynamicPathIndex:]
	for i, p := range dynamicPaths {
		if _, ok := m.route.params[m.dynamicPaths[i]]; !ok {
			return nil, fmt.Errorf(
				"route for handler does not contain parameter for dynamic path: %s",
				m.dynamicPaths[i],
			)
		}
		if _, ok := m.route.params[m.dynamicPaths[i]]; ok {
			if m.route.params[m.dynamicPaths[i]].source&sourceURL > 0 {
				toRet[m.dynamicPaths[i]] = p
			} else {
				return nil, fmt.Errorf(
					"parameter '%s' not allowed in URL", m.dynamicPaths[i],
				)
			}
		}
	}
	return toRet, nil
}

func (m *myHandler) doQueryParameters(queryStr string) (map[string]string, error) {
	values, err := url.ParseQuery(queryStr)
	if err != nil {
		return nil, err
	}
	toRet := make(map[string]string)
	for k, v := range values {
		if len(v) > 1 {
			return nil, fmt.Errorf("only single valued query parameters allowed")
		}
		param, ok := m.route.params[k]
		if !ok {
			return nil, fmt.Errorf("%s not found in parameter values for route", k)
		}
		if sourceQuery&param.source == 0 {
			return nil, fmt.Errorf("parameter '%s' not allowed in query", k)
		} else {
			toRet[k] = v[0]
		}
	}
	return toRet, nil
}

func (m *myHandler) doFormParameters(values url.Values) (map[string]string, error) {
	toRet := make(map[string]string)
	for k, v := range values {
		if len(v) > 1 {
			return nil, fmt.Errorf("only single valued post parameters allowed")
		}
		param, ok := m.route.params[k]
		if !ok {
			return nil, fmt.Errorf("%s not found in parameter values for route", k)
		}
		if sourceForm&param.source > 0 {
			toRet[k] = v[0]
		} else {
			return nil, fmt.Errorf("parameter '%s' not allowed in form", k)
		}
	}
	return toRet, nil
}

// TODO break this up, use buildDynamicParameters
// TODO contexts?
func (m *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	myLogger.Tracef("Serving HTTP for handler:\n%+v", m)
	myLogger.Tracef("request:\n%+v", r)
	validMethod := false
	for _, method := range m.route.methods {
		toCheck := strings.ToLower(r.Method)
		if string(method) == toCheck {
			validMethod = true
			break
		}
	}
	if !validMethod {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}
	myLogger.Tracef("valid method for request found, '%s'", r.Method)
	var errorCode int
	var message string
	var ok bool
	urlParameters, err := m.buildDynamicParameters(r.URL.Path)
	var fValues, qValues, parameterValues map[string]string
	if err != nil {
		errorCode = http.StatusBadRequest
		message = "unable to build dynamic parameter map from request url"
		myLogger.Debugf(
			"dynamic parameter build failed for url: '%s', error: '%s'",
			r.URL.Path, err,
		)
		goto doError
	}
	myLogger.Tracef("urlParameters: '%v'", urlParameters)
	qValues, err = m.doQueryParameters(r.URL.RawQuery)
	if err != nil {
		myLogger.Errorf("invalid query parameters: %v", r.URL.Query)
		message = "invalid query parameters"
		errorCode = http.StatusBadRequest
		goto doError
	}
	myLogger.Tracef("query parameters: '%v'", qValues)
	r.ParseForm()
	fValues, err = m.doFormParameters(r.PostForm)
	if err != nil {
		myLogger.Errorf("invalid form parameters: %v", r.PostForm)
		message = "invalid form parameters"
		errorCode = http.StatusBadRequest
		goto doError
	}
	myLogger.Tracef("form parameters: '%v'", fValues)

	parameterValues = make(map[string]string)
	for k, v := range fValues {
		parameterValues[k] = v
	}
	for k, v := range urlParameters {
		parameterValues[k] = v
	}
	for k, v := range qValues {
		parameterValues[k] = v
	}

	if err = m.route.params.validate(parameterValues); err != nil {
		errorCode = http.StatusBadRequest
		message = fmt.Sprintf("parameter not valid, error: '%s'", err)
		goto doError
	}
	//NOTE it's POSSIBLE params wouldn't be updated between sequential
	//callback calls, can't think of a clean way to test, moving on
	for _, callback := range m.callbacks {
		myLogger.Tracef("calling callback with parameters: %+v", parameterValues)
		ok, err = callback(parameterValues, w, r)
		myLogger.Debugf("callback returned %t", ok)
		if !ok || err != nil {
			//TODO logging, leaving callbacks to do their writing
			if err != nil {
				myLogger.Errorf("callback returned error: '%s'", err)
			}
			return
		}
	}
	return
doError:
	http.Error(w, message, errorCode)
	myLogger.Errorf(
		"ServerHTTP failed with error message: '%s', http code: %d",
		message, errorCode,
	)
}

func newHandler(
	path string, rte *route, callbackMap map[string]Callback,
) (*myHandler, error) {
	callbacks := make([]Callback, len(rte.callbacks))
	var ok bool
	for i, cb := range rte.callbacks {
		if callbacks[i], ok = callbackMap[cb]; !ok {
			return nil, fmt.Errorf(
				"could not find callback from callback map: '%s'", cb,
			)
		}
		myLogger.Tracef("adding callback '%s' for path '%s'", cb, path)
	}
	var dynamicPaths []string = nil
	var dynamicPathIndex uint8
	for i, p := range strings.Split(path, "/") {
		if dynamicPathRegex.MatchString(p) {
			if dynamicPaths == nil {
				dynamicPathIndex = uint8(i) - 1
				myLogger.Tracef(
					"setting dynamic path index to %d for path '%s'",
					dynamicPathIndex, path,
				)
				dynamicPaths = make([]string, 0)
			}
			toAdd := strings.Trim(p, "{}")
			dynamicPaths = append(dynamicPaths, toAdd)
			myLogger.Tracef("adding dynamic path '%s'", toAdd)
		} else if dynamicPaths != nil {
			return nil, fmt.Errorf(
				"non dynamic paths can't follow dynamic paths for path '%s'",
				path,
			)
		}
	}
	return &myHandler{
		callbacks:        callbacks,
		dynamicPaths:     dynamicPaths,
		dynamicPathIndex: dynamicPathIndex,
		route:            rte,
	}, nil
}

// When setting up routes in routes.yaml, anything with path form: {foo}
// is a dynamic path, this returns the beginning part of the path with NO
// dynamic paths to pass as the url pattern for net/http
// eg, /foo/bar/{baz}/{biff} -> /foo/bar/
// FUCK THIS FUCKING FUNCTION!!!!! so fucking finicky, stupidly harder than
// it should be
func getHandlePath(path string) string {
	myLogger.Tracef("getHandlePath('%s') called", path)
	toRet := strings.Trim(path, "/")
	subPaths := strings.Split(toRet, "/")
	myLogger.Tracef("parsing subpaths: '%v'", subPaths)
	isDynamic := false
	var subPath string
	var i int
	for i, subPath = range subPaths {
		if dynamicPathRegex.MatchString(subPath) {
			myLogger.Tracef("discovered dynamic path for subpath: '%s'", subPath)
			isDynamic = true
			break
		}
	}
	if isDynamic {
		toRet = (new(url.URL)).JoinPath(subPaths[0:i]...).Path
		if i > 0 {
			toRet = fmt.Sprintf("/%s/", toRet)
		}
	} else {
		toRet = fmt.Sprintf("/%s", toRet)
	}
	myLogger.Tracef("setting handler path to '%s'", toRet)
	return toRet
}

func AddGlobalLogger(pLogger logger.Logger) {
	addLogger(pLogger)
}

func NewServer(
	routes io.Reader, callbacks map[string]Callback,
) (Server, error) {
	loadedRoutes, err := loadRoutes(routes)
	if err != nil {
		myLogger.Errorf("could not load routes with error: '%s'", err)
		return nil, err
	}
	pathHandlers := make(map[string]*myHandler)
	for path, rte := range loadedRoutes {
		handlePath := getHandlePath(path)
		handler, err := newHandler(path, rte, callbacks)
		if err != nil {
			return nil, err
		}
		var ok bool
		if _, ok = pathHandlers[handlePath]; ok {
			return nil, fmt.Errorf(
				"multiple handlers assigned to same handle path: '%s'", handlePath,
			)
		}
		myLogger.Tracef("handler exists: %t", ok)
		pathHandlers[handlePath] = handler
		myLogger.Tracef("adding handler for path '%s'", handlePath)
	}
	return &server{pathHandlers: pathHandlers}, nil
}
