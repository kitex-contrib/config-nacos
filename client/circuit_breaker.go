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
	"strings"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/circuitbreak"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/kitex-contrib/config-nacos/nacos"
	"github.com/kitex-contrib/config-nacos/utils"
)

// WithCircuitBreaker sets the circuit breaker policy from nacos configuration center.
func WithCircuitBreaker(dest, src string, nacosClient nacos.Client,
	cfs ...nacos.CustomFunction,
) []client.Option {
	param := nacos.NacosConfigParam(&nacos.ConfigParamConfig{
		Category:          circuitBreakerConfigName,
		ServerServiceName: dest,
		ClientServiceName: src,
	}, cfs...)

	cbSuite := initCircuitBreaker(param, dest, src, nacosClient)

	return []client.Option{
		// the client identity is necessary when generate the key for service circuit breaker.
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: src,
		}),
		client.WithCircuitBreaker(cbSuite),
		client.WithCloseCallbacks(func() error {
			err := cbSuite.Close()
			if err != nil {
				return err
			}
			// cancel the configuration listener when client is closed.
			return nacosClient.DeregisterConfig(param)
		}),
	}
}

// Be consistent with the RPCInfo2Key function in kitex/pkg/circuitbreak/cbsuite.go.
// NOTE The fromService should keep consistent with the ServiceName field in rpcinfo.EndpointBasicInfo
// which can be set using client.WithClientBasicInfo function, should
// set it correctly.
func genServiceCBKey(fromService, toService, method string) string {
	sum := len(fromService) + len(toService) + len(method) + 2
	var buf strings.Builder
	buf.Grow(sum)
	buf.WriteString(fromService)
	buf.WriteByte('/')
	buf.WriteString(toService)
	buf.WriteByte('/')
	buf.WriteString(method)
	return buf.String()
}

func initCircuitBreaker(param vo.ConfigParam, dest, src string,
	nacosClient nacos.Client,
) *circuitbreak.CBSuite {
	cb := circuitbreak.NewCBSuite(nil)
	lcb := utils.ThreadSafeSet{}

	onChangeCallback := func(data string, parser nacos.ConfigParser) {
		set := utils.Set{}
		configs := map[string]circuitbreak.CBConfig{}
		err := parser.Decode(param.Type, data, &configs)
		if err != nil {
			klog.Warnf("[nacos] %s client nacos rpc timeout: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}

		for method, config := range configs {
			set[method] = true
			key := genServiceCBKey(src, dest, method)
			cb.UpdateServiceCBConfig(key, config)
		}

		for _, method := range lcb.DiffAndEmplace(set) {
			key := genServiceCBKey(src, dest, method)
			cb.UpdateServiceCBConfig(key, circuitbreak.GetDefaultCBConfig())
		}
	}

	nacosClient.RegisterConfigCallback(param, onChangeCallback)

	return cb
}
