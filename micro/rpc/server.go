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
		req := protocol.DecodeReq(reqBs)
		resp, err := s.Invoke(context.Background(), req)
		if err != nil {
			// 处理业务 error
			resp.Error = []byte(err.Error())
		}
		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()

		_, err = conn.Write(protocol.EncodeResp(resp))
		if err != nil {
			return err
		}
	}
}

func (s *Server) Invoke(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	// 还原了调用信息，已经知道了service name, method name和参数了，便可以发起业务调用
	service, ok := s.services[req.ServiceName]
	resp := &protocol.Response{
		RequestId:  req.RequestId,
		Version:    req.Version,
		Compresser: req.Compresser,
		Serializer: req.Serializer,
	}
	if !ok {
		return resp, errors.New("你要调用的服务不存在")
	}
	respData, err := service.Invoke(ctx, req.MethodName, req.Data)
	resp.Data = respData
	if err != nil {
		return resp, err
	}
	return resp, nil
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
		err = results[1].Interface().(error)
	}

	var res []byte
	if results[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = json.Marshal(results[0].Interface())
		if er != nil {
			return nil, er
		}
	}
	return res, err
}
