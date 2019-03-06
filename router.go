package httprouter

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

var StructTag = "json"

var DefaultRouter = Router{}

var ErrSubPath = errors.New("route subpath is invalid:%s")
var ErrParamExcessive = errors.New("handler function request param excessive,just have a one param struct")
var ErrStructNotStore = errors.New("this struct not call StoreStruct function,struct name is:%s")
var ErrValueType = errors.New("value type must pointer")
var ErrFieldTag = errors.New("invalid struct field tag, tag type is:%s")
var ErrNotSupportKind = errors.New("input not support type,now must support map")
var ErrReturnResultNum = errors.New("return result num just one")

var structStore sync.Map

type ParamRouteHandle struct {
	handle Handle
	params []string
}

type Router struct {
	route      map[string]Handle
	rootPath   string
	static     map[string][]byte
	next       []func(next http.Handler) http.Handler
	paramRoute map[string]ParamRouteHandle
	PanicHandle func(http.ResponseWriter,*http.Request)
}

//New function instance Router object
func New(rootPath string) *Router {
	rootPath = pathHandler(rootPath)
	r := &Router{
		rootPath:   rootPath,
		route:      make(map[string]Handle),
		next:       make([]func(next http.Handler) http.Handler, 0, 1),
	}
	r.paramRoute = make(map[string]ParamRouteHandle)
	structStore = sync.Map{}
	return r
}

//SubRouter Method is to add sub Router
func (r *Router) SubRouter(subPath string, f func(r *Router) *Router) {
	subPath = pathHandler(subPath)
	if subPath == "" {
		panic(fmt.Errorf(ErrSubPath.Error(), "subPath is nil").Error())
	}
	subR := &Router{
		route:r.route,
		rootPath :r.rootPath + subPath,
		next:r.next,
		paramRoute:r.paramRoute,
	}
	r = f(subR)
}

//StoreStruct save struct
func StoreStruct(ss interface{}) {
	of := reflect.TypeOf(ss).Elem()
	if reflect.Ptr != of.Kind() {
		panic(ErrValueType)
	}
	for i := 0; i < of.NumField(); i++ {
		field := of.Field(i)
		tag := field.Tag.Get(StructTag)
		if tag == "" {
			panic(fmt.Errorf(ErrFieldTag.Error(), StructTag))
		}
	}
	typeName := of.String()
	structStore.Store(typeName, ss)
}

func newStruct(name string) (interface{}, bool) {
	elem, ok := structStore.Load(name)
	if !ok {
		return nil, false
	}
	return elem, true
}

//ServeHTTP Method implement http.Handler interface
func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ro.PanicHandle != nil {
		defer func() {
			if rcv := recover(); rcv != nil {
				ro.PanicHandle(w,r)
				return
			}
		}()
	}
	if handle, ok := ro.route[r.Method+"_"+r.RequestURI]; ok {
		of := reflect.TypeOf(handle)
		numIn := of.NumIn()
		if numIn == 2 && of.In(0).String() == "http.ResponseWriter" && of.In(1).String() == "*http.Request" {
			handle.(func(w http.ResponseWriter, r *http.Request))(w, r)
			return
		}
		fc := reflect.ValueOf(handle)
		if of.NumIn() > 1 {
			panic(ErrParamExcessive)
		}
		in := of.In(0)
		params := make([]reflect.Value, 1)
		if in.Kind() == reflect.Map {
			request := handleRequest(r)
			params[0] = reflect.ValueOf(request)
		} else {
			panic(ErrNotSupportKind)
		}
		callResult := fc.Call(params)
		responseHandler(callResult, w)
	}

	hand, b, params := paramPathMatch(r.Method + "_" +r.RequestURI, ro.paramRoute)
	if b {
		fc := reflect.ValueOf(hand)
		ps := make([]reflect.Value, len(params))
		for i := range params {
			ps[i] = reflect.ValueOf(params[i])
		}
		callResult := fc.Call(ps)
		responseHandler(callResult, w)
		return
	}

	if handle, ok := ro.route["/404"]; ok {
		handl := handle.(func(...interface{}))
		handl(w, r)
	} else if handle, ok := ro.route["/404.html"]; ok {
		handl := handle.(func(...interface{}))
		handl(w, r)
	} else {
		http.NotFound(w, r)
	}
}

//NextHandler Method use handler links
func (r *Router) NextHandler(handler func(next http.Handler) http.Handler) {
	r.next = append(r.next, handler)
}

//Handler function
func Handler(r *Router) http.Handler {
	if r.next == nil || len(r.next) == 0 {
		return r
	}
	first := r.next[0]
	var last http.Handler = r
	for i := len(r.next) - 1; i > 0; i-- {
		if i == len(r.next)-1 {
			last = r.next[i](r)
		} else {
			last = r.next[i](last)
		}
	}
	return first(last)
}
