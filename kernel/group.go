package kernel

// IGroup 代表前缀分组
type IGroup interface {
	// Get 实现HttpMethod方法
	Get(string, ...ControllerHandler)
	Post(string, ...ControllerHandler)
	Put(string, ...ControllerHandler)
	Delete(string, ...ControllerHandler)

	// Group 实现嵌套group, Use用于注册中间件
	Group(string) IGroup
	Use(middlewares ...ControllerHandler)
}

// Group struct 实现了IGroup
type Group struct {
	core   *Core  // 指向core结构
	parent *Group //指向上一个Group，如果有的话
	prefix string // 这个group的通用前缀

	middlewares []ControllerHandler
}

// NewGroup 初始化Group
func NewGroup(core *Core, prefix string) *Group {
	return &Group{
		core:        core,
		parent:      nil,
		prefix:      prefix,
		middlewares: []ControllerHandler{},
	}
}

// Get 实现Get方法
func (g *Group) Get(url string, handlers ...ControllerHandler) {
	url = g.getAbsolutePrefix() + url
	allHandlers := append(g.getMiddlewares(), handlers...)
	g.core.Get(url, allHandlers...)
}

// Post 实现Post方法
func (g *Group) Post(url string, handlers ...ControllerHandler) {
	url = g.getAbsolutePrefix() + url
	allHandlers := append(g.getMiddlewares(), handlers...)
	g.core.Post(url, allHandlers...)
}

// Put 实现Put方法
func (g *Group) Put(uri string, handlers ...ControllerHandler) {
	uri = g.getAbsolutePrefix() + uri
	allHandlers := append(g.getMiddlewares(), handlers...)
	g.core.Put(uri, allHandlers...)
}

// Delete 实现Delete方法
func (g *Group) Delete(uri string, handlers ...ControllerHandler) {
	uri = g.getAbsolutePrefix() + uri
	allHandlers := append(g.getMiddlewares(), handlers...)
	g.core.Delete(uri, allHandlers...)
}

// 获取当前group的绝对路径
func (g *Group) getAbsolutePrefix() string {
	if g.parent == nil {
		return g.prefix
	}
	return g.parent.getAbsolutePrefix() + g.prefix
}

// Group 实现Group方法
func (g *Group) Group(url string) IGroup {
	cgroup := NewGroup(g.core, url)
	cgroup.parent = g
	return cgroup
}

func (g *Group) Use(middleware ...ControllerHandler) {
	g.middlewares = append(g.middlewares, middleware...)
}

// 获取某个group的middleware
// 这里就是获取除了Get/Post/Put/Delete之外设置的middleware
func (g *Group) getMiddlewares() []ControllerHandler {
	if g.parent == nil {
		return g.middlewares
	}
	return append(g.parent.getMiddlewares(), g.middlewares...)
}
