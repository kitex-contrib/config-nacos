// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/cloudwego/kitex/pkg/retry"
	"github.com/kitex-contrib/config-nacos/v2/nacos"
	"github.com/kitex-contrib/config-nacos/v2/utils"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// WithRetryPolicy sets the retry policy from nacos configuration center.
func WithRetryPolicy(dest, src string, nacosClient nacos.Client, opts utils.Options) []client.Option {
	param, err := nacosClient.ClientConfigParam(&nacos.ConfigParamConfig{
		Category:          retryConfigName,
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

	rc := initRetryContainer(param, dest, nacosClient, uniqueID)
	return []client.Option{
		client.WithRetryContainer(rc),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			err := nacosClient.DeregisterConfig(param, uniqueID)
			if err != nil {
				return err
			}
			return rc.Close()
		}),
	}
}

func initRetryContainer(param vo.ConfigParam, dest string,
	nacosClient nacos.Client, uniqueID int64,
) *retry.Container {
	retryContainer := retry.NewRetryContainerWithPercentageLimit()

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

	nacosClient.RegisterConfigCallback(param, onChangeCallback, uniqueID)

	return retryContainer
}
