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
	"bytes"
	"text/template"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/logger"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// Client the wrapper of nacos client.
type Client interface {
	SetParser(ConfigParser)
	ClientConfigParam(cpc *ConfigParamConfig) (vo.ConfigParam, error)
	ServerConfigParam(cpc *ConfigParamConfig) (vo.ConfigParam, error)
	RegisterConfigCallback(vo.ConfigParam, func(string, ConfigParser))
	DeregisterConfig(vo.ConfigParam) error
}

type client struct {
	ncli config_client.IConfigClient
	// support customise parser
	parser               ConfigParser
	groupTemplate        *template.Template
	serverDataIDTemplate *template.Template
	clientDataIDTemplate *template.Template
}

// Options nacos config options. All the fields have default value.
type Options struct {
	Address            string
	Port               uint64
	NamespaceID        string
	RegionID           string
	Group              string
	ServerDataIDFormat string
	ClientDataIDFormat string
	CustomLogger       logger.Logger
	ConfigParser       ConfigParser
}

// NewClient Create a default Nacos client
func NewClient(opts Options) (Client, error) {
	if opts.Address == "" {
		opts.Address = NacosDefaultServerAddr
	}
	if opts.Port == 0 {
		opts.Port = NacosDefaultPort
	}
	if opts.CustomLogger == nil {
		opts.CustomLogger = NewCustomNacosLogger()
	}
	if opts.ConfigParser == nil {
		opts.ConfigParser = defaultConfigParse()
	}
	if opts.Group == "" {
		opts.Group = NacosDefaultConfigGroup
	}
	if opts.ServerDataIDFormat == "" {
		opts.ServerDataIDFormat = NacosDefaultServerDataID
	}
	if opts.ClientDataIDFormat == "" {
		opts.ClientDataIDFormat = NacosDefaultClientDataID
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(opts.Address, opts.Port),
	}
	cc := constant.ClientConfig{
		NamespaceId:         opts.NamespaceID,
		RegionId:            opts.RegionID,
		NotLoadCacheAtStart: true,
		CustomLogger:        opts.CustomLogger,
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
	groupTemplate, err := template.New("group").Parse(opts.Group)
	if err != nil {
		return nil, err
	}
	serverDataIDTemplate, err := template.New("serverDataID").Parse(opts.ServerDataIDFormat)
	if err != nil {
		return nil, err
	}
	clientDataIDTemplate, err := template.New("clientDataID").Parse(opts.ClientDataIDFormat)
	if err != nil {
		return nil, err
	}
	c := &client{
		ncli:                 nacosClient,
		parser:               opts.ConfigParser,
		groupTemplate:        groupTemplate,
		serverDataIDTemplate: serverDataIDTemplate,
		clientDataIDTemplate: clientDataIDTemplate,
	}
	return c, nil
}

// SetParser support customise parser
func (c *client) SetParser(parser ConfigParser) {
	c.parser = parser
}

func (c *client) render(cpc *ConfigParamConfig, t *template.Template) (string, error) {
	var tpl bytes.Buffer
	err := t.Execute(&tpl, cpc)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// ServerConfigParam render server config parameters
func (c *client) ServerConfigParam(cpc *ConfigParamConfig) (vo.ConfigParam, error) {
	return c.configParam(cpc, c.serverDataIDTemplate)
}

// ClientConfigParam render client config parameters
func (c *client) ClientConfigParam(cpc *ConfigParamConfig) (vo.ConfigParam, error) {
	return c.configParam(cpc, c.clientDataIDTemplate)
}

// configParam render config parameters. All the parameters can be customized with CustomFunction.
// ConfigParam explain:
//  1. Type: data id format, support JSON and YAML, JSON by default. Could extend it by implementing the ConfigParser interface.
//  2. Content: empty by default. Customize with CustomFunction.
//  3. Group: DEFAULT_GROUP by default.
//  4. ServerDataId: {{.ServerServiceName}}.{{.Category}} by default.
//     ClientDataId: {{.ClientServiceName}}.{{.ServerServiceName}}.{{.Category}} by default.
func (c *client) configParam(cpc *ConfigParamConfig, t *template.Template) (vo.ConfigParam, error) {
	param := vo.ConfigParam{
		Type:    vo.JSON,
		Content: defaultContent,
	}
	var err error
	param.DataId, err = c.render(cpc, t)
	if err != nil {
		return param, err
	}
	param.Group, err = c.render(cpc, c.groupTemplate)
	if err != nil {
		return param, err
	}
	return param, nil
}

// DeregisterConfig deregister the config.
func (c *client) DeregisterConfig(cfg vo.ConfigParam) error {
	return c.ncli.CancelListenConfig(cfg)
}

// RegisterConfigCallback register the callback function to nacos client.
func (c *client) RegisterConfigCallback(param vo.ConfigParam,
	callback func(string, ConfigParser),
) {
	param.OnChange = func(namespace, group, dataId, data string) {
		klog.Debugf("[nacos] config %s updated, namespace %s group %s dataId %s data %s",
			param.DataId, namespace, group, dataId, data)
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
