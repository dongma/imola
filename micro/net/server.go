package net

import (
	"encoding/binary"
	"net"
)

// 长度字段使用的字节数量
const numOfLengthBytes = 8

type Server struct{}

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
		lenBs := make([]byte, numOfLengthBytes)
		_, err := conn.Read(lenBs)
		if err != nil {
			return err
		}

		// 我消息有多长？
		length := binary.BigEndian.Uint64(lenBs)
		reqBs := make([]byte, length)
		_, err = conn.Read(reqBs)
		if err != nil {
			return err
		}

		respData := handleMsg(reqBs)
		respLen := len(respData)

		// 在这里构建响应数据，data = respLen的64位表示 + respData
		res := make([]byte, respLen+numOfLengthBytes)
		// 第一步，把长度写进去前八个字节
		binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(respLen))
		// 第二步，写入数据
		copy(res[numOfLengthBytes:], respData)

		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}

func handleMsg(req []byte) []byte {
	res := make([]byte, 2*len(req))
	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	return res
}
