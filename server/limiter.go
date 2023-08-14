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
	"sync/atomic"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/limiter"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/config-nacos/nacos"
	"github.com/nacos-group/nacos-sdk-go/vo"
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

// updaterWrapper the wrapper maintains the configuration and limiter updater.
type updaterWrapper struct {
	service string
	updater atomic.Value
	opt     limit.Option
}

// UpdateLimit update the limiter.
func (uw *updaterWrapper) UpdateLimit(lc *limiter.LimiterConfig) {
	uw.opt.MaxConnections, uw.opt.MaxQPS = int(lc.ConnectionLimit), int(lc.QPSLimit)

	if !uw.opt.Valid() {
		klog.Warnf("[nacos] %s server nacos limiter config is invalid %v skip...", uw.service, uw.opt)
		return
	}

	// can't guarantee the bootstrap order of the nacos and limiter, you
	// should make sure the limit.Updater is initialized before the update.
	updater := uw.updater.Load()
	if updater == nil {
		return
	}
	if u, ok := updater.(limit.Updater); ok {
		u.UpdateLimit(&uw.opt)
	}
}

func initLimitOptions(param vo.ConfigParam, dest string, nacosClient nacos.Client) *limit.Option {
	uw := updaterWrapper{
		service: dest,
	}
	uw.opt.UpdateControl = func(u limit.Updater) {
		uw.updater.Store(u)
		if uw.opt.Valid() {
			u.UpdateLimit(&uw.opt)
		}
	}

	onChangeCallback := func(data string, parser nacos.ConfigParser) {
		lc := &limiter.LimiterConfig{}
		err := parser.Decode(param.Type, data, lc)
		if err != nil {
			klog.Warnf("[nacos] %s server nacos limiter config: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}
		uw.UpdateLimit(lc)
	}

	nacosClient.RegisterConfigCallback(param, onChangeCallback)

	return &uw.opt
}
