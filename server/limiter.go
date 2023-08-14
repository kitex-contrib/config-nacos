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

package server

import (
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/limiter"
	"github.com/cloudwego/kitex/server"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/kitex-contrib/config-nacos/nacos"
)

// WithLimiter sets the limiter config from nacos configuration center.
func WithLimiter(dest string, nacosClient nacos.Client,
	cfs ...nacos.CustomFunction,
) server.Option {
	param := nacos.NacosConfigParam(&nacos.ConfigParamConfig{
		Category:          limiterConfigName,
		ServerServiceName: dest,
	}, cfs...)

	return server.WithLimit(initLimitOptions(param, dest, nacosClient))
}

func initLimitOptions(param vo.ConfigParam, dest string, nacosClient nacos.Client) *limit.Option {
	opt := limit.DefaultOption()
	onChangeCallback := func(data string, parser nacos.ConfigParser) {
		lc := &limiter.LimiterConfig{}
		err := parser.Decode(param.Type, data, lc)
		if err != nil {
			klog.Warnf("[nacos] %s server nacos limiter config: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}
		if !opt.UpdateLimitConfig(int(lc.ConnectionLimit), int(lc.QPSLimit)) {
			klog.Warnf("[nacos] %s server nacos limiter config: data %s may do not take affect", dest, data, err)
		}
	}

	nacosClient.RegisterConfigCallback(param, onChangeCallback)
	return opt
}
