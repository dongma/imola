package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"imola/web/session"
	"time"
)

var errSessionNotExist = errors.New("redis-session: session不存在")

type StoreOption func(store *Store)

type Store struct {
	prefix     string
	client     redis.Cmdable
	expiration time.Duration
}

// NewStore 创建一个Store的实例，可以考虑使用Option设计模式，允许用户控制过期检查的问题
func NewStore(client redis.Cmdable, opts ...StoreOption) *Store {
	redisStore := &Store{
		client:     client,
		prefix:     "session",
		expiration: time.Minute * 15,
	}
	for _, opt := range opts {
		opt(redisStore)
	}
	return redisStore
}

func (s *Store) key(id string) string {
	return fmt.Sprintf("%s_%s", s.prefix, id)
}

// Generate 生成一个session
func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	const lua = `
redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
return redis.call("pexpire", KEYS[1], ARGV[3])
`
	key := s.key(id)
	_, err := s.client.Eval(ctx, lua, []string{key}, "_sess_id", id, s.expiration.Milliseconds()).Result()
	if err != nil {
		return nil, err
	}
	return &Session{
		key:    key,
		id:     id,
		client: s.client,
	}, nil
}

// Refresh 刷新同一个session id，使session不失效
func (s *Store) Refresh(ctx context.Context, id string) error {
	key := s.key(id)
	affected, err := s.client.Expire(ctx, key, s.expiration).Result()
	if err != nil {
		return err
	}
	if !affected {
		return errSessionNotExist
	}
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	_, err := s.client.Del(ctx, s.key(id)).Result()
	return err
}

// Get 获取session信息
func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	key := s.key(id)
	// 这里不需要考虑并发的问题，因为在你检测的当下，没有就是没有
	i, err := s.client.Exists(ctx, id).Result()
	if err != nil {
		return nil, err
	}
	if i < 0 {
		return nil, errors.New("redis-session: session不存在")
	}
	return &Session{
		id:     id,
		key:    key,
		client: s.client,
	}, nil
}

type Session struct {
	key    string
	id     string
	client redis.Cmdable
}

func (s *Session) Get(ctx context.Context, key string) (string, error) {
	return s.client.HGet(ctx, s.key, key).Result()
}

func (s *Session) Set(ctx context.Context, key string, val string) error {
	const lua = `
if redis.call("exists", KEYS[1])
then
	return redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
else
	return -1
end
`
	res, err := s.client.Eval(ctx, lua, []string{s.key}, key, val).Int()
	if err != nil {
		return err
	}
	if res < 0 {
		return errSessionNotExist
	}
	return nil
}

func (s *Session) ID() string {
	return s.id
}
