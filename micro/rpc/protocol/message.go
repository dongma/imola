package protocol

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
