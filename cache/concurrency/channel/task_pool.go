package channel

import "context"

type Task func()

type TaskPool struct {
	tasks chan Task
	close chan struct{}
}

// NewTaskPool numG是goroutine的数量，你要控制住的
// capacity 是缓存的容量
func NewTaskPool(numG int, capacity int) *TaskPool {
	result := &TaskPool{
		tasks: make(chan Task, capacity),
		close: make(chan struct{}),
	}

	// 要是没有退出goroutine的机制，那就是妥妥的goroutine泄漏
	for i := 0; i < numG; i++ {
		go func() {
			select {
			case <-result.close:
				return
			case t := <-result.tasks:
				t()
			}
		}()
	}
	return result
}

// Submit 提交任务
func (p *TaskPool) Submit(ctx context.Context, task Task) error {
	select {
	case p.tasks <- task:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// Close 方法会释放资源，不要重复调用
func (p *TaskPool) Close() error {
	// close方法被重复调用的话，会产生panic
	close(p.close)
	return nil
}
