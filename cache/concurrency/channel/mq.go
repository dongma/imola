package channel

import (
	"errors"
	"sync"
)

type Broker struct {
	Mutex sync.Mutex
	Chans []chan Msg
}

func (b *Broker) Send(m Msg) error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	for _, ch := range b.Chans {
		// 写法一: 会存在问题，当Chans的容量不足时，会阻塞住
		// ch <- m
		select {
		case ch <- m:
		default:
			return errors.New("消息队列已满")
		}
	}
	return nil
}

func (b *Broker) Subscribe(capacity int) (<-chan Msg, error) {
	res := make(chan Msg, capacity)
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.Chans = append(b.Chans, res)
	return res, nil
}

type Msg struct {
	Content string
}
