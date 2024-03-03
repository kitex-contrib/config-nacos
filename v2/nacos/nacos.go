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
	"bytes"
	"sync"
	"text/template"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// callbackHandler ...
type callbackHandler func(namespace, group, dataId, data string)

type configParam struct {
	DataID string
	Group  string
}

// NOTE: the nacos client use namespace + dataID + group as cache key, and the namespace
// in client is fixed.
func configParamKey(in vo.ConfigParam) configParam {
	return configParam{
		DataID: in.DataId,
		Group:  in.Group,
	}
}

// Client the wrapper of nacos client.
type Client interface {
	SetParser(ConfigParser)
	ClientConfigParam(cpc *ConfigParamConfig) (vo.ConfigParam, error)
	ServerConfigParam(cpc *ConfigParamConfig) (vo.ConfigParam, error)
	RegisterConfigCallback(vo.ConfigParam, func(string, ConfigParser), int64)
	DeregisterConfig(vo.ConfigParam, int64) error
}

type client struct {
	ncli config_client.IConfigClient
	// support customise parser
	parser               ConfigParser
	groupTemplate        *template.Template
	serverDataIDTemplate *template.Template
	clientDataIDTemplate *template.Template

	handlerMutex sync.RWMutex
	handlers     map[configParam]map[int64]callbackHandler
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
	Password           string
	Username           string
	ConfigParser       ConfigParser
	GrpcPort           uint64
}

// NewClient Create a default Nacos client
func NewClient(opts Options) (Client, error) {
	if opts.Address == "" {
		opts.Address = NacosAddr()
	}
	if opts.Port == 0 {
		opts.Port = NacosPort()
	}
	if opts.NamespaceID == "" {
		opts.NamespaceID = NacosNameSpaceId()
	}
	if opts.Group == "" {
		opts.Group = NacosDefaultConfigGroup
	}
	if opts.ConfigParser == nil {
		opts.ConfigParser = defaultConfigParse()
	}
	if opts.ServerDataIDFormat == "" {
		opts.ServerDataIDFormat = NacosDefaultServerDataID
	}
	if opts.ClientDataIDFormat == "" {
		opts.ClientDataIDFormat = NacosDefaultClientDataID
	}
	if opts.GrpcPort == 0 {
		opts.GrpcPort = NacosDefaultGrpcPorc
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(opts.Address, opts.Port),
	}

	cc := constant.ClientConfig{
		NamespaceId:         opts.NamespaceID,
		RegionId:            opts.RegionID,
		NotLoadCacheAtStart: true,
		Password:            opts.Password,
		Username:            opts.Username,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
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
		handlers:             map[configParam]map[int64]callbackHandler{},
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
		Type:    "json",
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
func (c *client) DeregisterConfig(cfg vo.ConfigParam, uniqueID int64) error {
	key := configParamKey(cfg)
	klog.Debugf("deregister key %v for uniqueID %d", key, uniqueID)
	c.handlerMutex.Lock()
	defer c.handlerMutex.Unlock()
	handlers, ok := c.handlers[key]
	if ok {
		delete(handlers, uniqueID)
	}
	if len(handlers) == 0 {
		klog.Debugf("the handlers for key %v is empty, cancel listen config from nacos", key)
		return c.ncli.CancelListenConfig(cfg)
	}
	return nil
}

func (c *client) onChange(namespace, group, dataId, data string) {
	handlers := make([]callbackHandler, 0, 5)
	c.handlerMutex.RLock()
	key := configParam{
		DataID: dataId,
		Group:  group,
	}
	for _, handler := range c.handlers[key] {
		handlers = append(handlers, handler)
	}
	c.handlerMutex.RUnlock()

	for _, handler := range handlers {
		handler(namespace, group, dataId, data)
	}
}

func (c *client) listenConfig(param vo.ConfigParam, uniqueID int64) {
	key := configParamKey(param)
	klog.Debugf("register key %v for uniqueID %d", key, uniqueID)
	c.handlerMutex.Lock()
	handlers, ok := c.handlers[key]
	if !ok {
		handlers = map[int64]callbackHandler{}
		c.handlers[key] = handlers
	}
	handlers[uniqueID] = param.OnChange
	c.handlerMutex.Unlock()

	if !ok {
		klog.Debugf("the first time %v register, listen config from nacos", key)
		err := c.ncli.ListenConfig(vo.ConfigParam{
			DataId:   param.DataId,
			Group:    param.Group,
			Content:  param.Content,
			Type:     param.Type,
			OnChange: c.onChange,
		})
		// Performs only local connection and fails only when the input params are invalid
		if err != nil {
			panic(err)
		}
	}
}

// RegisterConfigCallback register the callback function to nacos client.
func (c *client) RegisterConfigCallback(param vo.ConfigParam,
	callback func(string, ConfigParser), uniqueID int64,
) {
	param.OnChange = func(namespace, group, dataId, data string) {
		klog.Debugf("[nacos] uniqueID %d config %s updated, namespace %s group %s dataId %s data %s",
			uniqueID, param.DataId, namespace, group, dataId, data)
		callback(data, c.parser)
	}

	// NOTE: does not ensure that GetConfig succeeds, the govern policy may not be correct if it fails here.
	data, err := c.ncli.GetConfig(param)
	if err != nil {
		// If the initial connection fails and the reconnection is successful, the callback handler can also be invoked.
		// Ignore the error here and print the error info.
		klog.Warnf("get config %v from nacos failed %v", param, err)
	}

	callback(data, c.parser)

	c.listenConfig(param, uniqueID)
}
