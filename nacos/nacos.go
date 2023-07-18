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
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// Client the wrapper of nacos client.
type Client interface {
	SetParser(ConfigParser)
	RegisterConfigCallback(string, string, vo.ConfigParam, func(string, ConfigParser))
	DeregisterConfig(vo.ConfigParam) error
}

type client struct {
	ncli config_client.IConfigClient
	// support customise parser
	parser ConfigParser
}

// DefaultClient Create a default Nacos client
// It can create a client with default config by env variable.
// See: env.go
func DefaultClient() (Client, error) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(NacosAddr(), uint64(NacosPort())),
	}
	cc := constant.ClientConfig{
		NamespaceId:         NacosNameSpaceId(),
		RegionId:            NACOS_DEFAULT_REGIONID,
		NotLoadCacheAtStart: true,
		CustomLogger:        NewCustomNacosLogger(),
	}
	nacosClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return nil, err
	}
	c := &client{
		ncli:   nacosClient,
		parser: defaultConfigParse(),
	}
	return c, nil
}

// SetParser support customise parser
func (c *client) SetParser(parser ConfigParser) {
	c.parser = parser
}

// DeregisterConfig deregister the config.
func (c *client) DeregisterConfig(cfg vo.ConfigParam) error {
	return c.ncli.CancelListenConfig(cfg)
}

// RegisterConfigCallback register the callback function to nacos client.
func (c *client) RegisterConfigCallback(dest, category string,
	param vo.ConfigParam,
	callback func(string, ConfigParser),
) {
	param.OnChange = func(namespace, group, dataId, data string) {
		klog.Debugf("[nacos] %s client %s config updated, namespace %s group %s dataId %s data %s",
			dest, category, namespace, group, dataId, data)
		callback(data, c.parser)
	}
	data, err := c.ncli.GetConfig(param)
	// the nacos client has handled the not exist error.
	if err != nil {
		panic(err)
	}

	callback(data, c.parser)

	err = c.ncli.ListenConfig(param)
	if err != nil {
		panic(err)
	}
}
