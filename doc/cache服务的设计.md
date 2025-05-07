## Cache的设计
在`Cache`的API中，我们依旧保持在公开方法中接收一个`context`参数，这个参数在本地缓存中没用，但在接入`Redis`的时候就很有用。
```go
// Cache 屏蔽不同缓存中间件的差异
type Cache interface {
	// Get 按key从缓存中取出对应的值
	Get(ctx context.Context, key string) (any, error)

	// Set 指定key，在一定expiration内缓存value值
	Set(ctx context.Context, key string, value any, expiration time.Duration) error

	// Delete 缓存中删除指定的key
	Delete(ctx context.Context, key string) error

	// LoadAndDelete 缓存中加载key并返回，并删除key
	LoadAndDelete(ctx context.Context, key string) (any, error)

	OnEvicted(func(key string, val []byte))
}
```
缓存异常的3种情况，穿透、击穿和雪崩：
- 缓存穿透，读请求对应的数据根本不存在，因而每次都落到`DB`上，一般黑客使用非法的邮箱、ID等攻击数据库。`singleflight`能缓解问题，未命中`cache`时，再问下布隆过滤器、bit array等结构。
- 缓存击穿，缓存中没有对应`key`的数据，但是数据在`DB`里面是有的，所以要回写到缓存，此一次访问就是命中缓存，用`singleflight`就足以解决问题。
- 缓存雪崩，同一时刻，大量`key`国旗，查询都要回查数据库。在设置`key`过期时间时，加上一个随机的偏移值。

`singleflight`设计模式能够有效减轻对数据库的压力，在有多个`goroutine`试图去数据库加载同一个`key`的数据时，只允许一个`goroutine`过去查询，其它都在原地等待结果。
```go
// SFGet SingleFlightGet
func (r *ReadThroughCache) SFGet(ctx context.Context, key string) (any, error) {
	var group = &singleflight.Group{}
	val, err := r.Cache.Get(ctx, key)
	if errors.Is(err, ErrKeyNotFound) {
		// group.Do执行func允许单个goroutine去数据库加载数据
		val, err, _ = group.Do(key, func() (interface{}, error) {
			cval, cerr := r.LoadFunc(ctx, key)
			if cerr == nil {
				er := r.Cache.Set(ctx, key, cval, r.Expiration)
				if er != nil {
					return cval, fmt.Errorf("%w, 原因: %s", ErrFailedToRefreshCache, er.Error())
				}
			}
			return cval, cerr
		})
	}
	return val, err
}
```
用`mockgen`命令生成`Cmdable`客户端，`mockgen -package=mocks -destination=mocks/redis_cmdable.mock.go github.com/go-redis/redis/v9 Cmdable`。`-package`指生成`go`代码的包名称，`-destination`为目标位置，其余部分为要`mock`的接口。
## 分布式锁
`redis`的客户端使用`SetNX`实现分布式锁，在加锁时为避免死锁应考虑超时时间，但过期时间应该设置多长？可以考虑在锁还没有过期的时候，再一次延长过期时间。
```lua
-- ##### lock.lua
val = redis.call('get', KEYS[1])
if val == false then
    -- key不存在
    return redis.call('set', KEYS[1], ARGV[1], 'ex', ARGV[2])
elseif val == ARGV[1] then
    -- 你上次加锁成功了
    redis.call('expire', KEYS[1],  ARGV[2])
    return 'OK'
else
    -- 此时别人持有锁
    return ""
end

-- ##### unlock.lua，检查keys[1]是不是你的锁，如果确实是的，则del删除
-- KEYS[1] 就是分布式锁的key, ARGV[1] 就是你预期存在redis里面的value
if redis.call('get', KEYS[1]) == ARGV[1] then
    return redis.call('del', KEYS[1])
else
    return 0
end
```
缓存模式不能解决缓存不一致问题，数据不一致怎么办？要求强一致的业务就直接不要使用缓存，如何解决缓存不一致问题？本质上是无解的，只是看我们系统追求一致性究竟有多强。
