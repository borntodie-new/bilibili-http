package bilibili_http

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func Logger() MiddlewareHandleFunc {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			ctime := time.Now()
			fmt.Printf("请求进来的时间: %v\n", ctime.Format("2006-01-02 15:04:05"))
			time.Sleep(time.Second * 3)
			next(ctx)
			fmt.Printf("请求走的时间：%v, 总计耗时：%d \n", time.Now().Format("2006-01-02 15:04:05"), time.Since(ctime).Milliseconds())
		}
	}
}

func TestHTTP_Start(t *testing.T) {
	h := NewHTTP()
	h.GET("/study/login", func(ctx *Context) {
		// ctx.response.Write([]byte("静态路由 " + ctx.Pattern))
		ctx.TEXT(http.StatusOK, fmt.Sprintf("静态路由：%s", ctx.Pattern))
	})
	h.GET("/study/:course", func(ctx *Context) {
		// ctx.response.Write([]byte("参数路由 " + ctx.Pattern + "    " + ctx.params["course"]))
		course, err := ctx.Params("course")
		if err != nil {
			ctx.TEXT(http.StatusNotFound, "参数错误")
			return
		}
		ctx.TEXT(http.StatusOK, fmt.Sprintf("参数路由：%s  %s", ctx.Pattern, course))
	})
	h.GET("/assets/*filepath", func(ctx *Context) {
		// ctx.response.Write([]byte("通融符路由" + ctx.Pattern + "    " + ctx.params["filepath"]))
		filepath, err := ctx.Params("filepath")
		if err != nil {
			ctx.TEXT(http.StatusNotFound, "参数错误")
			return
		}
		ctx.TEXT(http.StatusOK, fmt.Sprintf("通融符路由：%s  %s", ctx.Pattern, filepath))
	})

	h.GET("/json", func(ctx *Context) {
		ctx.JSON(http.StatusOK, H{
			"code": 200,
			"msg":  "请求成功",
			"data": []string{
				"A", "B", "C",
			},
		})
	})
	h.GET("/html", func(ctx *Context) {
		ctx.HTML(http.StatusOK, `<h1 style="color:red;">hello world</h1>`)
	})

	h.GET("/query", func(ctx *Context) {
		username, err := ctx.Query("username")
		if err != nil {
			ctx.SetStatusCode(http.StatusNotFound)
			return
		}
		password, err := ctx.Query("password")
		if err != nil {
			ctx.SetStatusCode(http.StatusNotFound)
			return
		}
		ctx.JSON(http.StatusOK, H{
			"code":     200,
			"msg":      "请求成功",
			"username": username,
			"password": password,
		})
	})

	h.POST("/body", func(ctx *Context) {
		type User struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		var user User
		err := ctx.BindJSON(&user)
		if err != nil {
			ctx.TEXT(http.StatusNotFound, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, H{
			"code":     200,
			"msg":      "请求成功",
			"username": user.Username,
			"password": user.Password,
		})
	})
	v1 := h.Group("/v1")
	v1.Use(Logger())
	{
		v1.GET("/login", func(ctx *Context) {
			ctx.HTML(http.StatusOK, fmt.Sprintf(`<h1 style="color: red;">%s</h1>`, ctx.Pattern))
		})
		v1.POST("/register", func(ctx *Context) {
			ctx.HTML(http.StatusOK, fmt.Sprintf(`<h1 style="color: red;">%s</h1>`, ctx.Pattern))
		})
	}
	v2 := h.Group("/v2")
	{
		order := v2.Group("/order")
		order.GET("/xxx", func(ctx *Context) {
			ctx.HTML(http.StatusOK, fmt.Sprintf(`<h1 style="color: red;">%s</h1>`, ctx.Pattern))
		})
	}
	v3 := h.Group("/v3")
	v3.Use(Logger())
	{
		v3.GET("/login", func(ctx *Context) {
			ctx.TEXT(http.StatusOK, fmt.Sprintf("请求成功：%s", ctx.Pattern))
		}, func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("大家好1，我来了哈")
				if ctx == nil {
					ctx.SetStatusCode(http.StatusNotFound)
					return
					//panic()
				}
				panic("手动panic")
				next(ctx)
				fmt.Println("大家好1，我走了哈")
			}
		}, func(next HandleFunc) HandleFunc {
			return func(ctx *Context) {
				fmt.Println("大家好2，我来了哈")
				next(ctx)
				fmt.Println("大家好2，我走了哈")
			}
		})
		v3.GET("/register", func(ctx *Context) {
			ctx.TEXT(http.StatusOK, fmt.Sprintf("请求成功： %s", ctx.Pattern))
		})
	}
	err := h.Start(":8080")
	if err != nil {
		panic(err)
	}
}

/*
现在这种情况是什么原因呢？
是因为响应体里面的数据没有正确写入

*/
