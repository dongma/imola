package rpc

import (
	"context"
	"imola/micro/proto/gen"
	"log"
	"testing"
	"time"
)

type UserService struct {
	// GetById 用反射来赋值，本质上是一个字段
	// 其类型是函数的字段，它不是方法（它不是定义在UserService上的方法）
	GetById func(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error)

	GetByIdProto func(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error)
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
	Err error
	Msg string
}

func (u *UserServiceServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	log.Println(req)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}

func (u *UserServiceServer) GetByIdProto(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	log.Println(req)
	return &gen.GetByIdResp{
		User: &gen.User{
			Name: u.Msg,
		},
	}, u.Err
}

func (u UserServiceServer) Name() string {
	return "user-service"
}

// UserServiceServerTimeout 专门用来测试timeout
type UserServiceServerTimeout struct {
	T     *testing.T
	Sleep time.Duration
	Err   error
	Msg   string
}

func (u *UserServiceServerTimeout) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	if _, ok := ctx.Deadline(); !ok {
		u.T.Fatal("没有设置超时")
	}
	time.Sleep(u.Sleep)
	return &GetByIdResp{
		Msg: u.Msg,
	}, u.Err
}

func (u UserServiceServerTimeout) Name() string {
	return "user-service"
}
