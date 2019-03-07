package httprouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/modern-go/reflect2"
	"net/http"
	"reflect"
)

var StructTag = "json"

var DefaultRouter = Router{}

// 子路由路径无效
var ErrSubPath = errors.New("route sub path is invalid:%s")
//  Handler 函数的输入参数个数不能大于一个
var ErrParamExcessive = errors.New("handler function request param excessive,just have a one param struct，this in num is:%d")
//
var ErrValueType = errors.New("value type must pointer")
// 通过struct的Tag字段来赋值
var ErrFieldTag = errors.New("invalid struct field tag, tag type is:%s")
//
var ErrNotSupportKind = errors.New("input not support type,now must support map")
// Handler 函数的返回值只能有一个
var ErrReturnResultNum = errors.New("handler function out num max 1,this is num is:%d")
// Handler 函数的输入参数不是是指针
var ErrHandlerFunInParamType = errors.New("handler function in param kind is not ptr")
// PathVariable 方式只能使用Get方法
var ErrHandlerMethod = errors.New("PathVariable style just use on the GET method in，now use method is:%s")

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
	return r
}

//SubRouter Method is to add sub Router
//添加子路由，实现公共路径的传递
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


//ServeHTTP Method implement http.Handler interface
func (ro *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//匹配全路径
	if handle, ok := ro.route[r.Method+"_"+r.RequestURI]; ok {
		of := reflect2.TypeOf(handle).Type1()
		numIn := of.NumIn()
		if numIn == 2 && of.In(0).String() == "http.ResponseWriter" && of.In(1).String() == "*http.Request" {
			handle.(func(w http.ResponseWriter, r *http.Request))(w, r)
			return
		}
		if of.NumIn() > 1 {
			panic(fmt.Errorf(ErrParamExcessive.Error(),of.NumIn()))
		}
		params := make([]reflect.Value, 1)
		bytes, e := BodyParse.Parse(r)
		if e != nil {
			panic(e)
		}
		instance := reflect2.ConfigSafe.Type2(of.In(0)).New()
		e = json.Unmarshal(bytes, instance)
		if e != nil {
			panic(e)
		}
		params[0] = reflect.ValueOf(instance).Elem()
		callResult := reflect.ValueOf(handle).Call(params)
		responseHandler(callResult, w)
		return
	}
	//匹配参数路径,只有Get方法才能使用这种方式
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
// 按顺序添加切面函数
func (r *Router) NextHandler(handler func(next http.Handler) http.Handler) {
	r.next = append(r.next, handler)
}

//Handler function
// 主要记录http.Handler调用顺序
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
