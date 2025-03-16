package net

import (
	"encoding/binary"
	"net"
	"time"
)

type Client struct {
	Network string
	Addr    string
}

func (c *Client) Send(data string) (string, error) {
	conn, err := net.DialTimeout(c.Network, c.Addr, time.Second*3)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	reqLen := len(data)
	// 构造请求数据，data = reqLen的64位表示 + respData
	req := make([]byte, reqLen+numOfLengthBytes)
	// 第一步，把长度写进去前八个字节
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(len(data)))
	// 第二步，写入数据
	copy(req[numOfLengthBytes:], data)

	_, err = conn.Write(req)
	if err != nil {
		return "", err
	}

	lenBs := make([]byte, numOfLengthBytes)
	_, err = conn.Read(lenBs)
	if err != nil {
		return "", err
	}

	// 我响应有多长
	length := binary.BigEndian.Uint64(lenBs)
	respBs := make([]byte, length)
	_, err = conn.Read(respBs)
	if err != nil {
		return "", err
	}
	return string(respBs), nil
}
