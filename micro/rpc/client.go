package rpc

import (
	"context"
	"errors"
	"github.com/dongma/imola/micro/rpc/compresser"
	"github.com/dongma/imola/micro/rpc/protocol"
	"github.com/dongma/imola/micro/rpc/serialize"
	"github.com/dongma/imola/micro/rpc/serialize/json"
	"github.com/silenceper/pool"
	"net"
	"reflect"
	"strconv"
	"time"
)

// InitService 要为GetById之类的函数类型的字段赋值
func (c *Client) InitService(service Service) error {
	// 在这里初始化一个proxy
	return SetFuncField(service, c, c.serializer, c.compresser)
}

func SetFuncField(service Service, proxy Proxy, s serialize.Serializer,
	c compresser.Compresser) error {
	if service == nil {
		return errors.New("rpc: 不支持nil")
	}
	val := reflect.ValueOf(service)
	typ := val.Type()
	// 只支持指向结构体的一级指针
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return errors.New("rpc: 只支持指向结构体的一级指针")
	}

	val = val.Elem()
	typ = typ.Elem()
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)

		if fieldVal.CanSet() {
			// 此处才是真正将本地调用捕捉到的地方
			fn := func(args []reflect.Value) (results []reflect.Value) {
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())
				// args[0]是context
				ctx := args[0].Interface().(context.Context)
				// args[1]是req
				reqData, err := s.Encode(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				reqData, err = c.Compress(reqData)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}
				meta := make(map[string]string, 2)
				// 确实设置了超时
				if deadline, ok := ctx.Deadline(); ok {
					meta["deadline"] = strconv.FormatInt(deadline.Unix(), 10)
				}
				if isOneway(ctx) {
					meta["one-way"] = "true"
				}

				req := &protocol.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Data:        reqData,
					Serializer:  s.Code(),
					Meta:        meta,
					Compresser:  c.Code(),
				}

				req.CalculateHeaderLength()
				req.CalculateBodyLength()
				// 要真的发起调用了
				resp, err := proxy.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				var retErr error
				if len(resp.Error) > 0 {
					// 服务端传回来的 error
					retErr = errors.New(string(resp.Error))
				}

				if len(resp.Data) > 0 {
					err = s.Decode(resp.Data, retVal.Interface())
					if err != nil {
						// 反序列化的error
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
				}

				var retErrVal reflect.Value
				if retErr == nil {
					retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
				} else {
					retErrVal = reflect.ValueOf(retErr)
				}
				return []reflect.Value{retVal, retErrVal}
			}
			// 我要设置值给 GetById
			fnVal := reflect.MakeFunc(fieldTyp.Type, fn)
			fieldVal.Set(fnVal)
		}
	}

	return nil
}

const numOfLengthBytes = 8

type Client struct {
	pool       pool.Pool
	serializer serialize.Serializer
	compresser compresser.Compresser
}

type ClientOption func(client *Client)

func ClientWithCompresser(c compresser.Compresser) ClientOption {
	return func(client *Client) {
		client.compresser = c
	}
}

func ClientWithSerializer(sl serialize.Serializer) ClientOption {
	return func(client *Client) {
		client.serializer = sl
	}
}

func NewClient(addr string, opts ...ClientOption) (*Client, error) {
	p, err := pool.NewChannelPool(&pool.Config{
		InitialCap:  1,
		MaxCap:      30,
		MaxIdle:     10,
		IdleTimeout: time.Minute,
		Factory: func() (interface{}, error) {
			return net.DialTimeout("tcp", addr, time.Second*3)
		},
		Close: func(v interface{}) error {
			return v.(net.Conn).Close()
		},
	})
	if err != nil {
		return nil, err
	}
	res := &Client{
		pool:       p,
		serializer: &json.Serializer{},
		compresser: compresser.DoNothing{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func (c *Client) Invoke(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	ch := make(chan struct{})
	defer func() {
		close(ch)
	}()
	var (
		resp *protocol.Response
		err  error
	)
	go func() {
		resp, err = c.doInvoke(ctx, req)
		ch <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		return resp, err
	}
}

func (c *Client) doInvoke(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	data := protocol.EncodeReq(req)
	// 正儿八经的把请求发到服务器上
	resp, err := c.send(ctx, data)
	if err != nil {
		return nil, err
	}
	return protocol.DecodeResp(resp), nil
}

func (c *Client) send(ctx context.Context, data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}
	conn := val.(net.Conn)
	defer func() {
		c.pool.Put(val)
	}()
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}
	if isOneway(ctx) {
		return nil, errors.New("micro: 这是一个oneway调用，你不应该处理任何结果")
	}
	return ReadMsg(conn)
}
