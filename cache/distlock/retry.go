package distlock

import "time"

type RetryStrategy interface {
	// Next 第一个返回重试的间隔，第二个表示是否要继续重试
	Next() (time.Duration, bool)
}

type FixedRetryStrategy struct {
	Interval time.Duration
	MaxCnt   int
	cnt      int
}

func (f *FixedRetryStrategy) Next() (time.Duration, bool) {
	if f.cnt >= f.MaxCnt {
		return 0, false
	}
	return f.Interval, true
}
