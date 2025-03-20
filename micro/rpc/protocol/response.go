package protocol

import (
	"encoding/binary"
)

type Response struct {
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

	// 响应错误
	Error []byte
	// 响应内容
	Data []byte
}

func EncodeResp(resp *Response) []byte {
	bs := make([]byte, resp.HeadLength+resp.BodyLength)
	// 1. 写入头部长度
	binary.BigEndian.PutUint32(bs[:4], resp.HeadLength)
	// 2. 写入body长度
	binary.BigEndian.PutUint32(bs[4:8], resp.BodyLength)
	// 3. 写入request ID
	binary.BigEndian.PutUint32(bs[8:12], resp.RequestId)
	// 4. 写入version
	bs[12] = resp.Version
	bs[13] = resp.Compresser
	bs[14] = resp.Serializer
	// 5、写入serviceName，同时写入分隔符
	cur := bs[15:]

	// 7.写入data内容
	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]
	copy(cur, resp.Data)
	return bs
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}
	// 1. 头四个字节是头部长度
	resp.HeadLength = binary.BigEndian.Uint32(data[:4])
	// 2. 紧接着，又是四个字节，对英语body长度
	resp.BodyLength = binary.BigEndian.Uint32(data[4:8])
	// 3. 又是4个字节，对应于 request ID
	resp.RequestId = binary.BigEndian.Uint32(data[8:12])
	// 4.还原version
	resp.Version = data[12]
	resp.Compresser = data[13]
	resp.Serializer = data[14]
	// 写入error
	if resp.HeadLength > 15 {
		resp.Error = data[15:resp.HeadLength]
	}

	// 7.写入data内容
	if resp.BodyLength != 0 {
		resp.Data = data[resp.HeadLength:]
	}
	return resp
}

func (resp *Response) CalculateHeaderLength() {
	resp.HeadLength = 15 + uint32(len(resp.Error))
}

func (resp *Response) CalculateBodyLength() {
	resp.BodyLength = uint32(len(resp.Data))
}
