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

package nacos

import (
	"bytes"
	"os"
	"strconv"
	"text/template"

	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/cloudwego/kitex/pkg/klog"
)

const (
	NACOS_ENV_SERVER_ADDR      = "serverAddr"
	NACOS_ENV_PORT             = "serverPort"
	NACOS_ENV_NAMESPACE_ID     = "namespace"
	NACOS_ENV_CONFIG_GROUP     = "configGroup"
	NACOS_ENV_CONFIG_DATA_ID   = "configDataId"
	NACOS_DEFAULT_SERVER_ADDR  = "127.0.0.1"
	NACOS_DEFAULT_PORT         = 8848
	NACOS_DEFAULT_REGIONID     = "cn-hangzhou"
	NACOS_DEFAULT_CONFIG_GROUP = "DEFAULT_GROUP"
	NACOS_DEFAULT_DATA_ID      = "{{.ClientServiceName}}.{{.ServerServiceName}}.{{.Category}}"
)

const (
	defaultContext = "/nacos"
)

// CustomFunction use for customize the config param.
type CustomFunction func(*vo.ConfigParam)

// ConfigParamConfig use for render the dataId or group info by go template, ref: https://pkg.go.dev/text/template
// The fixed key shows as below.
type ConfigParamConfig struct {
	Category          string
	ClientServiceName string
	ServerServiceName string
}

func render(name, format string, cpc *ConfigParamConfig) string {
	t, err := template.New(name).Parse(format)
	if err != nil {
		panic(err)
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, cpc)
	if err != nil {
		panic(err)
	}
	return tpl.String()
}

// NaocsConfigParam Get nacos config from environment variables. All the parameters can be customised with CustomFunction.
// ConfigParam explain:
//  1. Type: data format, support JSON YMAL, JSON by default. Could extend it by implementing the ConfigParser interface.
//  2. Context: empty by default it use CustomFunction.
//  3. Group: DEFAULT_GROUP by default.
//  4. DataId: {{.ClientServiceName}}.{{.ServerServiceName}}.{{.Category}} by default. Customize it by CustomFunction or
//     use specified format. ref: nacos/env.go:46
func NaocsConfigParam(cpc *ConfigParamConfig, cfs ...CustomFunction) vo.ConfigParam {
	param := vo.ConfigParam{
		DataId:  render("dataId", NacosConfigDataId(), cpc),
		Group:   render("group", NacosConfigGroup(), cpc),
		Type:    vo.JSON,
		Content: defaultContext,
	}
	for _, cf := range cfs {
		cf(&param)
	}
	return param
}

// NacosConfigDataId Get nacos DataId from environment variables
func NacosConfigDataId() string {
	dataId := os.Getenv(NACOS_ENV_CONFIG_DATA_ID)
	if len(dataId) == 0 {
		return NACOS_DEFAULT_DATA_ID
	}
	return dataId
}

// NacosConfigGroup Get nacos config group from environment variables
func NacosConfigGroup() string {
	configGroup := os.Getenv(NACOS_ENV_CONFIG_GROUP)
	if len(configGroup) == 0 {
		return NACOS_DEFAULT_CONFIG_GROUP
	}
	return configGroup
}

// NacosPort Get Nacos port from environment variables
func NacosPort() int64 {
	portText := os.Getenv(NACOS_ENV_PORT)
	if len(portText) == 0 {
		return NACOS_DEFAULT_PORT
	}
	port, err := strconv.ParseInt(portText, 10, 64)
	if err != nil {
		klog.Errorf("ParseInt failed,err:%s", err.Error())
		return NACOS_DEFAULT_PORT
	}
	return port
}

// NacosAddr Get Nacos addr from environment variables
func NacosAddr() string {
	addr := os.Getenv(NACOS_ENV_SERVER_ADDR)
	if len(addr) == 0 {
		return NACOS_DEFAULT_SERVER_ADDR
	}
	return addr
}

// NacosNameSpaceId Get Nacos namespace id from environment variables
func NacosNameSpaceId() string {
	return os.Getenv(NACOS_ENV_NAMESPACE_ID)
}
