// Copyright 2023 CloudWeGo Authors
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
	"fmt"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"sigs.k8s.io/yaml"
)

// NewDefaultNacosClient Create a default Nacos client
// It can create a client with default config by env variable.
// See: env.go
func NewDefaultNacosClient() (config_client.IConfigClient, error) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(NacosAddr(), uint64(NacosPort())),
	}
	cc := constant.ClientConfig{
		NamespaceId:         NacosNameSpaceId(),
		RegionId:            NACOS_DEFAULT_REGIONID,
		NotLoadCacheAtStart: true,
		CustomLogger:        NewCustomNacosLogger(),
	}
	return clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
}

// RegistryConfigUpdateCallback registry the callback function to nacos client.
func RegistryConfigUpdateCallback(dest, category string,
	nacosClient config_client.IConfigClient,
	param vo.ConfigParam,
	callback func(string),
) {
	param.OnChange = func(namespace, group, dataId, data string) {
		klog.Debugf("[nacos] %s client %s config updated, namespace %s group %s dataId %s data %s",
			dest, category, namespace, group, dataId, data)
		callback(data)
	}
	data, err := nacosClient.GetConfig(param)
	if err != nil && !IsNotExistError(err) {
		panic(err)
	}

	callback(data)

	err = nacosClient.ListenConfig(param)
	if err != nil {
		panic(err)
	}
}

// Unmarshal unmarshals the data to struct in specified format.
func Unmarshal(kind vo.ConfigType, data string, config interface{}) error {
	switch kind {
	case vo.YAML, vo.JSON:
		// support json and yaml since YAML is a superset of JSON, it can parse JSON using a YAML parser
		return yaml.Unmarshal([]byte(data), config)
	default:
		return fmt.Errorf("unsupport config data type %s", kind)
	}
}
