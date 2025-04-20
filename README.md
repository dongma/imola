# imola
<p align="center">
<img align="center" width="168px" src="doc/imgs/424_cartoons-gophers.png">
</p>

imola是《go实战训练营》学习过程中写的代码，主要包括：`server`路由注册、`orm`中SQL生成与事务操作、分布式缓存和`go`微服务这4部分，具体功能实现有：手工写`web`路由树，自己实现`orm`生成复杂查询的`sql`，此外，使用`grpc`和`etcd`结合实现服务注册与发现，收获比较多，对后续学习`go`相应框架如`gin`、`gorm`和`gozero`等时有指导性意义。
## 🤷‍ 在项目中引入imola
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