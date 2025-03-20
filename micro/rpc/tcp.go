package rpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	// 协议头和协议体，总共8个字节，一次先读取完
	lenBs := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}

	headerLength := binary.BigEndian.Uint32(lenBs[:4])
	bodyLength := binary.BigEndian.Uint32(lenBs[4:])
	length := headerLength + bodyLength
	// 我响应有多长
	data := make([]byte, length)
	_, err = conn.Read(data[8:])
	copy(data[:8], lenBs)
	return data, err
}
