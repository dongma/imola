-- 检查keys[1]是不是你的锁，如果确实是的，则del删除
-- KEYS[1] 就是分布式锁的key, ARGV[1] 就是你预期存在redis里面的value
if redis.call('get', KEYS[1]) == ARGV[1] then
    return redis.call('del', KEYS[1])
else
    return 0
end