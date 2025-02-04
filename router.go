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
	route string

	path string
	// 子path到子节点的映射
	children map[string]*node
	// 用户注册的业务逻辑
	handler HandleFunc

	// 路径参数匹配
	paramChild *node
	// 通配符*表达的节点，任意匹配
	startChild *node
}

type matchInfo struct {
	n          *node
	pathParams map[string]string
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
		root.route = "/"
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
	root.route = path
}

func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	if path == "/" {
		return &matchInfo{
			n: root,
		}, true
	}
	// 把前置和后置的/都去掉
	path = strings.Trim(path, "/")
	// 按斜杠切割
	segs := strings.Split(path, "/")
	var pathParams map[string]string
	for _, seg := range segs {
		child, paramChild, found := root.childOf(seg)
		// 命中了路径参数
		if paramChild {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			// path是:id这种格式
			pathParams[child.path[1:]] = seg
		}
		if !found {
			return nil, false
		}
		root = child
	}
	return &matchInfo{
		n:          root,
		pathParams: pathParams,
	}, true
}

// childOrCreate 根据seg找子节点，当子节点不存在时，进行创建
func (n *node) childOrCreate(seg string) *node {
	if seg[0] == ':' {
		if n.startChild != nil {
			panic("web: 不允许同时注册路径参数和通配符，已有通配符匹配")
		}
		n.paramChild = &node{
			path: seg,
		}
		return n.paramChild
	}
	if seg == "*" {
		if n.paramChild != nil {
			panic("web: 不允许同时注册路径参数和通配符，已有路径参数")
		}
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

// childOf 优先考虑静态匹配，匹配不上考虑通配符匹配
// 第一个返回参数为：是否子节点，第二个返回：是否路径参数、第三个返回：是否通配符节点
func (n *node) childOf(path string) (*node, bool, bool) {
	if n.children == nil {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.startChild, false, n.startChild != nil
	}
	res, ok := n.children[path]
	if !ok {
		if n.paramChild != nil {
			return n.paramChild, true, true
		}
		return n.startChild, false, n.startChild != nil
	}
	return res, false, ok
}
