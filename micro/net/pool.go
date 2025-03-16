package net

import (
	"net"
	"time"
)

type Pool struct {
	// 空闲连接队列
	idlesConns chan *idleConn
	// 请求队列
	reqQueue []connReq
	// 最大连接数
	maxCnt int
	// 当前连接数
	cnt int
	// 最大空闲时间
	maxIdleTime time.Duration
}

type idleConn struct {
	c net.Conn
}

type connReq struct{}
