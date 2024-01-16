// TODO logging
package server

import (
	"fmt"
	"io"
	"landtitle/util"
	"net/http"
	"net/url"
	"strings"
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
	defMethod             = getMethod
)

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

func (m *myHandler) buildDynamicParameters(
	path string,
) (map[string]string, error) {
	toRet := make(map[string]string)
	if m.dynamicPaths == nil || len(m.dynamicPaths) == 0 {
		return toRet, nil
	}
	subPaths := strings.Split(strings.Trim(path, "/"), "/")
	if int(m.dynamicPathIndex) >= len(subPaths) {
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
		toRet[m.dynamicPaths[i]] = p
	}
	return toRet, nil
}

func (m *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	validMethod := false
	for _, method := range m.route.methods {
		if string(method) == r.Method {
			validMethod = true
			break
		}
	}
	if !validMethod {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}
	subPaths := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if int(m.dynamicPathIndex+1) >= len(subPaths) || len(subPaths) > pathBits {
		http.Error(
			w, "index out of range for dynamic paths",
			http.StatusNotAcceptable,
		)
		return
	}
	parameterValues := make(map[string]string)
	var param *routeParameter
	var ok bool
	var errorCode int
	var message string
	var err error
	if len(m.dynamicPaths) > 0 {
		dynamicParts := subPaths[m.dynamicPathIndex:]
		for i, p := range dynamicParts {
			if i < len(m.dynamicPaths) {
				if param, ok = m.route.params[m.dynamicPaths[i]]; !ok {
					errorCode = http.StatusBadRequest
					message = "index out of range for dynamic paths"
					goto doError
				}
				if param == nil {
					//shouldn't ever see this
					errorCode = http.StatusInternalServerError
					message = "couldn't find parameter in parameter map"
					goto doError
				}
				parameterValues[m.dynamicPaths[i]] = p
			}
		}
	}
	r.ParseForm()
	for pName, v := range r.Form {
		if len(v) > 1 {
			errorCode = http.StatusBadRequest
			message = "only a single request parameter is supported"
			goto doError
		}
		if param, ok = m.route.params[pName]; !ok {
			errorCode = http.StatusBadRequest
			message = "index out of range for dynamic paths"
			goto doError
		}
		if param == nil {
			//shouldn't ever see this
			errorCode = http.StatusInternalServerError
			message = "couldn't find parameter in parameter map"
			goto doError
		}
		parameterValues[pName] = v[0]
	}
	if err = m.route.params.validate(parameterValues); err != nil {
		errorCode = http.StatusBadRequest
		message = fmt.Sprintf("parameter not valid, error: '%s'", err)
		goto doError
	}
	for _, callback := range m.callbacks {
		if ok, err = callback(parameterValues, w, r); !ok || err != nil {
			//TODO logging
			return
		}
	}
doError:
	http.Error(w, message, errorCode)
	return
}

func getPathHandler(
	path string,
	rte *route,
	callbacks map[string][]Callback,
) (http.Handler, error) {
	//handlePath := getHandlePath(path)
	//return
	return nil, nil
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
	}
	var dynamicPaths []string = nil
	var dynamicPathIndex uint8
	for i, p := range strings.Split(path, "/") {
		if dynamicPathRegex.MatchString(p) {
			if dynamicPaths == nil {
				dynamicPathIndex = uint8(i) - 1
				dynamicPaths = make([]string, 0)
			}
			dynamicPaths = append(dynamicPaths, strings.Trim(p, "{}"))
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
func getHandlePath(path string) string {
	toRet := strings.Trim(path, "/")
	subPaths := strings.Split(toRet, "/")
	isDynamic := false
	var subPath string
	var i int
	for i, subPath = range subPaths {
		if dynamicPathRegex.MatchString(subPath) {
			isDynamic = true
			break
		}
	}
	if isDynamic {
		tmp := (new(url.URL)).JoinPath(subPaths[0:i]...)
		toRet = fmt.Sprintf("%s/", tmp.Path)
	}
	return fmt.Sprintf("/%s", toRet)
}

func NewServer(
	routes io.Reader, callbacks map[string]Callback, port uint,
) (Server, error) {
	loadedRoutes, err := loadRoutes(routes)
	if err != nil {
		return nil, err
	}
	pathHandlers := make(map[string]*myHandler)
	for path, rte := range loadedRoutes {
		handlePath := getHandlePath(path)
		handler, err := newHandler(path, rte, callbacks)
		if err != nil {
			return nil, err
		}
		pathHandlers[handlePath] = handler
	}
	return &server{pathHandlers: pathHandlers}, nil
}
