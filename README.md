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

### Environment Variable

| Environment Variable Name | Environment Variable Default Value | Environment Variable Introduction |
| ------------------------- | ---------------------------------- | --------------------------------- |
| serverAddr                | 127.0.0.1                          | nacos server address              |
| serverPort                | 8848                               | nacos server port                 |
| namespace                 |                                    | the namespaceId of nacos          |
| configDataId              | {{.ClientServiceName}}.{{.ServerServiceName}}.{{.Category}}  | the  format of config data id          |
| configGroup               | DEFAULT_GROUP                      | the group of config data          |

### More Info

Refer to [example](https://github.com/kitex-contrib/config-nacos/tree/main/example) for more usage.


## Compatibility
This Package use Nacos1.x client. The Nacos2.0 and Nacos1.0 Server are fully compatible with it. [see](https://nacos.io/en-us/docs/v2/upgrading/2.0.0-compatibility.html)

maintained by: [whalecold](https://github.com/whalecold)

