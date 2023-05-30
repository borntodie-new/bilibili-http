package flush

import "github.com/borntodie-new/bilibili-http"

type MiddlewareBuilder struct {
}

func (m *MiddlewareBuilder) Build() bilibili_http.MiddlewareHandleFunc {
	return func(next bilibili_http.HandleFunc) bilibili_http.HandleFunc {
		return func(ctx *bilibili_http.Context) {
			defer func() {

			}()
			next(ctx)
		}
	}
}

func NewMiddleware() *MiddlewareBuilder {
	return &MiddlewareBuilder{}
}
