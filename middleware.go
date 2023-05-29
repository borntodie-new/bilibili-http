package bilibili_http

// MiddlewareHandleFunc 中间件的函数签名
// 参数 HandleFunc 是下一次需要执行的中间件逻辑
// 返回值 HandleFunc 是当前的中间件逻辑
type MiddlewareHandleFunc func(next HandleFunc) HandleFunc
