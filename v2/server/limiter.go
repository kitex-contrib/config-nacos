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
	"github.com/kitex-contrib/config-nacos/v2/nacos"
	"github.com/kitex-contrib/config-nacos/v2/utils"

	"github.com/cloudwego/kitex/server"

	configserver "github.com/cloudwego-contrib/cwgo-pkg/config/nacos/v2/server"
)

// WithLimiter sets the limiter config from nacos configuration center.
func WithLimiter(dest string, nacosClient nacos.Client, opts utils.Options) server.Option {
	return configserver.WithLimiter(dest, nacosClient, opts)
}
