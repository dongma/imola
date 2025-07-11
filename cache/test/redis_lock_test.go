package test

import (
	"context"
	"github.com/dongma/imola/cache/distlock"
	"github.com/dongma/imola/cache/mocks"
	"github.com/golang/mock/gomock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestClient_Lock(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) redis.Cmdable
		key      string
		wantErr  error
		wantLock *distlock.Lock
	}{
		{
			name: "set nx error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, context.DeadlineExceeded)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "failed to preempt lock",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(false, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key:     "key1",
			wantErr: distlock.ErrFailedToPreemptLock,
		},
		{
			name: "locked",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewBoolResult(true, nil)
				cmd.EXPECT().SetNX(context.Background(), "key1", gomock.Any(), time.Minute).
					Return(res)
				return cmd
			},
			key: "key1",
			wantLock: &distlock.Lock{
				Key:        "key1",
				Expiration: time.Minute,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := distlock.NewClient(tc.mock(ctrl))
			l, err := client.TryLock(context.Background(), tc.key, time.Minute)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantLock.Key, l.Key)
			assert.Equal(t, tc.wantLock.Expiration, l.Expiration)
			assert.NotEmpty(t, l.Value)
		})
	}
}

func TestClient_UnLock(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		key     string
		value   string
		wantErr error
	}{
		{
			name: "eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), distlock.LuaUnlock, []string{"key1"}, []any{"value1"}).
					Return(res)
				return cmd
			},
			key:     "key1",
			value:   "value1",
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), distlock.LuaUnlock, []string{"key1"}, []any{"value1"}).
					Return(res)
				return cmd
			},
			key:     "key1",
			value:   "value1",
			wantErr: distlock.ErrLockNotHold,
		},
		{
			name: "unlocked",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), distlock.LuaUnlock, []string{"key1"}, []any{"value1"}).
					Return(res)
				return cmd
			},
			key:   "key1",
			value: "value1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &distlock.Lock{
				Key:    tc.key,
				Value:  tc.value,
				Client: tc.mock(ctrl),
			}
			err := lock.Unlock(context.Background())
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestClient_Refresh(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) redis.Cmdable
		key        string
		value      string
		wantErr    error
		expiration time.Duration
	}{
		{
			name: "eval error",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetErr(context.DeadlineExceeded)
				cmd.EXPECT().Eval(context.Background(), distlock.LuaRefresh, []string{"key1"}, []any{"value1", float64(60)}).
					Return(res)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Minute,
			wantErr:    context.DeadlineExceeded,
		},
		{
			name: "lock not hold",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(0))
				cmd.EXPECT().Eval(context.Background(), distlock.LuaRefresh, []string{"key1"}, []any{"value1", float64(60)}).
					Return(res)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Minute,
			wantErr:    distlock.ErrLockNotHold,
		},
		{
			name: "refreshed",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				res := redis.NewCmd(context.Background())
				res.SetVal(int64(1))
				cmd.EXPECT().Eval(context.Background(), distlock.LuaRefresh, []string{"key1"}, []any{"value1", float64(60)}).
					Return(res)
				return cmd
			},
			key:        "key1",
			value:      "value1",
			expiration: time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			lock := &distlock.Lock{
				Key:        tc.key,
				Value:      tc.value,
				Client:     tc.mock(ctrl),
				Expiration: tc.expiration,
			}
			err := lock.Refresh(context.Background())
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func ExampleLock_Refresh() {
	// 加锁成功，你拿到了一个Lock
	var l *distlock.Lock
	stopChan := make(chan struct{})
	errChan := make(chan error)
	timeoutChan := make(chan struct{}, 1)

	go func() {
		timeoutRetry := 0
		// 间隔多久续约一次
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				// 定时刷新
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := l.Refresh(ctx)
				// 出现err后，也要向errChan发消息，同时关闭channel
				if err == context.DeadlineExceeded {
					timeoutChan <- struct{}{}
					continue
				}
				if err != nil {
					errChan <- err
					close(errChan)
					close(stopChan)
					return
				}
				timeoutRetry = 0
			case <-timeoutChan:
				timeoutRetry++
				if timeoutRetry > 10 {
					errChan <- context.DeadlineExceeded
					return
				}
				// 定时刷新
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := l.Refresh(ctx)
				// 出现err后，也要向errChan发消息，同时关闭channel
				if err == context.DeadlineExceeded {
					timeoutChan <- struct{}{}
					continue
				}
				if err != nil {
					errChan <- err
					close(errChan)
					close(stopChan)
					return
				}
			case <-stopChan:
				l.Unlock(context.Background())
				return
			}
		}
	}()

	// 执行业务逻辑，等执行完成后，就要退出续约的循环
	for i := 0; i < 100; i++ {
		select {
		// 在这里，续约失败
		case <-errChan:
			break
		default:
		}
	}
	stopChan <- struct{}{}
}

func ExampleLock_AutoRefresh() {
	// 加锁成功，你拿到了一个Lock
	var l *distlock.Lock
	go func() {
		// 这里返回error了，你要中断业务
		l.AutoRefresh(time.Second*10, time.Second)
	}()
}
