// Copyright 2024 CloudWeGo Authors
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

package client

import (
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/rpctimeout"
	"github.com/kitex-contrib/config-nacos/nacos"
	"github.com/kitex-contrib/config-nacos/utils"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// WithRPCTimeout sets the RPC timeout policy from nacos configuration center.
func WithRPCTimeout(dest, src string, nacosClient nacos.Client, opts utils.Options) []client.Option {
	param, err := nacosClient.ClientConfigParam(&nacos.ConfigParamConfig{
		Category:          rpcTimeoutConfigName,
		ServerServiceName: dest,
		ClientServiceName: src,
	})
	if err != nil {
		panic(err)
	}

	for _, f := range opts.NacosCustomFunctions {
		f(&param)
	}

	uniqueID := nacos.GetUniqueID()

	return []client.Option{
		client.WithTimeoutProvider(initRPCTimeoutContainer(param, dest, nacosClient, uniqueID)),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			return nacosClient.DeregisterConfig(param, uniqueID)
		}),
	}
}

func initRPCTimeoutContainer(param vo.ConfigParam, dest string,
	nacosClient nacos.Client, uniqueID int64,
) rpcinfo.TimeoutProvider {
	rpcTimeoutContainer := rpctimeout.NewContainer()

	onChangeCallback := func(data string, parser nacos.ConfigParser) {
		configs := map[string]*rpctimeout.RPCTimeout{}
		err := parser.Decode(param.Type, data, &configs)
		if err != nil {
			klog.Warnf("[nacos] %s client nacos rpc timeout: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}
		rpcTimeoutContainer.NotifyPolicyChange(configs)
	}

	nacosClient.RegisterConfigCallback(param, onChangeCallback, uniqueID)

	return rpcTimeoutContainer
}
