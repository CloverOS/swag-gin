# swag-gin

<img align="right" width="180px" src="https://raw.githubusercontent.com/swaggo/swag/master/assets/swaggo.png">

本项目基于[swag](https://github.com/swaggo/swag)项目修改而成，基于gin的web框架，自动生成路由注册文件，无需手动注册各种路由。

# 快速开始

1. 将注释添加到API源代码中，请参阅声明性注释格式。
2. 使用如下命令下载swag：

```bash
$ go install github.com/CloverOS/swag-gin/cmd/swag-gin@latest
```

接下去使用的方法与[swag](https://github.com/swaggo/swag)一致

在main.go文件目录下执行以下命令，一键生成基于gin的路由注册文件。

```bash
swag-gin init --aco=true --ag=true
```

执行以上命令后，自动扫描整个会自动在原本生成的docs文件夹基础上新建了一个resource.go文件，里面包含了所有的用文档注释的接口信息，可用于管理接口权限等等。
另外会在每个服务模块的文件夹外层生成一个router.go
里面包含了两个函数(
按照 [testdata中的simple文件夹为例](https://github.com/CloverOS/swag-gin/tree/master/testdata/simple))

```go
func InitPublicRouter(r *gin.RouterGroup) {
r.GET("/testapi/get-string-by-int2/{some_id}", api.GetStringByInt2)
}

func InitPrivateRouter(r *gin.RouterGroup) {
r.GET("/testapi/get-string-by-int/{some_id}", api.GetStringByInt)
r.POST("/testapi/get-string-by-int3/{some_id}", api.GetStringByInt3)
}
```

1、这两个函数分别是公共接口和私有接口的路由注册函数。

2、判断私有接口和公共接口的依据是文档注释中的@Security，如果有的话则判断为私有接口，否则为公共接口。

3、之后就可以调用这个文件的InitPublicRouter和InitPrivateRouter函数注册到ginHandler中,不需要一个个手动写入了。

# 基于SwagCli添加的额外功能

## swag-gin cli

```bash
swag-gin init -h
NAME:
   swag init - Create docs.go

USAGE:
   swag init [command options] [arguments...]

OPTIONS:
   --autoRegisterGinRouter true\false,--ag 是否开启自动生成路由注册文件
   --ginServerPackage value, --pkg  	  指定路由注册文件的包名
   --ginRouterPath value, --rp            路由注册文件的生成路径文件名,默认"./router.go"
```