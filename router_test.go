package bilibili_http

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// TestRouterAdd 测试静态路由的注册节点功能
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

// TestRouterGet 测试静态路由的匹配节点功能
func TestRouterGet(t *testing.T) {
	testCases := []struct {
		name     string
		method   string
		pattern  string
		wantBool bool
	}{
		{
			name:     "test1",
			method:   "GET",
			pattern:  "/study/login",
			wantBool: true,
		},
		{
			name:     "test2",
			method:   "GET",
			pattern:  "/study/login/action",
			wantBool: true,
		},
		{
			name:     "test3",
			method:   "POST",
			pattern:  "/study/login",
			wantBool: true,
		},
		{
			name:     "test2",
			method:   "GET",
			pattern:  "/study/login1",
			wantBool: false,
		},
	}
	r := newRouter()
	var mockHandleFunc HandleFunc = func(ctx *Context) {}
	r.addRouter("GET", "/study/login", mockHandleFunc)
	r.addRouter("GET", "/study/login/action", mockHandleFunc)
	r.addRouter("POST", "/study/login", mockHandleFunc)
	r.addRouter("DELETE", "/study/login", mockHandleFunc)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, ok := r.getRouter(tc.method, tc.pattern)
			assert.Equal(t, tc.wantBool, ok)
			// assert.Equal(t, mockHandleFunc, n.handleFunc)
		})
	}
}

// TestRouterAdd 测试参数路由的注册节点功能
func TestRouterParamAdd(t *testing.T) {
	testCases := []struct {
		name    string
		method  string
		pattern string

		wantErr string
	}{
		{
			name:    "test1",
			method:  "GET",
			pattern: "/study/:course",
		},
	}
	r := newRouter()
	var mockHandleFunc HandleFunc = func(ctx *Context) {

	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r.addRouter(tc.method, tc.pattern, mockHandleFunc)
			// assert.PanicsWithError(t, tc.wantErr, func() {})
		})
	}
}

// TestRouterGet 测试静态路由的匹配节点功能
func TestRouterParamGet(t *testing.T) {
	testCases := []struct {
		name        string
		method      string
		addPattern  string
		findPattern string
		key         string
		value       string
		wantBool    bool
	}{
		{
			name:        "test1",
			method:      "GET",
			addPattern:  "/study/:course",
			findPattern: "/study/golang",
			key:         "course",
			value:       "golang",
			wantBool:    true,
		},
		{
			name:        "test1",
			method:      "GET",
			addPattern:  "/study/:course/action",
			findPattern: "/study/golang/action",
			key:         "course",
			value:       "golang",
			wantBool:    true,
		},
	}
	r := newRouter()
	var mockHandleFunc HandleFunc = func(ctx *Context) {}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r.addRouter(tc.method, tc.addPattern, mockHandleFunc)
			n, params, ok := r.getRouter(tc.method, tc.findPattern)
			assert.Equal(t, tc.wantBool, ok)
			if !ok {
				return
			}
			assert.Equal(t, tc.value, params[tc.key])
			// 这里的n其实是一个参数路由
			// 参数路由有一个特点：就是它的part是以 : 开头
			assert.True(t, tc.wantBool, strings.HasPrefix(n.part, ":"))
		})
	}
}
