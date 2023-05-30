package bilibili_http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// HandleFunc 视图函数签名
type HandleFunc func(ctx *Context)

// MiddlewareChains 中间件责任链
type MiddlewareChains []MiddlewareHandleFunc

// 在代码层面，或者说是编译阶段，就判断*HTTPServer有没有实现server接口
var _ server = &HTTPServer{}

type server interface {
	// Handler 硬性要求，必须组合http.Handler
	http.Handler
	// Start 启动服务
	Start(addr string) error
	// Stop 关闭服务
	Stop() error
	// addRouter 注册路由：这个是一个非常核心的API，表示他不能被外界使用【外界：开发者】
	// 造一些衍生API供开发者使用
	addRouter(method string, pattern string, handleFunc HandleFunc, middlewareChains ...MiddlewareHandleFunc)
}

type HTTPOption func(h *HTTPServer)

type HTTPServer struct {
	srv  *http.Server
	stop func() error
	// 第一个版本的路由树
	// routers 临时存放路由的位置
	// routers map[string]HandleFunc
	// 第二个版本的路由树——前缀树
	// *router
	router *router
	// *router和 router *router两者之间的关系？
	// 前者是直接嵌套 当前结构体可以直接通过结构体对象调用*router中的方法
	// 后者是组装，如果想要通过当前结构体调用*router的方法，是这样使用 httpServer.router.addRouter(....)
	// 嵌套还有一个好处，就是被嵌套的结构体中实现的方法可以当作是当前结构体实现的方法
	// 这里路由组其实是一个根路由组
	*RouterGroup

	// groups 维护整个项目所有的路由组
	groups []*RouterGroup
}

/*
暂时的路由设计
{
	"GET-login": HandleFun1,
	"POST-login": HandleFunc2,
	...
	...
}

*/

func WithHTTPServerStop(fn func() error) HTTPOption {
	return func(h *HTTPServer) {
		if fn == nil {
			fn = func() error {
				fmt.Println("1231231312")
				quit := make(chan os.Signal)
				signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
				<-quit
				log.Println("Shutdown Server ...")

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				// 关闭之前：需要做某些操作
				if err := h.srv.Shutdown(ctx); err != nil {
					log.Fatal("Server Shutdown:", err)
				}
				// 关闭之后，需要做某些操作
				select {
				case <-ctx.Done():
					log.Println("timeout of 5 seconds.")
				}
				return nil
			}
		}
		h.stop = fn
	}
}

func NewHTTP(opts ...HTTPOption) *HTTPServer {
	// HTTPServer和RouterGroup相互嵌套的初始化是在这里实现的
	rg := newRouterGroup()
	h := &HTTPServer{
		router:      newRouter(),
		RouterGroup: rg,
	}
	rg.engine = h
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// ServeHTTP 接收请求，转发请求
// 接收请求：接收前端传过来的请求
// 转发请求：转发前端过来的请求到咱们的框架中
// ServeHTTP方法向前对接前端请求，向后对接咱们的框架
func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// 1. 匹配路由
	n, params, ok := h.router.getRouter(r.Method, r.URL.Path)
	if !ok || n.handleFunc == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 NOT FOUND肯定失败"))
		return
	}
	// 2. 构造当前请求的上下文
	c := NewContext(w, r)
	c.params = params
	fmt.Printf("request %s - %s\n", c.Method, c.Pattern)
	// 将项目全局的中间件注册好
	mids := []MiddlewareHandleFunc{flush(), recovery()}
	// 搜集当前请求的所有中间件方法——路由组身上的中间件
	// mids = append(mids, h.filterMiddlewares(c.Pattern)...)
	gms := h.filterMiddlewares(c.Pattern)
	if len(gms) != 0 {
		mids = append(mids, gms...)
	}
	//if len(mids) == 0 {
	//	// 若当前请求上没有配备任何中间件，就需要创建一个mids，用来维护所有的中间件
	//	// 为什么？
	//	//
	//	mids = make([]MiddlewareHandleFunc, 0)
	//}

	// 将当前匹配大的视图节点中的中间件全部添加到mids切片中——当前视图身上的中间件
	mids = append(mids, n.middlewareChains...)

	// 重头：如何构建出类似这样的代码？
	handleFunc := n.handleFunc
	for i := len(mids) - 1; i >= 0; i-- {
		handleFunc = mids[i](handleFunc)
	}
	// 到这里之后，handleFunc其实就是mids[0]
	// 2. 转发请求
	handleFunc(c) // 这里是执行用户的视图函数
	// c.flashDataToResponse() // 大功告成
}

/*
如果有中间件，咱们需要将匹配到的视图函数添加到中间件切片中
如果没有，咱们也需要将匹配到的视图函数添加到中间件切片中
*/

// filterMiddlewares 匹配当前URL所对应的所有中间件
func (h *HTTPServer) filterMiddlewares(pattern string) []MiddlewareHandleFunc {
	// pattern /login
	// 先明确，中间件放在哪里？
	mids := make([]MiddlewareHandleFunc, 0)
	for _, group := range h.groups {
		if strings.HasPrefix(pattern, group.prefix) {
			// strings.HasPrefix("asuihjdfocrenifgv", "") => True
			// pattern = /v1/login
			/*
				[
					{prefix: "", middlewares:[mid1, mid2]},
					{prefix: "/v1", middlewares:[mid1, mid2]},
					{prefix: "/v2", middlewares:[mid1, mid2]},
				]
			*/
			mids = append(mids, group.middlewares...)
		}
	}
	return mids
}

/*
先明确，中间件放在哪里？
目前来说，中间件是放在每个路由组中
我们现在是在HTTPServer身上，拿不到所有的路由组信息
所以我们就像能不能在HTTPServer身上维护整个项目所有的路由组？
又所有的路由组，就表示能够拿到整个项目中所有的中间件

*/

// Start 启动服务
func (h *HTTPServer) Start(addr string) error {
	h.srv = &http.Server{
		Addr:    addr,
		Handler: h,
	}
	return h.srv.ListenAndServe()
}

// Stop 停止服务
func (h *HTTPServer) Stop() error {
	return h.stop()
}

// addRouter 注册路由
// 注册路由的时机：就是项目启动的时候注册，项目启动之后就不能注册了。
// 问题一：注册的路由放在那里？
//func (h *HTTPServer) addRouter(method string, pattern string, handleFunc HandleFunc) {
//	// 构建唯一的key
//	key := fmt.Sprintf("%s-%s", method, pattern)
//	fmt.Printf("add router %s - %s\n", method, pattern)
//	h.routers[key] = handleFunc
//}

//// GET GET请求
//func (h *HTTPServer) GET(pattern string, handleFunc HandleFunc) {
//	h.addRouter(http.MethodGet, pattern, handleFunc)
//}
//
//// POST GET请求
//func (h *HTTPServer) POST(pattern string, handleFunc HandleFunc) {
//	h.addRouter(http.MethodPost, pattern, handleFunc)
//}
//
//// DELETE GET请求
//func (h *HTTPServer) DELETE(pattern string, handleFunc HandleFunc) {
//	h.addRouter(http.MethodDelete, pattern, handleFunc)
//}
//
//// PUT GET请求
//func (h *HTTPServer) PUT(pattern string, handleFunc HandleFunc) {
//	h.addRouter(http.MethodPut, pattern, handleFunc)
//}

// 一个Server需要什么功能
// 1. 启动
// 2. 关闭
// 3. 注册路由

// 为什么要抽象这个server呢？
// 有些网站是走https协议的，那有些网站是走http
// 为了兼容性

//func main() {
//	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
//		// 业务逻辑
//	})
//	//http.ListenAndServe(":8080", nil)
//	//http.ListenAndServeTLS()
//	h := NewHTTP(WithHTTPServerStop(nil))
//	go func() {
//		err := h.Start(":8080")
//		if err != nil && err != http.ErrServerClosed {
//			panic("启动失败")
//		}
//	}()
//	err := h.Stop()
//	if err != nil {
//		panic("关闭失败")
//	}
//	g := gin.Default()
//	g.GET("/login", func())
//	g.POST("/register", func() {})
//}
