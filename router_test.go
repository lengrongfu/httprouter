package httprouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"testing"
)

type (
	Student struct {
		Name string `json:"name"`
		Age  int `json:"age"`
	}

	ReturnInfo struct {
		Code int `json:"code"`
		Msg string `json:"msg"`
		Data interface{} `json:"data"`
	}
)


var data map[string]Student = make(map[string]Student)

func StudentAdd(student Student) string {
	data[student.Name] = student
	return "success"
}

func GetStudent(name string) string {
	student,ok  := data[name]
	if ok {
		bytes, _ := json.Marshal(ReturnInfo{
			Code: 200,
			Msg:  "success",
			Data: student,
		})
		return string(bytes)
	}
	return ""
}

func Students() ReturnInfo {
	students := make([]Student,len(data),len(data))
	i := 0
	for k := range data {
		students[i] = data[k]
		i++
	}
	return ReturnInfo{
		Code:200,
		Msg:"success",
		Data:students,
	}
}


func Check(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("check success"))
}



func login(name,password string) ReturnInfo {
	fmt.Println("name:",name)
	fmt.Println("password:",password)
	r := ReturnInfo{
		Code:1,
		Msg:"success",
		Data:name,
	}
	panic(errors.New("exception"))
	return r
}

func Html404(w http.ResponseWriter, r *http.Request) {

}

func StudentRouter(r *Router) *Router {
	r.Post("/add/student", StudentAdd)
	r.Get("/get/student/{name}", GetStudent)
	r.Get("/students",Students)
	return r
}

func LoginRouter(r *Router) *Router {
	r.Get("/check", Check)
	r.Get("/login/{name}/{password}",login)
	return r
}

func AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("AuthHandler........")
		next.ServeHTTP(w, r)
	})
}

func CheckHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("CheckHandler........")
		next.ServeHTTP(w, r)
	})
}

func ExceptionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rcv := recover(); rcv != nil {
				ex := ReturnInfo{
					Code:-2,
					Msg:"服务器异常",
				}
				bytes, _ := json.Marshal(ex)
				w.Write(bytes)
				stack := debug.Stack()
				fmt.Println(string(stack))
				return
			}
		}()
		next.ServeHTTP(w,r)
	})
}

func JsonHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type","application/json")
		next.ServeHTTP(w,r)
	})
}

func TestNew(t *testing.T) {
	route := New("/api")
	route.SubRouter("/student", StudentRouter)
	route.SubRouter("/login", LoginRouter)
	//route.Get("/404.html", Html404)
	route.NextHandler(JsonHandler)
	route.NextHandler(AuthHandler)
	route.NextHandler(CheckHandler)
	route.NextHandler(ExceptionHandler)
	t.Fatal(http.ListenAndServe(":8088", Handler(route)))
}
