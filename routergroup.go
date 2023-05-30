package bilibili_http

import (
	"fmt"
	"net/http"
	"strings"
)

type RouterGroup struct {
	// prefix 当前路由组的唯一标识
	// 并且必须加上父级的prefix
	prefix string
	// parent 父路由组
	parent *RouterGroup
	// engine 维护一个全局的引擎
	// 这里最好是将engine申明成server接口类型
	engine *HTTPServer
	// ...
	// middlewares 当前路由组所有的中间件
	middlewares []MiddlewareHandleFunc
}

// Group 注册路由组
func (r *RouterGroup) Group(prefix string) *RouterGroup {
	// 保险起见，对prefix做一些校验工作
	// Group("/v1")
	// Group("/v2")
	// Group("/v3/order") 一般来说很少这样做一次性套多层
	// v4 := Group("/v4")
	// order := v4.Group("/order")

	// 保不齐用户瞎搞
	// Group("v2")
	// Group("v2/")
	// Group("/v2/")
	prefix = fmt.Sprintf("/%s", strings.Trim(prefix, "/"))
	rg := &RouterGroup{
		prefix: fmt.Sprintf("%s%s", r.prefix, prefix),
		engine: r.engine,
		parent: r,
	}
	// 将新建的路由组添加到HTTPServer中
	r.engine.groups = append(r.engine.groups, rg)
	return rg
}

// Use 注册中间件
func (r *RouterGroup) Use(mids ...MiddlewareHandleFunc) {
	// 问题：中间件放哪？维护在哪里？
	r.middlewares = append(r.middlewares, mids...)
}

// 抽取出来的公共方法

// GET GET请求
func (r *RouterGroup) GET(pattern string, handleFunc HandleFunc, middlewareChains ...MiddlewareHandleFunc) {
	r.addRouter(http.MethodGet, pattern, handleFunc, middlewareChains...)
}

// POST GET请求
func (r *RouterGroup) POST(pattern string, handleFunc HandleFunc, middlewareChains ...MiddlewareHandleFunc) {
	r.addRouter(http.MethodPost, pattern, handleFunc, middlewareChains...)
}

// DELETE GET请求
func (r *RouterGroup) DELETE(pattern string, handleFunc HandleFunc, middlewareChains ...MiddlewareHandleFunc) {
	r.addRouter(http.MethodDelete, pattern, handleFunc, middlewareChains...)
}

// PUT GET请求
func (r *RouterGroup) PUT(pattern string, handleFunc HandleFunc, middlewareChains ...MiddlewareHandleFunc) {
	r.addRouter(http.MethodPut, pattern, handleFunc, middlewareChains...)
}

// addRouter1 这里是注册路由的唯一路径
// 这里是和router路由树直接交互的入口，所以必须调用router的addRouter方法
func (r *RouterGroup) addRouter(method string, pattern string, handleFunc HandleFunc, middlewareChains ...MiddlewareHandleFunc) {
	// 这里就是将路由组的唯一标识和需要注册的路由进行绑定
	pattern = fmt.Sprintf("%s%s", r.prefix, pattern)
	r.engine.router.addRouter(method, pattern, handleFunc, middlewareChains...)
}

func newRouterGroup() *RouterGroup {
	return &RouterGroup{}
}

/*
问题来了，这个routerGroup里面要有什么字段，或者说什么属性？
怎么办？
抄一个

Gin版本的路由组
type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	engine   *Engine
	root     bool
}

兔兔的版本
type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // support middleware
	parent      *RouterGroup  // support nesting
	engine      *Engine       // all groups share a Engine instance
}

*/

/*
问题又来了，为什么要engine，或者是咱们框架中的HTTPServer，为什么要它？


*/

/*
思考，这个路由组需要提供哪些功能？

e := gin.Default()
e.GET("/login", func(ctx *gin.Context){})
v1 := e.Group("/v1")
{
	v1.GET("/user/:id", func(ctx *gin.Context){})
	v1.DELETE("/user/:id", func(ctx *gin.Context){})
	v1.POST("/user", func(ctx *gin.Context){})
}
v2 := e.Group("/v1")
{
	v2.GET("/user/:id", func(ctx *gin.Context){})
	v2.DELETE("/user/:id", func(ctx *gin.Context){})
	v2.POST("/user", func(ctx *gin.Context){})
}
e.Run(":8080")
这是咱们正常的思路

根据上述使用，咱们就总结出几个点
1. Group方法返回的应该是一个RouterGroup对象
2. Engine和RouterGroup在某些方法上是一样的，可以认为是”一个东西“
可以认为是一个东西，是不是可以想到，将Engine和RouterGroup嵌套
这样嵌套的话，是不是就能将二者联系起来
让二者嵌套起来，有什么好处？
1. 二者存在父子关系：这个就可以理解成其他面向对象的编程语言中的继承。是不是子可以使用父的方法

引出一个新问题：
到底是Engine套RouterGroup还是RouterGroup套Engine?
其实这里是相互嵌套的

引入一个新问题：
Engine和RouterGroup在某些方法上是一样的，那那些方法是一样的？这些方法需要放在谁身上？
一些共用的方法，其实是放在父对象中的
那些方法？是不是那些GET、POST...


可以问一下大家：就是目前咱们的Engine和RouterGroup谁是父，谁是子？
直接抛出结论，Engine是子，RouterGroup是父


class XXX(object):
	pass

class YYY(XXX):
	pass



需求：
1. 需要有一个注册路由组的方法
*/

/*
目前来说，将一些公共的方法抽取出来之后，直接报错，报错的问题是：没有addRouter方法
addRouter在哪？什么作用？

分析出来：addRouter是在router身上的
router又在HTTPServer身上
两种思路
1. 在RouterGroup身上维护一个router对象
2. 在RouterGroup身上维护一个HTTPServer对象

思路一分析
先抛出结论，这样不好，为什么？
路由树是全局的，我们当初把路由树放在HTTPServer身上，也就是看中HTTPServer的全局唯一的
路由树也是全局唯一的对象，所以我们自然而然就能想到把router维护在HTTPServer
那路由组可不是全局唯一的

思路二分析
先抛出结论，这样做好
虽然说HTTPServer和router都是全局唯一的，
router身上的功能太少了，而且还不常用
HTTPServer身上的功能就很多，而且还常用
*/

/*
目前所作的所有努力，还不够？
注册路由组没问题了，那怎么让它生效？
e := gin.Default()
e.GET("/login", func(ctx *gin.Context){})
v1 := e.Group("/v1")
{
	v1.GET("/user/:id", func(ctx *gin.Context){})
	v1.DELETE("/user/:id", func(ctx *gin.Context){})
	v1.POST("/user", func(ctx *gin.Context){})
}
v2 := e.Group("/v1")
{
	v2.GET("/user/:id", func(ctx *gin.Context){})
	v2.DELETE("/user/:id", func(ctx *gin.Context){})
	v2.POST("/user", func(ctx *gin.Context){})
}

e.Run(":8080")

GET - /login
GET - /v1/user/:id
DELETE - /v1/user/:id
POST - /v1/user/:id
GET - /v2/user/:id
DELETE - /v2/user/:id
POST - /v2/user/:id

整个框架还没有和路由组搭上关系

目前，
1. RouterGroup有一个addRouter方法，router也有一个addRouter方法
2. HTTPServer还需要和router做交互吗？【注册、匹配】
现在谁和路由树做交互？是RouterGroup吧！
*/

/*
建议大家以后写结构体的时候，最好搭配一个构造方法
为什么？
扩展性

*/

/*
e := gin.Default()
e.GET("/login", func(ctx *gin.Context){})
v1 := e.Group("/v1")
v1.Use(func(ctx *gin.Context){})
{
	v1.GET("/user/:id",func(ctx *gin.Context){}, func(ctx *gin.Context){})
	v1.DELETE("/user/:id", func(ctx *gin.Context){})
	v1.POST("/user", func(ctx *gin.Context){})
}
v2 := e.Group("/v1")
{
	v2.GET("/user/:id", func(ctx *gin.Context){})
	v2.DELETE("/user/:id", func(ctx *gin.Context){})
	v2.POST("/user", func(ctx *gin.Context){})
}

e.Run(":8080")

想一想：我们注册某个或者多个中间件到具体的某个视图上，那这些注册的中间件需要维护在哪？
天然就能想到应该是维护在当前这个路由的节点上。
*/

//func Inter(number int, values ...interface{})  {
//	// 这里需要的是一个打散或的数据
//	// 执行逻辑的时候，就会报错
//}
//
//func Func(values ...string)  {
//	Inter(1, values)
//}
//
//func Test()  {
//	Func("as")
//}
/*
不定长参数：不定长参数，他只能在最后作为参数
*/
