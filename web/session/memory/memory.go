package memory

import (
	"context"
	"errors"
	"github.com/dongma/imola/web/session"
	cache "github.com/patrickmn/go-cache"
	"sync"
	"time"
)

type Store struct {
	// 锁机制，避免多个goroutine同时来操作此id
	mutex sync.Mutex
	// 利用一个内存缓存协助我们管理过期时间
	cache      *cache.Cache
	expiration time.Duration
}

// NewStore 创建一个Store实例，也可以考虑用Option模式，允许用户控制过期检查的问题
func NewStore(expiration time.Duration) *Store {
	return &Store{
		cache:      cache.New(expiration, time.Second),
		expiration: expiration,
	}
}

func (m *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	session := &memorySession{
		id:   id,
		data: make(map[string]string),
	}
	m.cache.Set(session.ID(), session, m.expiration)
	return session, nil
}

// Refresh 刷新同一个session id，使session不失效
func (m *Store) Refresh(ctx context.Context, id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	session, ok := m.cache.Get(id)
	if !ok {
		return errors.New("session not found")
	}
	m.cache.Set(session.(*memorySession).ID(), session, m.expiration)
	return nil
}

func (m *Store) Remove(ctx context.Context, id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cache.Delete(id)
	return nil
}

// Get 获取session信息
func (m *Store) Get(ctx context.Context, id string) (session.Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	session, ok := m.cache.Get(id)
	if !ok {
		return nil, errors.New("session not found")
	}
	return session.(*memorySession), nil
}

type memorySession struct {
	mutex sync.RWMutex
	id    string
	data  map[string]string
}

func (m *memorySession) Get(ctx context.Context, key string) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	val, ok := m.data[key]
	if !ok {
		return "", errors.New("找不到这个key")
	}
	return val, nil
}

func (m *memorySession) Set(ctx context.Context, key string, val string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[key] = val
	return nil
}

func (m *memorySession) ID() string {
	return m.id
}
