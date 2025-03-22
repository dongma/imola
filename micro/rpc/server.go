package rpc

import (
	"context"
	"errors"
	"imola/micro/rpc/compresser"
	"imola/micro/rpc/protocol"
	"imola/micro/rpc/serialize"
	"imola/micro/rpc/serialize/json"
	"net"
	"reflect"
	"strconv"
	"time"
)

type Server struct {
	services    map[string]ReflectionStub
	serializers map[uint8]serialize.Serializer
	compressers map[uint8]compresser.Compresser
}

func NewServer() *Server {
	res := &Server{
		services:    make(map[string]ReflectionStub, 16),
		serializers: make(map[uint8]serialize.Serializer, 16),
		compressers: make(map[uint8]compresser.Compresser, 4),
	}
	res.RegisterSerializer(&json.Serializer{})
	res.RegisterCompresser(compresser.DoNothing{})
	return res
}

func (s *Server) RegisterSerializer(sl serialize.Serializer) {
	s.serializers[sl.Code()] = sl
}

func (s *Server) RegisterCompresser(c compresser.Compresser) {
	s.compressers[c.Code()] = c
}

func (s *Server) RegisterService(service Service) {
	s.services[service.Name()] = ReflectionStub{
		s:           service,
		value:       reflect.ValueOf(service),
		serializers: s.serializers,
		compressers: s.compressers,
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
		ctx := context.Background()

		cancel := func() {}
		if deadlineStr, ok := req.Meta["deadline"]; ok {
			if deadline, er := strconv.ParseInt(deadlineStr, 10, 64); er == nil {
				ctx, cancel = context.WithDeadline(ctx, time.UnixMilli(deadline))
			}
		}
		oneway, ok := req.Meta["one-way"]
		if ok && oneway == "true" {
			ctx = CtxWithOneway(ctx)
		}

		resp, err := s.Invoke(ctx, req)
		cancel()
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

	if isOneway(ctx) {
		go func() {
			_, _ = service.Invoke(ctx, req)
		}()
		return resp, errors.New("micro: 微服务服务端 oneway 请求")
	}
	respData, err := service.Invoke(ctx, req)
	resp.Data = respData
	if err != nil {
		return resp, err
	}
	return resp, nil
}

type ReflectionStub struct {
	s           Service
	value       reflect.Value
	serializers map[uint8]serialize.Serializer
	compressers map[uint8]compresser.Compresser
}

func (r *ReflectionStub) Invoke(ctx context.Context, req *protocol.Request) ([]byte, error) {
	// 反射找到方法，并执行调用
	method := r.value.MethodByName(req.MethodName)
	in := make([]reflect.Value, 2)

	// 暂时我们不知道如何传这个context,所以我们就直接写死
	in[0] = reflect.ValueOf(ctx)
	inReq := reflect.New(method.Type().In(1).Elem())

	serializer, ok := r.serializers[req.Serializer]
	if !ok {
		return nil, errors.New("micro: 不支持序列化协议")
	}
	c, ok := r.compressers[req.Compresser]
	if !ok {
		return nil, errors.New("micro: 不支持压缩算法")
	}
	data, err := c.UnCompress(req.Data)
	if err != nil {
		return nil, err
	}
	err = serializer.Decode(data, inReq.Interface())
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
		res, er = serializer.Encode(results[0].Interface())
		if er != nil {
			return nil, er
		}
	}
	return res, err
}
