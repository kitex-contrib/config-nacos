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

package nacos

import (
	"github.com/cloudwego-contrib/cwgo-pkg/config/nacos/nacos"
)

// keep consistent with the env with alicloud
const (
	NacosAliServerAddrEnv = nacos.NacosAliNamespaceEnv
	NacosAliPortEnv       = nacos.NacosAliPortEnv
	NacosAliNamespaceEnv  = nacos.NacosAliNamespaceEnv
)

// NacosPort Get Nacos port from environment variables
func NacosPort() uint64 {
	return nacos.NacosPort()
}

// NacosAddr Get Nacos addr from environment variables
func NacosAddr() string {
	return nacos.NacosAddr()
}

// NacosNameSpaceId Get Nacos namespace id from environment variables
func NacosNameSpaceId() string {
	return nacos.NacosNameSpaceId()
}
