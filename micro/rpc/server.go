package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"imola/micro/rpc/protocol"
	"net"
	"reflect"
)

type Server struct {
	services map[string]ReflectionStub
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]ReflectionStub, 16),
	}
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = ReflectionStub{
		s:     service,
		value: reflect.ValueOf(service),
	}
}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		// 比较常见的就是端口被占用
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if er := s.handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

// handleConn 我们可以认为，一个请求包含两部分:
// 1、长度字段: 用八个字节表示； 2、请求数据，响应也是这个规范
func (s *Server) handleConn(conn net.Conn) error {
	for {
		// lenBs是长度字段的字节表示
		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		// 还原调用信息
		req := &protocol.Request{}
		err = json.Unmarshal(reqBs, req)
		if err != nil {
			return err
		}
		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			// TODO: 这个可能是你的业务error,暂时不知道怎么回传，所以简单记录一下
			return err
		}
		res := EncodeMsg(resp.Data)
		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	// 还原了调用信息，已经知道了service name, method name和参数了，便可以发起业务调用
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("你要调用的服务不存在")
	}
	resp, err := service.Invoke(ctx, req.MethodName, req.Data)
	if err != nil {
		return nil, err
	}
	return &protocol.Response{
		Data: resp,
	}, err
}

type ReflectionStub struct {
	s     Service
	value reflect.Value
}

func (r *ReflectionStub) Invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {
	// 反射找到方法，并执行调用
	method := r.value.MethodByName(methodName)
	in := make([]reflect.Value, 2)

	// 暂时我们不知道如何传这个context,所以我们就直接写死
	in[0] = reflect.ValueOf(context.Background())
	inReq := reflect.New(method.Type().In(1).Elem())
	err := json.Unmarshal(data, inReq.Interface())
	if err != nil {
		return nil, err
	}

	in[1] = inReq
	results := method.Call(in)
	// results[0]是返回值，results[1]是error
	if results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}
	return json.Marshal(results[0].Interface())
}
