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
	"github.com/kitex-contrib/config-nacos/nacos"
	"github.com/kitex-contrib/config-nacos/utils"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/retry"
)

// WithRetryPolicy sets the retry policy from nacos configuration center.
func WithRetryPolicy(dest, src string, nacosClient nacos.Client,
	cfs ...nacos.CustomFunction,
) []client.Option {
	param, err := nacosClient.ClientConfigParam(&nacos.ConfigParamConfig{
		Category:          retryConfigName,
		ServerServiceName: dest,
		ClientServiceName: src,
	}, cfs...)
	if err != nil {
		panic(err)
	}

	return []client.Option{
		client.WithRetryContainer(initRetryContainer(param, dest, nacosClient)),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			return nacosClient.DeregisterConfig(param)
		}),
	}
}

func initRetryContainer(param vo.ConfigParam, dest string,
	nacosClient nacos.Client,
) *retry.Container {
	retryContainer := retry.NewRetryContainer()

	ts := utils.ThreadSafeSet{}

	onChangeCallback := func(data string, parser nacos.ConfigParser) {
		// the key is method name, wildcard "*" can match anything.
		rcs := map[string]*retry.Policy{}
		err := parser.Decode(param.Type, data, &rcs)
		if err != nil {
			klog.Warnf("[nacos] %s client nacos retry: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}

		set := utils.Set{}
		for method, policy := range rcs {
			set[method] = true
			if policy.BackupPolicy != nil && policy.FailurePolicy != nil {
				klog.Warnf("[nacos] %s client policy for method %s BackupPolicy and FailurePolicy must not be set at same time",
					dest, method)
				continue
			}
			if policy.BackupPolicy == nil && policy.FailurePolicy == nil {
				klog.Warnf("[nacos] %s client policy for method %s BackupPolicy and FailurePolicy must not be empty at same time",
					dest, method)
				continue
			}
			retryContainer.NotifyPolicyChange(method, *policy)
		}

		for _, method := range ts.DiffAndEmplace(set) {
			retryContainer.DeletePolicy(method)
		}
	}

	nacosClient.RegisterConfigCallback(param, onChangeCallback)

	return retryContainer
}
