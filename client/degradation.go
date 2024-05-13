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
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/kitex-contrib/config-nacos/nacos"
	"github.com/kitex-contrib/config-nacos/utils"
)

// WithDegradation sets the degradation policy from nacos configuration center.
func WithDegradation(dest, src string, nacosClient nacos.Client, opts utils.Options) []client.Option {
	param, err := nacosClient.ClientConfigParam(&nacos.ConfigParamConfig{
		Category:          degradationName,
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

	dgContainer := initDegradation(param, dest, src, nacosClient, uniqueID)

	return []client.Option{
		client.WithACLRules(dgContainer.GetACLRule),
		client.WithCloseCallbacks(func() error {
			err := nacosClient.DeregisterConfig(param, uniqueID)
			if err != nil {
				return err
			}
			// cancel the configuration listener when client is closed.
			return nil
		}),
	}
}

func initDegradation(param vo.ConfigParam, dest, src string,
	nacosClient nacos.Client, uniqueID int64,
) *degradation.Container {
	dgContainer := degradation.NewDeGradationContainer()

	onChangeCallback := func(data string, parser nacos.ConfigParser) {
		config := degradation.Config{}
		err := parser.Decode(param.Type, data, &config)
		if err != nil {
			klog.Warnf("[nacos] %s client nacos rpc degradation: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}
		// update degradation config & rule
		dgContainer.NotifyPolicyChange(config)
	}

	nacosClient.RegisterConfigCallback(param, onChangeCallback, uniqueID)

	return dgContainer
}
