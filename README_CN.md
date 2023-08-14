# config-nacos 

[English](https://github.com/kitex-contrib/config-nacos/blob/main/README.md)

使用 **nacos** 作为 **Kitex** 的配置中心

##  这个项目应当如何使用?

### 基本使用

#### 服务端

```go
import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"

	"github.com/kitex-contrib/registry-nacos/registry"
	"github.com/kitex-contrib/config-nacos/nacos"
	nacosserver "github.com/kitex-contrib/config-nacos/server"
)

var _ api.Echo = &EchoImpl{}

// EchoImpl implements the last service interface defined in the IDL.
type EchoImpl struct{}

// Echo implements the Echo interface.
func (s *EchoImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	klog.Info("echo called")
	time.Sleep(2 * time.Second)
	return &api.Response{Message: req.Message}, nil
}

func main() {
	r, err := registry.NewDefaultNacosRegistry()
	if err != nil {
		panic(err)
	}
	nacosClient, err := nacos.DefaultClient()
	if err != nil {
		panic(err)
	}

	serviceName := "echo"

	opts := []server.Option{
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
		server.WithRegistry(r),
	}

	opts = append(opts, nacosserver.NewSuite(serviceName, nacosClient).Options()...)

	svr := echo.NewServer(
		new(EchoImpl),
		opts...,
	)
	if err := svr.Run(); err != nil {
		log.Println("server stopped with error:", err)
	} else {
		log.Println("server stopped")
	}
}

```

#### 客户端

```go
import (
    // ...
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/client"
	nacosclient "github.com/kitex-contrib/config-nacos/client"
	"github.com/kitex-contrib/config-nacos/nacos"
    // ...
)

func main() {
    // ... 
	nacosClient, err := nacos.DefaultClient()
	if err != nil {
		panic(err)
	}
	fn := func(cp *vo.ConfigParam) {
		cp.Type = vo.TEXT
	}
	opts := []client.Option{
		client.WithHostPorts("0.0.0.0:8888"),
		client.WithMiddleware(mymiddleware.CommonMiddleware),
		client.WithMiddleware(mymiddleware.ClientMiddleware),
		//client.WithResolver(r),
	}

	opts = append(opts, nacosclient.NewSuite("echo", "test", nacosClient, fn).Options()...)

	client, err := echo.NewClient(
		"echo",
		opts...,
	)
    // ...
}
```

### 环境变量

| 变量名 | 变量默认值 | 作用 |
| ------------------------- | ---------------------------------- | --------------------------------- |
| serverAddr               | 127.0.0.1                          | nacos 服务器地址 |
| serverPort               | 8848                               | nacos 服务器端口            |
| namespace                 |                                    | nacos 中的 namespace Id |
| configDataId              | {{.ClientServiceName}}.{{.ServerServiceName}}.{{.Category}}  | the  format of config data id          |
| configGroup               | DEFAULT_GROUP                      | the group of config data          |

### 更多信息

更多示例请参考 [example](https://github.com/kitex-contrib/config-nacos/tree/main/example)

## 兼容性
该包使用 Nacos1.x 客户端，Nacos2.0 和 Nacos1.0 服务端完全兼容该版本. [详情](https://nacos.io/zh-cn/docs/v2/upgrading/2.0.0-compatibility.html)

主要贡献者： [whalecold](https://github.com/whalecold)
