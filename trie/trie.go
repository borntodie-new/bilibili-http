package trie

import (
	"errors"
	"strings"
)

type Router struct {
	root map[string]*Node
}

// AddRouter 最开始我们手中有的数据肯定是类似这样的字符串
// /user/login		hello
// /user/login/ =》["", "user", "login", ""]
// /user/register	world
// /yuser//yuasd/asdjo => ["yuser", "", “yuasd”, "asdjo"]
// 就是将/user/login等字符串进行切割，分块保存到前缀树上
func (r *Router) AddRouter(pattern string, data string) {
	// 处理Router的root属性没初始化好的情况
	if r.root == nil {
		r.root = make(map[string]*Node)
	}
	root, ok := r.root["/"]
	//	创建根路由
	if !ok {
		root = &Node{
			part: "/",
		}
		r.root["/"] = root
	}
	// ["user", "login"]
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		if part == "" {
			panic("pattern不符合格式")
		}
		root = root.addNode(part)
	}
	// 循环结束后，此时的root是什么？
	// 这时候，咱们得统一设置data的值
	root.data = data
}

func (r *Router) GetRouter(pattern string) (*Node, error) {
	root, ok := r.root["/"]
	//	创建根路由
	if !ok {
		return nil, errors.New("根节点不存在")
	}
	// 切割pattern
	// ["user", "login"]
	parts := strings.Split(strings.Trim(pattern, "/"), "/")
	for _, part := range parts {
		if part == "" {
			return nil, errors.New("pattern格式不对")
		}
		root = root.getNode(part)
		if root == nil {
			return nil, errors.New("pattern不存在")
		}
	}
	return root, nil
}

type Node struct {
	// part 当前节点的唯一标识
	part string
	// children 维护子节点数据
	// 怎么保存，或者说用什么结构保存
	// 1. map
	// 2. slice
	children map[string]*Node
	// data 当前节点需要保存的数据
	data string
}

// 这个节点有什么功能？
// 1. 注册节点：新建一个Node节点
// 2. 查找节点

// 问题：我们创建节点的时候，是将data直接赋值好还是最后赋值？
//
func (n *Node) addNode(part string) *Node {
	// 判断当前节点有没有children属性，就是说，是不是nil
	if n.children == nil {
		n.children = make(map[string]*Node)
	}
	child, ok := n.children[part]
	if !ok {
		child = &Node{
			part: part,
		}
		n.children[part] = child
	}
	return child
}

func (n *Node) getNode(part string) *Node {
	// n 的 children属性都不存在
	if n.children == nil {
		return nil
	}
	// 正常思路
	child, ok := n.children[part]
	if !ok {
		return nil
	}
	return child
}
