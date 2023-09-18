# config-nacos 

[English](https://github.com/kitex-contrib/config-nacos/blob/main/README.md)

使用 **nacos** 作为 **Kitex** 的配置中心

##  这个项目应当如何使用?

### 基本使用

#### 服务端

```go
package main

import (
	"context"
	"log"

	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/config-nacos/nacos"
	nacosserver "github.com/kitex-contrib/config-nacos/server"
)

var _ api.Echo = &EchoImpl{}

// EchoImpl implements the last service interface defined in the IDL.
type EchoImpl struct{}

// Echo implements the Echo interface.
func (s *EchoImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	klog.Info("echo called")
	return &api.Response{Message: req.Message}, nil
}

func main() {
	klog.SetLevel(klog.LevelDebug)
	nacosClient, err := nacos.New(nacos.Options{})
	if err != nil {
		panic(err)
	}
	serviceName := "server"
	svr := echo.NewServer(
		new(EchoImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
		server.WithSuite(nacosserver.NewSuite(serviceName, nacosClient)),
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
package main

import (
	"context"
	"log"

	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	nacosclient "github.com/kitex-contrib/config-nacos/client"
	"github.com/kitex-contrib/config-nacos/nacos"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

func main() {
	klog.SetLevel(klog.LevelDebug)

	nacosClient, err := nacos.New(nacos.Options{})
	if err != nil {
		panic(err)
	}

	fn := func(cp *vo.ConfigParam) {
		klog.Infof("nacos config %v", cp)
	}

	serviceName := "server"
	clientName := "client"
	client, err := echo.NewClient(
		serviceName,
		client.WithHostPorts("0.0.0.0:8888"),
		client.WithSuite(nacosclient.NewSuite(serviceName, clientName, nacosClient, fn)),
	)
	if err != nil {
		log.Fatal(err)
	}
	for {
		req := &api.Request{Message: "my request"}
		resp, err := client.Echo(context.Background(), req)
		if err != nil {
			klog.Errorf("take request error: %v", err)
		} else {
			klog.Infof("receive response %v", resp)
		}
	}
}
```
### Nacos 配置

根据 Options 的参数初始化 client, 如果参数为空则会根据环境变量获取到 nacos 的 addr, port 以及 namespace 链接到 nacos 服务器上，建立链接之后 suite 会根据 configGroup 以及 configDataId 订阅对应的配置并动态更新自身策略，具体参数参考下面环境变量。 

配置的格式默认支持 `json` 和 `yaml`，可以使用函数 `SetParser` 进行自定义格式解析方式，并在 `NewSuite` 的时候使用 `CustomFunction` 函数修改订阅函数的格式。

#### CustomFunction

允许用户自定义 nacos 的参数. 

#### 环境变量


| 变量名 | 变量默认值 | 作用 |
| ------------------------- | ---------------------------------- | --------------------------------- |
| KITEX_CONFIG_NACOS_SERVER_ADDR               | 127.0.0.1                          | nacos 服务器地址 |
| KITEX_CONFIG_NACOS_SERVER_PORT               | 8848                               | nacos 服务器端口            |
| KITEX_CONFIG_NACOS_NAMESPACE                 |                                    | nacos 中的 namespace Id |
| KITEX_CONFIG_NACOS_DATA_ID              | {{.ClientServiceName}}.{{.ServerServiceName}}.{{.Category}}  | 使用 go [template](https://pkg.go.dev/text/template) 语法渲染生成对应的 ID, 使用 `ClientServiceName` `ServiceName` `Category` 三个元数据，可以自定义          |
| KITEX_CONFIG_NACOS_GROUP               | DEFAULT_GROUP                      | 使用固定值，也可以动态渲染，用法同 configDataId          |

#### 治理策略

下面例子中的 configDataId 以及 configGroup 均使用默认值，服务名称为 ServiceName，客户端名称为 ClientName

##### 限流 Category=limit
> 限流目前只支持服务端，所以 ClientServiceName 为空。

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/limiter/item_limiter.go#L33)

|字段|说明|
|----|----|
|connection_limit|最大并发数量| 
|qps_limit|每 100ms 内的最大请求数量| 

例子：
```
configDataID: ServiceName.limit
{
  "connection_limit": 100, 
  "qps_limit": 2000        
}
```
注：

- 限流配置的粒度是 Server 全局，不分 client、method
- 「未配置」或「取值为 0」表示不开启
- connection_limit 和 qps_limit 可以独立配置，例如 connection_limit = 100, qps_limit = 0

##### 重试 Category=retry

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/retry/policy.go#L63)

|参数|说明|
|----|----|
|type| 0: failure_policy 1: backup_policy| 
|failure_policy.backoff_policy| 可以设置的策略： `fixed` `none` `random` | 

例子：
```
configDataId: ClientName.ServiceName.retry
{
    "*": {  
        "enable": true,
        "type": 0,                 
        "failure_policy": {
            "stop_policy": {
                "max_retry_times": 3,
                "max_duration_ms": 2000,
                "cb_policy": {
                    "error_rate": 0.5
                }
            },
            "backoff_policy": {
                "backoff_type": "fixed", 
                "cfg_items": {
                    "fix_ms": 50
                }
            },
            "retry_same_node": false
        }
    },
    "echo": { 
        "enable": true,
        "type": 1,                 
        "backup_policy": {
            "retry_delay_ms": 100,
            "retry_same_node": false,
            "stop_policy": {
                "max_retry_times": 2,
                "max_duration_ms": 300,
                "cb_policy": {
                    "error_rate": 0.2
                }
            }
        }
    }
}
```
注：retry.Container 内置支持用 * 通配符指定默认配置（详见 [getRetryer](https://github.com/cloudwego/kitex/blob/v0.5.1/pkg/retry/retryer.go#L240) 方法）

##### 超时 Category=rpc_timeout

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/rpctimeout/item_rpc_timeout.go#L42)

例子：
```
configDataId: ClientName.ServiceName.rpc_timeout
{
  "*": {
    "conn_timeout_ms": 100, 
    "rpc_timeout_ms": 3000
  },
  "echo": {
    "conn_timeout_ms": 50,
    "rpc_timeout_ms": 1000
  }
}
```
注：kitex 的熔断实现目前不支持修改全局默认配置（详见 [initServiceCB](https://github.com/cloudwego/kitex/blob/v0.5.1/pkg/circuitbreak/cbsuite.go#L195)）

##### 熔断: Category=circuit_break

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/circuitbreak/item_circuit_breaker.go#L30)

|参数|说明|
|----|----|
|min_sample| 最小的统计样本数| 
例子：
```
Echo 方法使用下面的配置（0.3、100），其他方法使用全局默认配置（0.5、200）
configDataId: `ClientName.ServiceName.circuit_break`
{
  "Echo": {
    "enable": true,
    "err_rate": 0.3, 
    "min_sample": 100 
  }
}
```

### 更多信息

更多示例请参考 [example](https://github.com/kitex-contrib/config-nacos/tree/main/example)

## 兼容性
该包使用 Nacos1.x 客户端，Nacos2.0 和 Nacos1.0 服务端完全兼容该版本. [详情](https://nacos.io/zh-cn/docs/v2/upgrading/2.0.0-compatibility.html)

主要贡献者： [whalecold](https://github.com/whalecold)
