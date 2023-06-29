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
	"github.com/cloudwego/kitex/client"

	"github.com/kitex-contrib/config-nacos/nacos"
)

// NacosClientSuite nacos client config suite, configure retry timeout limit and circuitbreak dynamically from nacos.
type NacosClientSuite struct {
	nacosClient nacos.Client
	service     string
	fns         []nacos.CustomFunction
}

// NewSuite ...
func NewSuite(service string, cli nacos.Client,
	cfs ...nacos.CustomFunction,
) *NacosClientSuite {
	return &NacosClientSuite{
		service:     service,
		nacosClient: cli,
		fns:         cfs,
	}
}

// Options return a list client.Option
func (s *NacosClientSuite) Options() []client.Option {
	opts := make([]client.Option, 0, 8)
	opts = append(opts, WithRetryPolicy(s.service, s.nacosClient, s.fns...)...)
	return opts
}
