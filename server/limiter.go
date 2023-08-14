package server

import (
	"sync/atomic"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
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

	return server.WithLimit(initLimiteOptions(param, dest, nacosClient))
}

// LimiterConfig the limiter config
type LimiterConfig struct {
	ConnectionLimit int64 `json:"connection_limit"`
	QPSLimit        int64 `json:"qps_limit"`
}

// Valid checks if the config is valid.
func (lc *LimiterConfig) Valid() bool {
	return lc.ConnectionLimit > 0 || lc.QPSLimit > 0
}

// updaterWrapper can't guarantee the bootstrap order of the nacos and limiter, you
// should make sure the limit.Updater is initialized before the update.
type updaterWrapper struct {
	service string
	updater atomic.Value
	opt     limit.Option
}

// UpdateLimit update the limiter.
func (uw *updaterWrapper) UpdateLimit(lc *LimiterConfig) {
	if !lc.Valid() {
		klog.Warnf("[nacos] %s server nacos limiter config is invalid %v skip...", uw.service, uw.opt)
		return
	}
	uw.opt.MaxConnections, uw.opt.MaxQPS = int(lc.ConnectionLimit), int(lc.QPSLimit)
	updater := uw.updater.Load()
	if updater == nil {
		return
	}
	if u, ok := updater.(limit.Updater); ok {
		u.UpdateLimit(&uw.opt)
	}
}

func initLimiteOptions(param vo.ConfigParam, dest string, nacosClient nacos.Client) *limit.Option {
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
		lc := &LimiterConfig{}
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
