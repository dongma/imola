# imola
<p align="center">
    <img align="center" width="138px" src="doc/imgs/424_cartoons-gophers.png">
</p>

《`go`实战训练营》学习过程中写的代码，主要包括：`server`路由注册、`orm`中SQL生成与事务操作、分布式缓存和`go`微服务这`4`部分，具体功能实现有：手工写`web`路由树，自己实现`orm`生成复杂查询的`sql`，此外，使用`grpc`和`etcd`结合实现服务注册与发现，收获比较多，对后续学习`go`相应框架如`gin`、`gorm`和`gozero`等时有指导性意义。
### `web`路由注册
对比现有的`Beego`、`Gin`和`Echo`，对于一个`web`框架来说，至少要提供三个抽象：`Server`-代表服务器的抽象，`Context`-代表上下文的抽象以及路由树，详细设计请阅读 [web服务的设计](doc/web服务的设计.md)。
```go
import (
    "imola/web"
)

func main() {
    var server := web.NewHTTPServer()
	server.GET("/user/123", func(ctx *web.Context) {
        ctx.RespJSON(202, User{
            Name: "Tom",
        })
    })
    server.Start(":8081")    
}
```
在命令行调用接口，返回如下:
```shell
% curl 'http://localhost:8081/user/123'
{"name":"Tom"}% 
```

### 操作数据库
实现自定义的`orm`框架，根据`model`对象生成的查询语句支持`mysql`、`sqlite`以及可扩展`postgre`，同时也支持`native sql`查询，并映射结果集到`model`对象，详细设计请阅读 [orm框架的设计](doc/orm服务的设计.md)。
```go
import (
    "context"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "github.com/stretchr/testify/require"
    "log"
    "testing"
    "time"
)

func main() {
    db, err := sql.Open("sqlite3", "file:test.db?cache=shared&mode=memory")
    require.NoError(t, err)
    defer db.Close()

    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    // 操作原生sql
	res, err := db.ExecContext(ctx, "INSERT INTO `test_model` VALUES(?, ?, ?, ?)", 1, "Tom", 18, "Jerry")
    require.NoError(t, err)

    // 执行查询语句
    row := db.QueryRowContext(ctx,
    "select id, first_name, age, last_name from `test_model` where id = ?", 1)
    require.NoError(t, row.Err())
    tm := TestModel{}
    err = row.Scan(&tm.Id, &tm.FirstName, &tm.Age, &tm.LastName)
}
```
### `Cache`
业务但凡对性能有点要求，几乎都会考虑使用缓存。缓存大体上分成两类：本地缓存和分布式缓存，如`Redis`、`memecache`,详细设计请阅读 [cache服务的设计](doc/cache服务的设计.md)。
```go
import (
    "context"
    "github.com/redis/go-redis/v9"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "imola/cache"
    "testing"
    "time"
)
func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr: "127.0.0.1:6379",
    })
    rcache := cache.NewRedisCache(rdb)
    ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
    defer cancel()
    err := rcache.Set(ctx, "key1", "value1", time.Minute)
    require.NoError(t, err)
    val, err := rcache.Get(ctx, "key1")
    require.NoError(t, err)
    assert.Equal(t, "value1", val)
}
```