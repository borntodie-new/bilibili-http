package bilibili_http

import (
	"testing"
)

func TestHTTP_Start(t *testing.T) {
	h := NewHTTP()
	h.GET("/study/login", func(ctx *Context) {
		ctx.response.Write([]byte("静态路由 " + ctx.Pattern))
	})
	h.GET("/study/:course", func(ctx *Context) {
		ctx.response.Write([]byte("参数路由 " + ctx.Pattern + "    " + ctx.params["course"]))
	})
	h.GET("/assets/*filepath", func(ctx *Context) {
		ctx.response.Write([]byte("通融符路由" + ctx.Pattern + "    " + ctx.params["filepath"]))
	})
	err := h.Start(":8080")
	if err != nil {
		panic(err)
	}
}
