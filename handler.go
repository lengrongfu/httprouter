package httprouter

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrUriRepeat = errors.New("uri path is repeat,exist path is:%s")

type Handle interface{}

func (r *Router) Get(path string, handle Handle) {
	r.handle(path, http.MethodGet, handle)
}

func (r *Router) Head(path string, handle Handle) {
	r.handle(path, http.MethodHead, handle)
}

func (r *Router) Post(path string, handle Handle) {
	r.handle(path, http.MethodHead, handle)
}

func (r *Router) Put(path string, handle Handle) {
	r.handle(path, http.MethodPut, handle)
}

func (r *Router) Patch(path string, handle Handle) {
	r.handle(path, http.MethodPatch, handle)
}

func (r *Router) Delete(path string, handle Handle) {
	r.handle(path, http.MethodDelete, handle)
}

func (r *Router) Options(path string, handle Handle) {
	r.handle(path, http.MethodOptions, handle)
}

func (r *Router) Static(path string, staticPath string) {
	if _, ok := r.static[path]; ok {
		panic(fmt.Errorf(ErrUriRepeat.Error(), path))
	}
}

func (r *Router) handle(path, method string, handle Handle) {
	path = pathHandler(path)
	if path == "" {
		panic(errors.New("path is not nil"))
	}
	params, newPath := paramPathHandler(path)
	if len(params) > 0 {
		key := method + "_" + r.rootPath + newPath
		paramRouter := ParamRouteHandle{
			handle: handle,
			params: params,
		}
		r.paramRoute[key] = paramRouter
		return
	}
	path = method + "_" + r.rootPath + path
	if _, ok := r.route[path]; ok {
		panic(fmt.Errorf(ErrUriRepeat.Error(), path))
	}
	r.route[path] = handle
}

//pathHandler function handler uri path
func pathHandler(path string) string {
	if path != "" && path[0] != '/' {
		path = "/" + path
	}
	if path != "" && path[len(path)-1] == '/' {
		path = path[:len(path)-2]
	}
	return path
}

func paramPathHandler(path string) ([]string, string) {
	paths := strings.Split(path[1:], "/")
	if len(paths) == 0 {
		return nil, path
	}
	params := make([]string, 0, 1)
	newPath := "/"
	for i := 0; i < len(paths); i++ {
		path := paths[i]
		if path[0] == '{' && path[len(path)-1] == '}' {
			params = append(params, path[1:len(path)-1])
		} else {
			newPath += path + "/"
		}
	}
	if newPath != "" && newPath[len(newPath)-1] == '/' {
		newPath = newPath[:len(newPath)-1]
	}
	return params, newPath
}

func paramPathMatch(path string, pm map[string]ParamRouteHandle) (Handle, bool, []string) {
	for k, _ := range pm {
		if strings.Contains(path, k) {
			path := strings.Replace(path, k, "", -1)
			paramNum := strings.Split(path[1:], "/")
			paramRouteHandle, ok := pm[k]
			if ok && len(paramNum) == len(paramRouteHandle.params) {
				return paramRouteHandle.handle, true, paramNum
			}
		}
	}
	return nil, false, nil
}
