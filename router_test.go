package bilibili_http

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRouterAdd(t *testing.T) {
	testCases := []struct {
		name    string
		method  string
		pattern string

		wantErr string
	}{
		{
			name:    "test1",
			method:  "GET",
			pattern: "/study/golang",
		},
		{
			name:    "test2",
			method:  "GET",
			pattern: "study/golang",
			wantErr: "web: 路由必须 / 开头",
		},
		{
			name:    "test2",
			method:  "GET",
			pattern: "study/golang/",
			wantErr: "web: 路由不准 / 结尾",
		},
		{
			name:    "test2",
			method:  "GET",
			pattern: "",
			wantErr: "web: 路由不能为空",
		},
		{
			name:    "test2",
			method:  "GET",
			pattern: "/study//golang",
			wantErr: "web: 路由不能来连续出现 / ",
		},
	}
	r := newRouter()
	var mockHandleFunc HandleFunc = func(ctx *Context) {

	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r.addRouter(tc.method, tc.pattern, mockHandleFunc)
			assert.PanicsWithError(t, tc.wantErr, func() {})
		})
	}
}
