package queue

import (
	"context"
	"sync"
	"time"
)

type DelayQueue[T Delayable] struct {
	pq            *PriorityQueue[T]
	mu            sync.RWMutex
	dequeueSignal *cond
	enqueueSignal *cond
}

func NewDelayQueue[T Delayable](capacity int) *DelayQueue[T] {
	return &DelayQueue[T]{
		pq: NewPriorityQueue[T](capacity, func(src T, dst T) int {
			srcDelay := src.Delay()
			dstDelay := dst.Delay()
			if srcDelay > dstDelay {
				return 1
			}
			if srcDelay == dstDelay {
				return 0
			}
			return -1
		}),
	}
}

// Enqueue 元素压入到堆栈中
func (d *DelayQueue[T]) Enqueue(ctx context.Context, data T) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 如果入队后的元素，过期时间更短，要么就要唤醒出队的
		// 或者，一点都不管,就直接唤醒出队
		d.mu.Lock()

		err := d.pq.Enqueue(data)
		switch err {
		case nil:
			// 入队成功，发送入队信号，唤醒出队阻塞的
			d.enqueueSignal.broadcast()
			return nil
		case ErrOutOfCapacity:
			// 阻塞，开始睡觉了
			ch := d.enqueueSignal.signalCh()
			select {
			case <-ch:
			case <-ctx.Done():
				return ctx.Err()
			}
		default:
			d.mu.Unlock()
			return err
		}
	}
}

// DeQueue 出队就有讲究了，Delay返回<=0时才能出队
// 2、如果队首Delay=300ms >0, 要是sleep, 等待Delay()降下去
// 3、如果正在sleep的过程，有新元素来了，并且Delay()=200比你正在sleep的时间还短,你要调整你的sleep时间
// 4、如果sleep的时间还没到，就超时了，那么就返回; sleep本质上是阻塞（你可以用time.sleep，也可以使用channel）
func (c *DelayQueue[T]) DeQueue(ctx context.Context) (T, error) {
	var timer *time.Timer
	for {
		select {
		case <-ctx.Done():
			var t T
			return t, ctx.Err()
		default:
		}

		// 我该干啥?
		c.mu.Lock()

		// 主要是顾虑锁被人持有很久，以至于早就超时了
		select {
		case <-ctx.Done():
			var t T
			c.mu.Unlock()
			return t, ctx.Err()
		default:
		}

		// 我拿到堆顶
		val, err := c.pq.Peek()
		switch err {
		case nil:
			// 拿到堆顶元素了
			delay := val.Delay()
			if delay <= 0 {
				val, _ := c.pq.Dequeue()
				c.dequeueSignal.broadcast()
				return val, nil
			}
			// 要在这里解锁
			signal := c.enqueueSignal.signalCh()
			if timer == nil {
				timer = time.NewTimer(delay)
			} else {
				timer.Reset(delay)
			}

			// 你一定要在进去select之前解锁
			select {
			case <-timer.C:
			case <-ctx.Done():
				var t T
				return t, ctx.Err()
			case <-signal:
			}

		case ErrEmptyQueue:
			// 这个分支代表，队列为空
			signal := c.enqueueSignal.signalCh()
			// 你一定要在进去select之前解锁
			select {
			case <-ctx.Done():
				var t T
				return t, ctx.Err()
			case <-signal:
			}
		default:
			c.mu.Unlock()
			var t T
			return t, err
		}

	}
}

type Delayable interface {
	Delay() time.Duration
}

type cond struct {
	signal chan struct{}
	l      sync.Locker
}

func newCond(l sync.Locker) *cond {
	return &cond{
		signal: make(chan struct{}),
		l:      l,
	}
}

// broadcast 唤醒等待者，如果没有人等待，那么什么也不发生，必须加锁之后才能调用这个方法
// 广播之后锁会被释放，这也是为了确保用户必然是在锁范围内调用
func (c *cond) broadcast() {
	signal := make(chan struct{})
	old := c.signal
	c.signal = signal
	c.l.Unlock()
	close(old)
}

// signalCh 返回一个 channel，用于监听广播信号,必须在锁范围内使用
// 调用后，锁会被释放，这也是为了确保用户必然是在锁的范围内调用的
func (c *cond) signalCh() <-chan struct{} {
	res := c.signal
	c.l.Unlock()
	return res
}
