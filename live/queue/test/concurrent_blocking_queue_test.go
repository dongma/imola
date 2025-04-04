package test

import (
	"context"
	"imola/live/queue"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestConcurrentBlockingQueue(t *testing.T) {
	// 只能确保没有死锁
	q := queue.NewConcurrentBlockingQueue[int](10000)

	// 并发的问题都落在m上
	var wg sync.WaitGroup
	wg.Add(30)
	for i := 0; i < 20; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				// 你没有办法校验这里面的中间结果
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				val := rand.Int()
				_ = q.Enqueue(ctx, val)
				cancel()
			}
			wg.Done()
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_, _ = q.Dequeue(ctx)
				// 怎么断言 error
				cancel()
			}
			wg.Done()
		}()
	}
	// 怎么校验 q 对还是不对
	wg.Wait()
}
