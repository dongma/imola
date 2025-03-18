package rpc

import (
	"context"
	"log"
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
	Msg string
}

// UserServiceServer 具体的service实现，同时也需实现GetById方法
type UserServiceServer struct {
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	log.Println(req)
	return &GetByIdResp{
		Msg: "hello, world",
	}, nil
}

func (u UserServiceServer) Name() string {
	return "user-service"
}
