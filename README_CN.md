# config-nacos 

[English](https://github.com/kitex-contrib/config-nacos/blob/main/README.md)

使用 **nacos** 作为 **Kitex** 的配置中心

##  这个项目应当如何使用?

### 基本使用

#### 客户端

```go
import (
    // ...
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/client"
	retry "github.com/kitex-contrib/config-nacos/client"
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

	opts = append(opts, retry.NewSuite("echo", "test", nacosClient, fn).Options()...)

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

更多示例请参考 [example](example)

## 兼容性
该包使用 Nacos1.x 客户端，Nacos2.0 和 Nacos1.0 服务端完全兼容该版本. [详情](https://nacos.io/zh-cn/docs/v2/upgrading/2.0.0-compatibility.html)
