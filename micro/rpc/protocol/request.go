package protocol

import "encoding/binary"

type Request struct {
	// 头部，消息长度
	HeadLength uint32
	// 消息体长度
	BodyLength uint32
	// 请求Id
	RequestId uint32
	// 协议版本
	Version uint8
	// 压缩算法
	Compresser uint8
	// 序列化协议
	Serializer uint8
	// 服务名和方法名称
	ServiceName string
	MethodName  string
	// 扩展字段
	Meta map[string]string
	// 请求体内容
	Data []byte
}

func EncodeReq(req *Request) []byte {
	bs := make([]byte, req.HeadLength+req.BodyLength)
	// 1. 写入头部长度
	binary.BigEndian.PutUint32(bs[:4], req.HeadLength)
	// 2. 写入body长度
	binary.BigEndian.PutUint32(bs[4:8], req.BodyLength)
	return bs
}

func DecodeReq(data []byte) *Request {
	req := &Request{}
	// 1. 头四个字节是头部长度
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	// 2. 紧接着，又是四个字节，对英语body长度
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	return req
}
