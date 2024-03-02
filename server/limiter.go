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

package server

import (
	"sync/atomic"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/limiter"
	"github.com/cloudwego/kitex/server"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/kitex-contrib/config-nacos/nacos"
	"github.com/kitex-contrib/config-nacos/utils"
)

// WithLimiter sets the limiter config from nacos configuration center.
func WithLimiter(dest string, nacosClient nacos.Client, opts utils.Options) server.Option {
	param, err := nacosClient.ServerConfigParam(&nacos.ConfigParamConfig{
		Category:          limiterConfigName,
		ServerServiceName: dest,
	})
	if err != nil {
		panic(err)
	}

	for _, f := range opts.NacosCustomFunctions {
		f(&param)
	}
	uniqueID := nacos.GetUniqueID()
	server.RegisterShutdownHook(func() {
		nacosClient.DeregisterConfig(param, uniqueID)
	})
	return server.WithLimit(initLimitOptions(param, dest, nacosClient, uniqueID))
}

func initLimitOptions(param vo.ConfigParam, dest string, nacosClient nacos.Client, uniqueID int64) *limit.Option {
	var updater atomic.Value
	opt := &limit.Option{}
	opt.UpdateControl = func(u limit.Updater) {
		klog.Debugf("[nacos] %s server nacos limiter updater init, config %v", dest, *opt)
		u.UpdateLimit(opt)
		updater.Store(u)
	}
	onChangeCallback := func(data string, parser nacos.ConfigParser) {
		lc := &limiter.LimiterConfig{}
		err := parser.Decode(param.Type, data, lc)
		if err != nil {
			klog.Warnf("[nacos] %s server nacos limiter config: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}
		opt.MaxConnections = int(lc.ConnectionLimit)
		opt.MaxQPS = int(lc.QPSLimit)
		u := updater.Load()
		if u == nil {
			klog.Warnf("[nacos] %s server nacos limiter config failed as the updater is empty", dest)
			return
		}
		if !u.(limit.Updater).UpdateLimit(opt) {
			klog.Warnf("[nacos] %s server nacos limiter config: data %s may do not take affect", dest, data)
		}
	}

	nacosClient.RegisterConfigCallback(param, onChangeCallback, uniqueID)
	return opt
}
