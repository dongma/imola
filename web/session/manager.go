package session

import (
	"github.com/dongma/imola/web"
)

// Manager 用户操作的友好封装
type Manager struct {
	Store
	Propagator
	SessionCtxKey string
}

// GetSession 从ctx中拿到Session，同时在ctx的UserValues中进行缓存
func (m *Manager) GetSession(ctx *web.Context) (Session, error) {
	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]any, 8)
	}
	val, ok := ctx.UserValues[m.SessionCtxKey]
	if ok {
		return val.(Session), nil
	}
	// 从请求上下文中ctx中获取session id，取到session后在ctx中缓存
	sId, err := m.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}
	session, err := m.Get(ctx.Req.Context(), sId)
	if err != nil {
		return nil, err
	}
	ctx.UserValues[m.SessionCtxKey] = session
	return session, nil
}

// InitSession 初始化一个session，并且注入到http session中
func (m *Manager) InitSession(ctx *web.Context, id string) (Session, error) {
	session, err := m.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}
	if err = m.Inject(id, ctx.Resp); err != nil {
		return nil, err
	}
	return session, nil
}

// RefreshSession 刷新session
func (m *Manager) RefreshSession(ctx *web.Context) (Session, error) {
	session, err := m.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	// 刷新存储的过期时间
	err = m.Refresh(ctx.Req.Context(), session.ID())
	if err != nil {
		return nil, err
	}
	// 重新注入到http里面
	if err = m.Inject(session.ID(), ctx.Resp); err != nil {
		return nil, err
	}
	return session, nil
}

// RemoveSession 删除session
func (m *Manager) RemoveSession(ctx *web.Context) error {
	session, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	err = m.Store.Remove(ctx.Req.Context(), session.ID())
	if err != nil {
		return err
	}
	return m.Propagator.Remove(ctx.Resp)
}
