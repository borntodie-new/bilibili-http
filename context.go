package bilibili_http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// H 提供一个新类型，方便操作
type H map[string]any

// Context 上下文
type Context struct {
	// 响应
	response http.ResponseWriter
	// 请求
	request *http.Request
	// Method 当前请求的方法
	Method string
	// 请求URL
	Pattern string
	// params 参数路由参数
	params map[string]string

	// 请求相关的信息
	// 1. 请求参数:
	// GET /user/:id
	// DELETE /user/:id
	// GET /user/1 获取ID是1的用户信息
	// DELETE /user/1 删除ID是1的用户信息

	// 2. 查询参数
	// http://www.baidu.com?search=周杰伦&page=1&limit=10
	// {search:周杰伦,page:1,limit:10}
	// 3. 请求体

	// http包的坑：https://juejin.cn/post/7224682976651165756

	// cacheQuery 内部维护一份查询查参数数据
	cacheQuery url.Values
	// cacheBody 内部维护一份请求体数据
	cacheBody io.ReadCloser

	// 下面是响应相关的信息
	// 1. 状态码
	status int
	// 2. 响应头
	header map[string]string
	// 3. 响应体
	data []byte
}

// Params 获取请求参数
func (c *Context) Params(key string) (string, error) {
	value, ok := c.params[key]
	if !ok {
		return "", errors.New(fmt.Sprintf("web: [%s]不存在", key))
	}
	return value, nil
}

// Query 获取查询参数
// 我怎么判断用户存的是一个空串还是就是没有值
func (c *Context) Query(key string) (string, error) {
	if c.cacheQuery == nil {
		c.cacheQuery = c.request.URL.Query()
	}
	// {search: "", limit: 10}
	// 情况1：get("age") -> ""
	// 情况2：get("search") -> ""
	value, ok := c.cacheQuery[key]
	if !ok {
		return "", errors.New(fmt.Sprintf("web: [%s]不存在", key))
	}
	return value[0], nil
}

// Form 获取请求体中的数据
// 只解析urlencoded编码格式的数据
func (c *Context) Form(key string) (string, error) {
	if c.cacheBody == nil {
		c.cacheBody = c.request.Body
	}
	// 必须先使用ParseForm方法
	if err := c.request.ParseForm(); err != nil {
		return "", err
	}
	// 从请求体中解析数据
	return c.request.FormValue(key), nil
}

// BindJSON 解析JSON格式数据的请求
func (c *Context) BindJSON(dest any) error {
	if c.cacheQuery == nil {
		c.cacheBody = c.request.Body
	}
	decoder := json.NewDecoder(c.cacheBody)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		response: w,
		request:  r,
		Method:   r.Method,
		Pattern:  r.URL.Path,
		header:   map[string]string{},
	}
}

// SetStatusCode 设置状态码
func (c *Context) SetStatusCode(code int) {
	c.status = code
}

// SetHeader 设置响应头
func (c *Context) SetHeader(key string, value string) {
	c.header[key] = value
}

// DelHeader 删除响应头
func (c *Context) DelHeader(key string) {
	delete(c.header, key)
}

// SetData 设置响应体
func (c *Context) SetData(data []byte) {
	c.data = data
}

// 所以，SetStatusCode、SetHeader、SetData就类似一些小零件，我们需要提供一些成型的方法给到用户使用
// 1. 响应JSON格式
// 2. 响应HTML格式
// 3. 响应纯文本格式
// ....

// JSON 响应JSON格式数据
func (c *Context) JSON(code int, data any) {
	c.SetStatusCode(code)
	c.SetHeader("Context-Type", "application/json")
	res, err := json.Marshal(data)
	if err != nil {
		// 现在程序存在的问题：咱们这里是直接panic
		// 那我们之前设置的状态码和响应头需要去掉吗？
		// 最好是去掉
		c.SetStatusCode(http.StatusInternalServerError)
		c.DelHeader("Context-Type")
		panic(err)
	}
	c.SetData(res)
}

// HTML 响应HTML格式数据
func (c *Context) HTML(code int, html string) {
	c.SetStatusCode(code)
	c.SetHeader("Context-Type", "text/html")
	c.SetData([]byte(html))
}

// TEXT 响应HTML格式数据
func (c *Context) TEXT(code int, text string) {
	c.SetStatusCode(code)
	c.SetHeader("Context-Type", "text/plain")
	c.SetData([]byte(text))
}

// 将数据全部写入响应中
func (c *Context) flashDataToResponse() {
	// 写入状态码
	c.response.WriteHeader(c.status)
	// 写入响应头
	for key, value := range c.header {
		c.response.Header().Set(key, value)
	}
	// 写入响应体
	c.response.Write(c.data)
}

// 注意：
// 1. 一般来说，请求对象不需要变动，直接 *http.Request
// 2. 响应对象，咱们最好是封装一个自己的response

// 上下文抽象出来就是一个请求和一个响应

//type MyResponse struct {
//	http.ResponseWriter
//	Message string
//}
//
//func (m *MyResponse) SetMsg(msg string) {
//	m.Message = msg
//}

// 疑问：我们在Context上下文中，维护一些请求相关的数据是可以理解的，因为可能有很多个视图函数需要用到这些数据
// 那为什么还需要维护一些响应数据呢？
// 比如：状态码、响应头、响应体
// 就是说，一旦将数据写入到了响应体中，下次就不能再写入了
