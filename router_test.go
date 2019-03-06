package httprouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"testing"
)

type Student struct {
	Name string
	Age  int
}

func StudentAdd(student Student) string {

	return "success"
}

func StudentInsert(map[string]interface{}) string {

	return "ok"
}

func GetStudent(name string) *Student {

	return nil
}

func List(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("list"))
}

func Check(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("check success"))
}

type ReturnInfo struct {
	Code int
	Msg string
	Data interface{}
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

func ThirdRouter(r *Router) *Router {
	r.Get("/list", List)
	r.Post("/add/student", StudentAdd)
	r.Get("/get/student/{name}", GetStudent)
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

func ExceptionHanler(next http.Handler) http.Handler {
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

func TestNew(t *testing.T) {
	route := New("/api")
	route.SubRouter("/third", ThirdRouter)
	route.SubRouter("/login", LoginRouter)
	//route.Get("/404.html", Html404)
	route.NextHandler(AuthHandler)
	route.NextHandler(CheckHandler)
	route.NextHandler(ExceptionHanler)

	t.Fatal(http.ListenAndServe(":8088", Handler(route)))
}
