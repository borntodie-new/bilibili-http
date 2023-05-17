package bilibili_http

import (
	"fmt"
	"strings"
)

// router 路由树，其实应该叫路由森林
/**
{
	"GET": node{},
	"POST": node{},
	"DELETE": node{}
	....
}
根节点，咱们直接用 / 代替
**/
type router struct {
	trees map[string]*node
}

func newRouter() *router {
	return &router{trees: map[string]*node{}}
}

// addRouter 注册路由
// 注册路由有很多东西需要考虑的
// 什么样的pattern是合法的？
// method 的问题不需要考虑
// 1. pattern = ""
// 2. pattern = /asdh/asd//////
// 3. pattern = /
// 4. pattern需不需要以 / 开头
// 5. pattern需不需要以 / 结尾
// 咱们就是定死：必须 / 开头， 不准 / 结尾
// ....
// 那在这里返回error可以吗？
// 可以的，但是不好
// method = GET
// pattern = /
// handleFunc = HandleFunc()
// 意思是什么呢？就是说为 / 节点绑定一个视图函数
func (r *router) addRouter(method string, pattern string, handleFunc HandleFunc) {
	// method = GET
	// pattern = /
	fmt.Printf("add router %s - %s\n", method, pattern)
	if pattern == "" {
		panic("web: 路由不能为空")
	}
	// 获取根节点
	root, ok := r.trees[method]
	if !ok {
		// 根节点不存在
		// 1. 创建根节点
		// 2. 把根节点放到trees里面
		root = &node{
			part: "/",
		}
		r.trees[method] = root
	}
	// TODO 如果是根路由怎么办？
	if pattern == "/" {
		root.handleFunc = handleFunc
		return
	}
	if !strings.HasPrefix(pattern, "/") {
		panic("web: 路由必须 / 开头")
	}
	if strings.HasSuffix(pattern, "/") {
		panic("web: 路由不准 / 结尾")
	}
	// 切割pattern
	// /user/login => ["", "user", "login"]
	parts := strings.Split(pattern[1:], "/")
	for _, part := range parts {
		if part == "" {
			panic("web: 路由不能来连续出现 / ")
		}
		root = root.addNode(part)
	}
	root.handleFunc = handleFunc
}

// getRouter 匹配路由
// method 需要校验吗？method = qjwadrksnghjvkrnf
// pattern需要校验吗？
// pattern 一些简单的可以校验：就是说 /awbudijs/asudfnio/asdhuio
// pattern = /user/login/ 合法的
// pattern = /user//login 非法的
func (r *router) getRouter(method string, pattern string) (*node, map[string]string, bool) {
	// 问题：为什么是一个kv都是string类型的
	params := make(map[string]string)
	if pattern == "" {
		return nil, params, false
	}
	// 获取根节点
	root, ok := r.trees[method]
	if !ok {
		return nil, params, false
	}
	// TODO / 这种路由怎么办
	if pattern == "/" {
		return nil, params, true
	}
	// 切割pattern
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		if part == "" {
			return nil, params, false
		}
		root = root.getNode(part)
		if root == nil {
			return nil, params, false
		}
		// 想一想：我们注册的路由是 /study/:course
		// 					    /study/golang
		// {"course": "golang"}
		// 节点找到了
		// 1. 是静态路由 pass
		// 2. 是动态路由中的参数路由-特殊处理：把参数维护住
		if strings.HasPrefix(root.part, ":") {
			params[root.part[1:]] = part
		}
		// /study/:course/action
		// /study/*filepath
	}
	return root, params, root.handleFunc != nil
}

type node struct {
	part string
	// children 其实就是静态路由
	children map[string]*node
	// handleFunc 这里存的是当前节点上的视图函数
	// 就是咱们之前讲的data
	handleFunc HandleFunc
	// paramChild 参数路由
	// 问题一：为什么这里是一个纯的node节点呢？
	// /study/:course
	// /study/:programming
	// /study/golang
	// 问题二：静态路由和动态路由的优先级问题
	// 大家认为这两个那个优先级高
	// 注册的路由 /study/golang
	// 注册的路由 /study/:course
	// 请求的地址 /study/golang
	// 结论：静态路由的优先级高于动态路由
	paramChild *node
}

// addNode 这个方法是在服务启动前调用
func (n *node) addNode(part string) *node {
	if strings.HasPrefix(part, ":") && n.paramChild == nil {
		n.paramChild = &node{part: part}
		return n.paramChild
	}
	// 判断当前节点有没有children属性，就是说，是不是nil
	if n.children == nil {
		n.children = make(map[string]*node)
	}
	child, ok := n.children[part]
	if !ok {
		child = &node{
			part: part,
		}
		n.children[part] = child
	}
	return child
}

func (n *node) getNode(part string) *node {
	// n 的 children属性都不存在
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild
		}
		return nil
	}
	// 正常思路：从静态路由中找
	child, ok := n.children[part]
	if !ok {
		// 到这里了，就说明没有匹配到静态路由
		if n.paramChild != nil {
			return n.paramChild
		}
		return nil
	}
	return child
}

// 刚才的意思是：一个路由的同一个位置，不能同时有静态路由和动态路由

/**
路由分为动态的和静态的
- 静态路由
	/study/golang
	/user/login
	/register
	...

- 动态路由
	1. 参数路由
		/study/:course 这是咱们注册的路由
			匹配的时候能匹配到什么路由：
				/study/golang、/study/python:能匹配到
				/study/golang/action:匹配不到
	2. 通配符路由:贪婪匹配
		/static/*filepath 这是咱们注册的路由
			匹配的时候能匹配到什么路由：
				/static/css/stylADAHSDCJUVKJSSVDSEKJ FCDNVNe.css
				/static/js/index.js
	3. 正则路由

**/
