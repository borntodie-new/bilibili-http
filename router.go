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
func (r *router) addRouter(method string, pattern string, handleFunc HandleFunc, middlewareChains ...MiddlewareHandleFunc) {
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
		root, ok = root.addNode(part)
		if !ok {
			panic(fmt.Sprintf("web: 路由冲突 - %s", pattern))
		}
	}
	if root.handleFunc != nil {
		panic(fmt.Sprintf("web: 路由冲突 - %s", pattern))
	}
	// 设置视图函数
	root.handleFunc = handleFunc
	// 设置中间件列表
	root.middlewareChains = middlewareChains
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
		return root, params, true
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

		// 参数路由和通配符路由是特殊的静态路由
		// 既然是特殊的路由，那咱们就得特殊处理
		// 参数路由和通配符路由还有点区别，就是通配符路由是贪婪匹配的
		if strings.HasPrefix(root.part, "*") {
			// /assets/*filepath
			// /assets/css/index.css
			index := strings.Index(pattern, part)
			params[root.part[1:]] = pattern[index:]
			// 直接return就表示后面的不在匹配节点了
			return root, params, root.handleFunc != nil
		}
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
	// 单一路由上的中间件列表
	middlewareChains MiddlewareChains

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

	// starChild 通配符路由
	starChild *node
}

// addNode 这个方法是在服务启动前调用
func (n *node) addNode(part string) (*node, bool) {
	if strings.HasPrefix(part, "*") {
		// 这里是通配符路由
		if n.paramChild != nil {
			// 当前节点的参数路由上是有值的，直接判定是冲突路由
			// study/:course
			// study/*filepath
			return nil, false
		}
		n.starChild = &node{part: part}
		return n.starChild, n.paramChild == nil
	}
	if strings.HasPrefix(part, ":") {
		// 这里是参数路由
		if n.starChild != nil {
			// 当前节点的通配符路由上是有值的，直接判定是冲突路由
			// study/*filepath
			// study/:course
			return nil, false
		}

		if n.paramChild == nil {
			// 创建参数路由
			n.paramChild = &node{part: part}
		}
		if n.paramChild.part != part {
			// /study/:course
			// /study/:action
			// 冲突路由，直接返回false
			return nil, false
		}
		// /study/:course
		// /study/:course/action
		return n.paramChild, n.starChild == nil

		// 第一版
		//if n.paramChild != nil {
		//	if n.paramChild.part == part{
		//		// /study/:course: 这时候，:course节点的handleFunc属性上是有数据的
		//		// /study/:action/action：这时候，action节点的handleFunc属性上也是有数据的
		//		return n.paramChild, n.starChild == nil
		//	}
		//	// /study/:course
		//	// /study/:action
		//	return nil, false
		//}
		//// 问题：
		///*
		//	/study/:course: 这时候，:course节点的handleFunc属性上是有数据的
		//	/study/:course/action：这时候，action节点的handleFunc属性上也是有数据的
		//*/
		//n.paramChild = &node{part: part}
		//return n.paramChild, n.starChild == nil
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
	return child, true
}

func (n *node) getNode(part string) *node {
	// n 的 children属性都不存在
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild
		}
		if n.starChild != nil {
			return n.starChild
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
		if n.starChild != nil {
			return n.starChild
		}
		return nil
	}
	return child
}

// 我们目前的添加节点的逻辑存在些问题
// 就是说，我们的添加节点的逻辑处理路由冲突的情况

// 路由冲突有哪些情况
/*
/study/login
/study/login
这是一个冲突的路由

/study/:course
/study/:action
/study/golang进来，到底是匹配那个呢？


同一个位置，参数路由和通配符路由不能同时存在
/study/*filepath
/study/:course
/study/golang进来，到底是匹配那个呢？
*/

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
			filepath = js/index.js

	3. 正则路由

**/

/**
1. 静态路由
2. 动态
	- 参数路由
	- 通配符路由

问题一：这三个路由之间的优先级？
结论：静态路由>参数路由>通配符路由
1. /study/golang
2. /study/:course
3. /study/*action

/study/:course 其实完全可以当成是一个静态路由
/study/*action 也可以认为是一个静态路由
只不过上述两个路由是我们人为地设置为动态路由

问题二：参数路由和通配符路由的优先级？
完全取决于设计者
咱们的设计是：参数路由的优先级高于通配符的优先级

**/

// 添加节点操作
/*
1. 如果一条路由能够成功添加成一个通配符路由，是不是就意味着它也能添加到参数路由。保证没错的话就是
如果能够添加成是参数路由，那一定能够添加成是一个静态路由。所以咱们先从小范围判断
*/

/*
匹配路由
1. /study/login
2. /study/:course

现在进来/study/login路由，是匹配1号还是2号。肯定是1号
现在进来/study/register路由，是匹配1号还是2号。肯定是2号

同理
1. /study/login
2. /study/*filepath

抛出结论：就是说，优先判断是否是静态路由，在判断是否是参数路由。最后判断是否是通配符路由
*/
