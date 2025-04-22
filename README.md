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
`orm`支持`native sql`，也支持基于`struct`对象进行操作，
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