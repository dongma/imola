//go:build e2e

package test

import (
	"fmt"
	"imola/cache/concurrency/channel"
	"sync"
	"testing"
	"time"
)

func TestBroker_Send(t *testing.T) {
	broker := &channel.Broker{}

	// 模拟发送者
	go func() {
		for {
			err := broker.Send(channel.Msg{Content: time.Now().String()})
			if err != nil {
				t.Log(err)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(3)
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("消费者%d", i)
		go func() {
			defer wg.Done()
			msgs, err := broker.Subscribe(100)
			if err != nil {
				t.Log(err)
				return
			}
			for msg := range msgs {
				fmt.Println(name, msg.Content)
			}
		}()
	}
	wg.Wait()
}
