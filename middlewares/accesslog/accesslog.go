package accesslog

import (
	"encoding/json"
	"fmt"
	"github.com/borntodie-new/bilibili-http"
)

type MiddlewareBuilder struct {
	logFunc func(information string)
}

func (m *MiddlewareBuilder) Build() bilibili_http.MiddlewareHandleFunc {
	return func(next bilibili_http.HandleFunc) bilibili_http.HandleFunc {
		return func(ctx *bilibili_http.Context) {
			defer func() {
				// 记录我们想要保存当前请求想要保存的信息
				// 如果我们想要记录命中的路由是什么，那从哪里拿呢？
				// 在这个作用域中，我们能和外界取得联系的只有Context上下文，所以我们只能通过Context获取到信息
				// 具体操作就是在Context中新增一个属性，在这里取就好
				l := accessLog{
					Method:  ctx.Method,
					Pattern: ctx.Pattern,
				}
				data, _ := json.Marshal(l)
				m.logFunc(string(data))
			}()
			next(ctx)
		}
	}
}

func NewMiddleware(logFunc func(information string)) *MiddlewareBuilder {
	if logFunc == nil {
		logFunc = func(information string) {
			fmt.Println(information)
		}
	}
	return &MiddlewareBuilder{
		logFunc: logFunc,
	}
}

// accessLog 日志抽象结构体，可自定义
// 这里只是简单的示例
type accessLog struct {
	Method  string `json:"method"`  // 请求的方法
	Pattern string `json:"pattern"` // 请求的路径
}
