# httprouter

> 项目中难免会有一些问题，文档写的也不够详情，如果有问题，请提交`issues`，如果遇到`bug`,要是大佬自己愿意顺便动动手,解决了之后欢迎提交`pr`。


> 目前已经有很多`go`开源的路由处理库了，为什么我还要重复造轮子呢？因为我是使用`spring boot`的用户，我更喜欢里面的路由处理方式，如果是从`spring boot`转过来的开发人员，应该会更喜欢这种风格，不需要自己处理`ResponseWriter`和`Request`对象。



## 已开发功能
- 公共路由设置
- PathVariable 使用方式
- 自定义函数做接口处理
- 拦截器自定义实现和注册

## 未开发功能
### 一期
- 请求体转换为`struct`类型
- 静态文件压缩
- `swagger josn` 支持


### 二期
- jwt 功能添加
- 上下文管理
- 目前使用的是`map`去做的映射，如果压测处理性能不怎么好，就会进行路由算法替换。

## 使用

### 1、下载
```go
$ go get github.com/lengrongfu/httprouter
```

### 2、使用公共路由设置
```go
    import "github.com/lengrongfu/httprouter"
    //根路由
    route := httprouter.New("/api")
    //公共路由一，所有的/api/user都走这个路由
	route.SubRouter("/user", UserRouter)
    //公共路由二,所有的/api/login都走这个路由
	route.SubRouter("/login", LoginRouter)
    
    //定义user内部路由
    func UserRouter(r *Router) *Router {
    	r.Get("/list", List)
    	r.Post("/add/student", StudentAdd)
    	r.Get("/get/student/{name}", GetStudent)
    	return r
    }
    //定义login内部路由
    func LoginRouter(r *Router) *Router {
    	r.Get("/check", Check)
    	r.Get("/login/{name}/{password}",login)
    	return r
    }
```

### 3、使用PathVariable方式
使用`spring boot`类似的`restful`的`PathVariable`方式
```go
//添加name和password这两个字段到URI中，处理函数为login
r.Get("/login/{name}/{password}",login)

//接受参数为uri中的值，参数定义顺序要和uri中定义的一致，不然赋值会错误。
func login(name,password string) ReturnInfo {
	fmt.Println("name:",name)
	fmt.Println("password:",password)
	r := ReturnInfo{
		Code:1,
		Msg:"success",
		Data:name,
	}
	return r
}
```

### 4、自定义函数做接口处理
如上第三点中的`login(name,password string) ReturnInfo`函数,自定义了两个接受参数，自定义了返回类型.
函数定义有几点需要注意:
- 接受参数目前不能使用结构体接受，后续支持。
- 接受参数不能出现`...`可变参数接受.
- 接受参数可以是`map`和当值字段.
- 返回值只能有一个，可以是`map`,`struct`或基本类型。

### 5、拦截器自定义实现和注册
支持自定义注册拦截器。如下定义了三个拦截器链，调用顺序按`先注册先调用`.
```go
    route := New("/api")
	route.NextHandler(AuthHandler)
	route.NextHandler(CheckHandler)
    //添加全局异常处理
	route.NextHandler(ExceptionHanler)
    
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
    
    //当有注册自定义拦截器时要调用`httprouter.Handler`方法。
    http.ListenAndServe(":8088", httprouter.Handler(route))
    //若没有自定义拦截器则可以直接使用`router`对象
    http.ListenAndServe(":8088", route)

```

## 优化目标
- 使用[modern-go/reflect2](https://github.com/modern-go/reflect2)来优化反射，提高速度。
- 补全函数的入参和出参参数类型支持。
- 目前未做性能[压测](https://github.com/julienschmidt/go-http-routing-benchmark)，等未开发功能开发完之后进行测试。
