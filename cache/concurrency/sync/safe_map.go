package sync

import "sync"

type SafeMap[K comparable, V any] struct {
	Data  map[K]V
	Mutex sync.RWMutex
}

// Put 将key->value键值对 放入map中，defer 用于释放锁(无论是否发生异常)
func (s *SafeMap[K, V]) Put(key K, value V) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Data[key] = value
}

// Get 用于从map中按照key读取值
func (s *SafeMap[K, V]) Get(key K) (any, bool) {
	s.Mutex.RLock()
	defer s.Mutex.RUnlock()
	res, ok := s.Data[key]
	return res, ok
}

// LoadOrStore key存在于map中时，则独去出数据；否则，存储到map中
func (s *SafeMap[K, V]) LoadOrStore(key K, newVal V) (val V, loaded bool) {
	// 用RLock(读锁)先检查一次
	s.Mutex.RLock()
	res, ok := s.Data[key]
	if ok {
		return res, true
	}
	s.Mutex.RUnlock()

	// 再用Mutex写锁再检查一次
	s.Mutex.Lock()
	res, ok = s.Data[key]
	if ok {
		return res, true
	}
	defer s.Mutex.Unlock()
	s.Data[key] = newVal
	return newVal, false
}
