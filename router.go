package imola

import (
	"fmt"
	"strings"
)

// router 用来支持对路由树的操作
type router struct {
	// Beego Gin HTTP method也对应一棵树
	trees map[string]*node
}

type node struct {
	path string
	// 子path到子节点的映射
	children map[string]*node
	// 用户注册的业务逻辑
	handler HandleFunc

	// 通配符*表达的节点，任意匹配
	startChild *node
}

func newRouter() router {
	return router{
		trees: map[string]*node{},
	}
}

// addRoute 添加一些限制，path必须以/开头，不能以/结尾，另外中间也不能有连续的 //
func (r *router) addRoute(method string, path string, handleFunc HandleFunc) {
	if path == "" {
		panic("web: 路径不能为空字符串")
	}

	root, ok := r.trees[method]
	// 说明还没有root节点
	if !ok {
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	if path[0] != '/' {
		panic("web: 路径必须以/为开头")
	}
	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路径不能以/为结尾")
	}

	// 根节点特殊处理一下, "/"
	if path == "/" {
		if root.handler != nil {
			panic("web: 路由冲突，重复注册[/]")
		}
		root.handler = handleFunc
		return
	}

	// 切割这个path，/user/home在Split切分后，会变为["","user", "home"]
	path = path[1:]
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		if seg == "" {
			panic("web: 不能有连续的 /")
		}
		// 递归下去找位置，如果中途有节点不存在，就需要创建出来
		child := root.childOrCreate(seg)
		root = child
	}
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突，重复注册 [%s]", path))
	}
	root.handler = handleFunc
}

func (r *router) findRoute(method string, path string) (*node, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	if path == "/" {
		return root, true
	}
	// 把前置和后置的/都去掉
	path = strings.Trim(path, "/")
	// 按斜杠切割
	segs := strings.Split(path, "/")
	for _, seg := range segs {
		child, found := root.childOf(seg)
		if !found {
			return nil, false
		}
		root = child
	}
	return root, true
}

// childOrCreate 根据seg找子节点，当子节点不存在时，进行创建
func (n *node) childOrCreate(seg string) *node {
	if seg == "*" {
		n.startChild = &node{
			path: seg,
		}
		return n.startChild
	}
	// 1、children为nil，创建子节点
	if n.children == nil {
		n.children = map[string]*node{}
	}
	// 2、子节点中不存在seg节点，则创建新的节点
	res, ok := n.children[seg]
	if !ok {
		res = &node{
			path: seg,
		}
		n.children[seg] = res
	}
	return res
}

func (n *node) childOf(path string) (*node, bool) {
	if n.children == nil {
		return n.startChild, n.startChild != nil
	}
	res, ok := n.children[path]
	if !ok {
		return n.startChild, n.startChild != nil
	}
	return res, ok
}
