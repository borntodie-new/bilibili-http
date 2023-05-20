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

func TestRouterParamsAddB(t *testing.T) {
	r := newRouter()
	mockHandleFunc := func(ctx *Context) {}
	// 正常注册的路由
	r.addRouter("GET", "/:lang/intro", mockHandleFunc)
	r.addRouter("GET", "/:lang/tutorial", mockHandleFunc)
	r.addRouter("GET", "/:lang/doc", mockHandleFunc)
	r.addRouter("GET", "/about", mockHandleFunc)
	r.addRouter("GET", "/p/blog", mockHandleFunc)
	r.addRouter("GET", "/p/related", mockHandleFunc)

	//// 参数路由冲突
	// r.addRouter("GET", "/:action/user", mockHandleFunc) // 报错
	//r.addRouter("GET", "/study/:action/user", mockHandleFunc) // 正常
	//r.addRouter("GET", "/study/:action", mockHandleFunc)      // 正常
	//
	//// 正常注册通配符路由
	//r.addRouter("GET", "/assets/*filepath", mockHandleFunc)
	//
	//// 通配符路由冲突
	r.addRouter("GET", "/assets/*filename", mockHandleFunc)
	//
	//// 参数路由和通配符路由冲突
	r.addRouter("GET", "/assets/:course", mockHandleFunc)
}

func TestRouterGetB(t *testing.T) {

	testCases := []struct {
		name    string
		method  string
		pattern string

		wantBool bool
		key      string
		value    string
	}{
		{
			name:     "success",
			method:   "GET",
			pattern:  "/python/intro",
			wantBool: true,
			key:      "lang",
			value:    "python",
		},
		{
			name:     "success",
			method:   "GET",
			pattern:  "/golang/doc",
			wantBool: true,
			key:      "lang",
			value:    "golang",
		},
		{
			name:     "success",
			method:   "GET",
			pattern:  "/p/related",
			wantBool: true,
		},
		{
			name:     "success",
			method:   "GET",
			pattern:  "/study/golang/user",
			wantBool: true,
			key:      "action",
			value:    "golang",
		},
		{
			name:     "success",
			method:   "GET",
			pattern:  "/assets/css/index.css",
			wantBool: true,
			key:      "filepath",
			value:    "css/index.css",
		},
		{
			name:     "success",
			method:   "GET",
			pattern:  "/study/golang/user/asdio",
			wantBool: false,
		},
	}
	r := newRouter()
	mockHandleFunc := func(ctx *Context) {}
	r.addRouter("GET", "/:lang/intro", mockHandleFunc)
	r.addRouter("GET", "/:lang/tutorial", mockHandleFunc)
	r.addRouter("GET", "/:lang/doc", mockHandleFunc)
	r.addRouter("GET", "/about", mockHandleFunc)
	r.addRouter("GET", "/p/blog", mockHandleFunc)
	r.addRouter("GET", "/p/related", mockHandleFunc)
	r.addRouter("GET", "/study/:action/user", mockHandleFunc) // 正常
	r.addRouter("GET", "/assets/*filepath", mockHandleFunc)   // 正常

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, params, ok := r.getRouter(tc.method, tc.pattern)
			assert.Equal(t, tc.wantBool, ok)
			if !ok {
				return
			}
			assert.Equal(t, tc.value, params[tc.key])
		})
	}
}
