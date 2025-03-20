package protocol

import (
	"bytes"
	"encoding/binary"
)

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
	// 3. 写入request ID
	binary.BigEndian.PutUint32(bs[8:12], req.RequestId)
	// 4. 写入version
	bs[12] = req.Version
	bs[13] = req.Compresser
	bs[14] = req.Serializer
	// 5、写入serviceName，同时写入分隔符
	cur := bs[15:]
	copy(cur, req.ServiceName)
	cur = cur[len(req.ServiceName):]
	cur[0] = '\n'
	cur = cur[1:]
	copy(cur, req.MethodName)
	// 6、写入meta
	cur = cur[len(req.MethodName):]
	cur[0] = '\n'
	cur = cur[1:]

	for key, value := range req.Meta {
		copy(cur, key)
		cur = cur[len(key):]
		cur[0] = '\r'
		cur = cur[1:]
		copy(cur, value)
		cur = cur[len(value):]
		// 再加一个\n
		cur[0] = '\n'
		cur = cur[1:]
	}

	// 7.写入data内容
	copy(cur, req.Data)
	return bs
}

func DecodeReq(data []byte) *Request {
	req := &Request{}
	// 1. 头四个字节是头部长度
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	// 2. 紧接着，又是四个字节，对英语body长度
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	// 3. 又是4个字节，对应于 request ID
	req.RequestId = binary.BigEndian.Uint32(data[8:12])
	// 4.还原version
	req.Version = data[12]
	req.Compresser = data[13]
	req.Serializer = data[14]
	// 5.反序列化serviceName，要引入分隔符，切分service name和method name
	header := data[15:]
	// 近似于，user-service、GetById
	index := bytes.IndexByte(header, '\n')
	req.ServiceName = string(header[:index])
	header = header[index+1:]

	// 切出来methodName
	index = bytes.IndexByte(header, '\n')
	req.MethodName = string(header[:index])
	header = header[index+1:]

	index = bytes.IndexByte(header, '\n')
	// 5.解析meta的部分
	if index != -1 {
		meta := make(map[string]string, 4)
		for index != -1 {
			pair := header[:index]
			// \r的位置
			pairIndex := bytes.IndexByte(pair, '\r')
			key := string(pair[:pairIndex])
			value := string(pair[pairIndex+1:])
			meta[key] = value
			header = header[index+1:]
			index = bytes.IndexByte(header, '\n')
		}
		req.Meta = meta
	}
	// 7.写入data内容
	if req.BodyLength != 0 {
		req.Data = data[req.HeadLength:]
	}
	return req
}

func (req *Request) CalculateHeaderLength() {
	headLength := 15 + len(req.ServiceName) + 1 + len(req.MethodName) + 1
	for key, value := range req.Meta {
		headLength += len(key)
		// key和value之间的分隔符
		headLength++
		headLength += len(value)
		headLength++
		// 和下一个key value的分隔符
	}
	req.HeadLength = uint32(headLength)
}

func (req *Request) CalculateBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}
