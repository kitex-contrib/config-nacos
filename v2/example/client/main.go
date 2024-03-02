// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	nacosclient "github.com/kitex-contrib/config-nacos/v2/client"
	"github.com/kitex-contrib/config-nacos/v2/nacos"
	"github.com/kitex-contrib/config-nacos/v2/utils"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"log"
	"time"
)

type configLog struct{}

func (cl *configLog) Apply(opt *utils.Options) {
	fn := func(cp *vo.ConfigParam) {
		klog.Infof("nacos config %v", cp)
	}
	opt.NacosCustomFunctions = append(opt.NacosCustomFunctions, fn)
}

func main() {
	nacosClient, err := nacos.NewClient(nacos.Options{})
	if err != nil {
		panic(err)
	}
	fmt.Println(nacosClient)

	cl := &configLog{}

	serviceName := "ServiceName" // your server-side service name
	clientName := "ClientName"   // your client-side service name
	client, err := echo.NewClient(
		serviceName,
		client.WithHostPorts("0.0.0.0:8888"),
		client.WithSuite(nacosclient.NewSuite(serviceName, clientName, nacosClient, cl)),
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
		time.Sleep(time.Second * 10)
	}
}
