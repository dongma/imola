package rpc

import (
	"context"
)

type UserService struct {
	// GetById 用反射来赋值，本质上是一个字段
	// 其类型是函数的字段，它不是方法（它不是定义在UserService上的方法）
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)
}

func (u UserService) Name() string {
	return "user-service"
}

type GetByIdReq struct {
	Id int
}

type GetByIdResp struct {
}
