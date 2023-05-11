package bilibili_http

import (
	"testing"
)

// 注册路由是在启动服务之前就应该完成的
func Login(ctx *Context) {
	ctx.response.Write([]byte("login请求成功"))
}
func Register(ctx *Context) {
	ctx.response.Write([]byte("register请求成功"))
}
func TestHTTP_Start(t *testing.T) {
	h := NewHTTP()
	h.GET("/login", Login)
	h.POST("/register", Register)
	err := h.Start(":8080")
	if err != nil {
		panic(err)
	}
}
