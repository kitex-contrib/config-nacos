# config-nacos (*This is a community driven project*)

[中文](https://github.com/kitex-contrib/config-nacos/blob/main/README_CN.md)

Nacos as config centre.

## How to use?

### Basic

#### Server

```go
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
	nacosClient, err := nacos.DefaultClient()
	if err != nil {
		panic(err)
	}
	serviceName := "echo"

	opts := []server.Option{
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
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

#### Client

```go
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

	nacosClient, err := nacos.DefaultClient()
	if err != nil {
		panic(err)
	}

	fn := func(cp *vo.ConfigParam) {
		klog.Infof("nacos config %v", cp)
	}

	opts := []client.Option{
		client.WithHostPorts("0.0.0.0:8888"),
	}

	serviceName := "echo"
	clientName := "test"

	opts = append(opts, nacosclient.NewSuite(serviceName, clientName, nacosClient, fn).Options()...)

	client, err := echo.NewClient(
		serviceName,
		opts...,
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

### Nacos Configuration

The client obtains the nacos address, port and namespace from the environment variables and connects to the nacos server. After the connection is established, the suite subscribes to the appropriate configuration based on configGroup and configDataId and dynamically updates its own policy. See the environment variables below for specific parameters.

The configured format supports `json` and `yaml` by default. You can use the `SetParser` function to customise the format parsing method, and the `CustomFunction` function to customise the format of the subscription function during `NewSuite`.
####

#### Environment Variable

| Environment Variable Name | Environment Variable Default Value | Environment Variable Introduction |
| ------------------------- | ---------------------------------- | --------------------------------- |
| serverAddr                | 127.0.0.1                          | nacos server address              |
| serverPort                | 8848                               | nacos server port                 |
| namespace                 |                                    | the namespaceId of nacos          |
| configDataId              | {{.ClientServiceName}}.{{.ServerServiceName}}.{{.Category}}  | Use go [template](https://pkg.go.dev/text/template) syntax rendering to generate the appropriate ID, and use `ClientServiceName` `ServiceName` `Category` three metadata that can be customised          |
| configGroup               | DEFAULT_GROUP                      | Use fixed values or dynamic rendering. Usage is the same as configDataId.          |

#### Governance Policy
> The configDataId and configGroup in the following example use default values, the service name is echo and the client name is requester.

##### Rate Limit Category=limit
> Currently, current limiting only supports the server side, so ClientServiceName is empty.

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/limiter/item_limiter.go#L33)

Example:
```
configDataID: .echo.limit
{
  "connection_limit": 100, // Maximum 100 concurrent connections
  "qps_limit": 2000        // Maximum 2000 QPS per 100ms
}
```

Note:

- The granularity of the current limit configuration is server global, regardless of client or method.
- Not configured or value is 0 means not enabled.
- connection_limit and qps_limit can be configured independently, e.g. connection_limit = 100, qps_limit = 0

##### Retry Policy Category=retry
[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/retry/policy.go#L63)

Example：
```
configDataId: requester.echo.retry
{
    "*": {  // * default value, If you do not configure all fallbacks to this policy
        "enable": true,
        "type": 0,        // failed retry(type=0)         
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
            }
        }
    },
    "echo": { // the method, lower-case
        "enable": true,
        "type": 1,  // backoff_policy(type=1)               
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
            }
        }
    }
}
```
Note: retry.Container has built-in support for specifying the default configuration using the * wildcard (see the [getRetryer](https://github.com/cloudwego/kitex/blob/v0.5.1/pkg/retry/retryer.go#L240) method for details).

##### RPC Timeout Category=rpc_timeout

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/rpctimeout/item_rpc_timeout.go#L42)

Example：
```
configDataId: requester.echo.rpc_timeout
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
Note: The circuit breaker implementation of kitex does not currently support changing the global default configuration (see [initServiceCB](https://github.com/cloudwego/kitex/blob/v0.5.1/pkg/circuitbreak/cbsuite.go#L195) for details).

##### Circuit Break: Category=circuit_break

[JSON Schema](https://github.com/cloudwego/kitex/blob/develop/pkg/circuitbreak/item_circuit_breaker.go#L30)

Example：
```
The Echo method uses the following configuration (0.3, 100) and other methods use the global default configuration (0.5, 200)
configDataId: `requester.echo.circuit_break`
{
  "Echo": {
    "enable": true,
    "err_rate": 0.3, 
    "min_sample": 100 
  }
}
```
### More Info

Refer to [example](https://github.com/kitex-contrib/config-nacos/tree/main/example) for more usage.


## Compatibility
This Package use Nacos1.x client. The Nacos2.0 and Nacos1.0 Server are fully compatible with it. [see](https://nacos.io/en-us/docs/v2/upgrading/2.0.0-compatibility.html)

maintained by: [whalecold](https://github.com/whalecold)

