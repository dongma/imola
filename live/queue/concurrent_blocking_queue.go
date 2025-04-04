package queue

import (
	"context"
	"sync"
)

type ConcurrentBlockingQueue[T any] struct {
	mutex   *sync.Mutex
	data    []T
	maxSize int

	notEmptyCond *Cond
	notFullCond  *Cond

	count int
	head  int
	tail  int
	zero  T
}

// NewConcurrentBlockingQueue 创建并发队列
func NewConcurrentBlockingQueue[T any](maxSize int) *ConcurrentBlockingQueue[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueue[T]{
		// 即使是ring buffer, 一次性分配完内存，也是有缺陷的；如果不想一开始就把所有内存都分配好，
		// 可以使用链表
		data:         make([]T, maxSize),
		mutex:        m,
		maxSize:      maxSize,
		notEmptyCond: NewCond(m),
		notFullCond:  NewCond(m),
	}
}

func (c *ConcurrentBlockingQueue[T]) Enqueue(ctx context.Context, elem T) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.mutex.Lock()
	for c.isFull() {
		err := c.notFullCond.WaitWithTimeout(ctx)
		if err != nil {
			return err
		}
	}

	c.data[c.tail] = elem
	c.tail++
	if c.tail == c.maxSize {
		c.tail = 0
	}

	c.notEmptyCond.Broadcast()
	c.mutex.Unlock()
	// 没有人等notEmpty的信号，这一句就会阻塞住
	return nil
}

func (c *ConcurrentBlockingQueue[T]) Dequeue(ctx context.Context) (T, error) {
	if ctx.Err() != nil {
		var elem T
		return elem, ctx.Err()
	}
	c.mutex.Lock()
	for c.isEmpty() {
		err := c.notEmptyCond.WaitWithTimeout(ctx)
		if err != nil {
			var t T
			return t, err
		}
	}

	// 这里要不要考虑缩容
	t := c.data[c.head]
	c.data[c.head] = c.zero
	c.head++
	c.count--
	if c.head == c.maxSize {
		c.head = 0
	}
	c.notFullCond.Broadcast()
	c.mutex.Unlock()
	// 没有人等notFull的信号，这一句就会阻塞住
	return t, nil
}

func (c *ConcurrentBlockingQueue[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isEmpty()
}

func (c *ConcurrentBlockingQueue[T]) isEmpty() bool {
	return c.count == 0
}

func (c *ConcurrentBlockingQueue[T]) Len() uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return uint64(c.count)
}

func (c *ConcurrentBlockingQueue[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isFull()
}

func (c *ConcurrentBlockingQueue[T]) isFull() bool {
	return c.count == c.maxSize
}
