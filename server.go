package bilibili_http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// HandleFunc 视图函数签名
type HandleFunc func(ctx *Context)

type server interface {
	// Handler 硬性要求，必须组合http.Handler
	http.Handler
	// Start 启动服务
	Start(addr string) error
	// Stop 关闭服务
	Stop() error
	// addRouter 注册路由：这个是一个非常核心的API，表示他不能被外界使用【外界：开发者】
	// 造一些衍生API供开发者使用
	addRouter(method string, pattern string, handleFunc HandleFunc)
}

type HTTPOption func(h *HTTPServer)

type HTTPServer struct {
	srv  *http.Server
	stop func() error
	// 第一个版本的路由树
	// routers 临时存放路由的位置
	// routers map[string]HandleFunc
	// 第二个版本的路由树——前缀树
	*router
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
	h := &HTTPServer{
		router: newRouter(),
	}
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
	n, params, ok := h.getRouter(r.Method, r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 NOT FOUND"))
		return
	}
	// 2. 构造当前请求的上下文
	c := NewContext(w, r)
	c.params = params
	fmt.Printf("request %s - %s\n", c.Method, c.Pattern)
	// 2. 转发请求
	n.handleFunc(c)
}

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

// GET GET请求
func (h *HTTPServer) GET(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodGet, pattern, handleFunc)
}

// POST GET请求
func (h *HTTPServer) POST(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodPost, pattern, handleFunc)
}

// DELETE GET请求
func (h *HTTPServer) DELETE(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodDelete, pattern, handleFunc)
}

// PUT GET请求
func (h *HTTPServer) PUT(pattern string, handleFunc HandleFunc) {
	h.addRouter(http.MethodPut, pattern, handleFunc)
}

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
