package queue

import "context"

type Queue[T any] interface {
	// Enqueue 将elem压入到queue中, ctx中有withTimeout选项 (Go的设计风格)
	Enqueue(ctx context.Context, elem T) error

	// Dequeue 从queue中弹出元素，ctx的超时能做到级连控制
	Dequeue(ctx context.Context) (T, error)

	IsFull() bool

	IsEmpty() bool

	Len() uint64
}
