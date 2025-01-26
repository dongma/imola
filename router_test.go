package imola

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_addRouter(t *testing.T) {

	// 构造路由树、验证路由树
	var testRoutes []struct {
		method string
		path   string
	}

	testRoutes = []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
		{
			method: http.MethodPost,
			path:   "",
		},
	}

	var mockHandler HandleFunc = func(ctx Context) {}
	r := newRouter()
	for _, route := range testRoutes {
		r.addRoute(route.method, route.path, mockHandler)
	}

	// 断言路由树和预期的一摸一样
	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"user": &node{
						path:    "user",
						handler: mockHandler,
						children: map[string]*node{
							"home": &node{
								path:    "home",
								handler: mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
				},
			},
			http.MethodPost: &node{
				path: "/",
				children: map[string]*node{
					"order": &node{
						path: "order",
						children: map[string]*node{
							"create": &node{
								path:    "create",
								handler: mockHandler,
							},
						},
					},
					"login": &node{
						path:    "login",
						handler: mockHandler,
					},
				},
			},
		},
	}

	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)

	r = newRouter()
	assert.Panics(t, func() {
		r.addRoute(http.MethodGet, "", mockHandler)
	})
}

// equal 返回string错误信息，帮助排查问题
func (r *router) equal(y *router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("找不到对应的http method"), false
		}
		// v和dst要相等
		msg, equal := v.equal(dst)
		if !equal {
			return msg, false
		}
	}
	return "", true
}

func (n *node) equal(y *node) (string, bool) {
	if n.path != y.path {
		return fmt.Sprintf("节点路径不匹配"), false
	}
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("子节点数量不相等"), false
	}
	// 比较handler
	nhandler := reflect.ValueOf(n.handler)
	yhandler := reflect.ValueOf(y.handler)
	if nhandler != yhandler {
		return fmt.Sprintf("handler不相等"), false
	}

	for path, c := range n.children {
		dst, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("子节点 %s 不存在", path), false
		}
		msg, ok := c.equal(dst)
		if !ok {
			return msg, false
		}
	}
	return "", true
}

func TestRouter_findRoute(t *testing.T) {
	testRoute := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
	}

	r := newRouter()
	var mockHandler HandleFunc = func(ctx Context) {}
	for _, route := range testRoute {
		r.addRoute(route.method, route.path, mockHandler)
	}

	testCase := []struct {
		method    string
		path      string
		name      string
		wantFound bool
		wantNode  *node
	}{
		{
			// 方法不存在
			name:   "method not found",
			method: http.MethodOptions,
			path:   "/order/detail",
		},
		{
			// 完全命中
			name:      "order create,创建 订单",
			method:    http.MethodPost,
			path:      "/order/create",
			wantFound: true,
			wantNode: &node{
				handler: mockHandler,
				path:    "create",
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.wantFound, found)
			if !found {
				return
			}
			assert.Equal(t, tc.wantNode.path, n.path)
			assert.Equal(t, tc.wantNode.children, n.children)
			// 比较handler
			nhandler := reflect.ValueOf(n.handler)
			yhandler := reflect.ValueOf(tc.wantNode.handler)
			assert.True(t, nhandler == yhandler)
		})
	}
}
