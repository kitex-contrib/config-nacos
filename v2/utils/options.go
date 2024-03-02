package utils

import (
	"github.com/kitex-contrib/config-nacos/v2/nacos"
)

// Option is used to custom Options.
type Option interface {
	Apply(*Options)
}

// Options is used to initialize the nacos config suit or option.
type Options struct {
	NacosCustomFunctions []nacos.CustomFunction
}
