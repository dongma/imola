package queue

import (
	"context"
	"sync"
)

type ConcurrentBlockingQueue[T any] struct {
	mutex    *sync.Mutex
	data     []T
	notfull  *sync.Cond
	notempty *sync.Cond
}

// NewConcurrentBlockingQueue 创建并发队列
func NewConcurrentBlockingQueue[T any](maxSize int) *ConcurrentBlockingQueue[T] {
	return &ConcurrentBlockingQueue[T]{
		data: make([]T, 0, maxSize),
	}
}

func (c *ConcurrentBlockingQueue[T]) Enqueue(ctx context.Context, elem T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for c.IsFull() {
		// 我阻塞我自己，直到有人唤醒我
		c.notfull.Wait()
	}
	c.data = append(c.data, elem)
	return nil
}

func (c *ConcurrentBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var elem T
		return elem, ctx.Err()
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for c.IsEmpty() {
		c.notempty.Wait()
	}
	elem := c.data[0]
	c.data = c.data[1:]
	return elem, nil
}

func (c *ConcurrentBlockingQueue[T]) IsEmpty() bool {
	return len(c.data) == 0
}

func (c *ConcurrentBlockingQueue[T]) Len() uint64 {
	return uint64(len(c.data))
}

func (c *ConcurrentBlockingQueue[T]) IsFull() bool {
	panic("implement me")
}
