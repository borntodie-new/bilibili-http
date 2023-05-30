package bilibili_http

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// MiddlewareHandleFunc 中间件的函数签名
// 参数 HandleFunc 是下一次需要执行的中间件逻辑
// 返回值 HandleFunc 是当前的中间件逻辑
type MiddlewareHandleFunc func(next HandleFunc) HandleFunc

// flush 统一刷新数据到响应对象中
func flush() MiddlewareHandleFunc {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			defer func() {
				// 写入状态码
				ctx.response.WriteHeader(ctx.status)
				// 写入响应头
				for key, value := range ctx.header {
					ctx.response.Header().Set(key, value)
				}
				// 写入响应体
				_, _ = ctx.response.Write(ctx.data)
			}()
			next(ctx)
		}
	}
}

// recovery 兜底的错误恢复
func recovery() MiddlewareHandleFunc {
	return func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			// 这个defer负责hook住所有的panic错误
			defer func() {
				if err := recover(); err != nil {
					// 下面的是输出给客户端看的
					ctx.SetStatusCode(http.StatusInternalServerError)
					ctx.SetData([]byte("Server Internal Error, Please Try Again Later!"))
					// 下面的输出给开发者看的
					fmt.Println(trace(fmt.Sprintf("%s\n", err)))
					return
				}
			}()
			next(ctx)
		}
	}
}

/*
flush和recovery两个中间件的优先级问题
recovery应该是咱们整个框架的兜底操作，就是放在最前
flush是统一刷新数据的，至少也是放在比较靠前的位置
recovery《flush
*/

func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

// accessLog 框架层面记录请求或响应信息
//func accesslog() MiddlewareHandleFunc {
//	return func(next HandleFunc) HandleFunc {
//		return func(ctx *Context) {
//			defer func() {
//				// 记录我们想要保存当前请求想要保存的信息
//				// 如果我们想要记录命中的路由是什么，那从哪里拿呢？
//				// 在这个作用域中，我们能和外界取得联系的只有Context上下文，所以我们只能通过Context获取到信息
//				// 具体操作就是在Context中新增一个属性，在这里取就好
//				l := accessLog{
//					Method:  ctx.Method,
//					Pattern: ctx.Pattern,
//				}
//				data, _ := json.Marshal(l)
//				m.logFunc(string(data))
//			}()
//			next(ctx)
//		}
//	}
//}
//// accessLog 日志抽象结构体，可自定义
//// 这里只是简单的示例
//type accessLog struct {
//	Method  string `json:"method"`  // 请求的方法
//	Pattern string `json:"pattern"` // 请求的路径
//}
