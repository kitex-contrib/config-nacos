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

package client

import (
	"sync"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/rpctimeout"
	"github.com/kitex-contrib/config-nacos/nacos"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// WithRPCTimeout sets the RPC timeout policy from nacos configuration center.
func WithRPCTimeout(dest, src string, nacosClient nacos.Client,
	cfs ...nacos.CustomFunction,
) []client.Option {
	param := nacos.NaocsConfigParam(&nacos.ConfigParamConfig{
		Category:          rpcTimeoutConfigName,
		ServerServiceName: dest,
		ClientServiceName: src,
	}, cfs...)

	return []client.Option{
		client.WithTimeoutProvider(initRPCTimeoutContainer(param, dest, nacosClient)),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			return nacosClient.DeregisterConfig(param)
		}),
	}
}

// rpcTimeoutContainer the implementation of timeout provider.
type rpcTimeoutContainer struct {
	// the key is method name, wildcard "*" can match anything.
	configs map[string]*rpctimeout.RPCTimeout
	sync.RWMutex
}

func newRPCTimeoutContainer() *rpcTimeoutContainer {
	return &rpcTimeoutContainer{
		configs: map[string]*rpctimeout.RPCTimeout{},
	}
}

func (rtc *rpcTimeoutContainer) notifyPolicyChange(configs map[string]*rpctimeout.RPCTimeout) {
	rtc.Lock()
	defer rtc.Unlock()
	rtc.configs = configs
}

// Timeouts return the rpc timeout config by the method name of rpc info.
func (rtc *rpcTimeoutContainer) Timeouts(ri rpcinfo.RPCInfo) rpcinfo.Timeouts {
	rtc.RLock()
	defer rtc.RUnlock()
	if config, ok := rtc.configs[ri.Invocation().MethodName()]; ok {
		klog.Debugf("[nacos] find the config %v for method %s", config, ri.Invocation().MethodName())
		return config
	}
	if config, ok := rtc.configs[wildcardMethod]; ok {
		klog.Debugf("[nacos] can't find method %s, use the wildcard method config %v", ri.Invocation().MethodName(), config)
		return config
	}
	klog.Debugf("[nacos] can't find config, return empty, configs %v", rtc.configs)
	return &rpctimeout.RPCTimeout{}
}

func initRPCTimeoutContainer(param vo.ConfigParam, dest string,
	nacosClient nacos.Client,
) rpcinfo.TimeoutProvider {
	rpcTimeoutContainer := newRPCTimeoutContainer()

	onChangeCallback := func(data string, parser nacos.ConfigParser) {
		configs := map[string]*rpctimeout.RPCTimeout{}
		err := parser.Decode(param.Type, data, &configs)
		if err != nil {
			klog.Warnf("[nacos] %s client nacos rpc timeout: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}
		rpcTimeoutContainer.notifyPolicyChange(configs)
	}

	nacosClient.RegisterConfigCallback(dest,
		rpcTimeoutConfigName,
		param,
		onChangeCallback,
	)

	return rpcTimeoutContainer
}
