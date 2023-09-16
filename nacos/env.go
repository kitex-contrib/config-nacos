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
	"os"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/cloudwego/kitex/pkg/klog"
)

const (
	NACOS_ENV_SERVER_ADDR      = "KITEX_CONFIG_NACOS_SERVER_ADDR"
	NACOS_ENV_PORT             = "KITEX_CONFIG_NACOS_SERVER_PORT"
	NACOS_ENV_NAMESPACE_ID     = "KITEX_CONFIG_NACOS_NAMESPACE"
	NACOS_ENV_CONFIG_GROUP     = "KITEX_CONFIG_NACOS_GROUP"
	NACOS_ENV_CONFIG_DATA_ID   = "KITEX_CONFIG_NACOS_DATA_ID"
	NACOS_DEFAULT_SERVER_ADDR  = "127.0.0.1"
	NACOS_DEFAULT_PORT         = 8848
	NACOS_DEFAULT_CONFIG_GROUP = "DEFAULT_GROUP"
	NACOS_DEFAULT_DATA_ID      = "{{.ClientServiceName}}.{{.ServerServiceName}}.{{.Category}}"
)

const (
	defaultContent = ""
)

// CustomFunction use for customize the config parameters.
type CustomFunction func(*vo.ConfigParam)

// ConfigParamConfig use for render the dataId or group info by go template, ref: https://pkg.go.dev/text/template
// The fixed key shows as below.
type ConfigParamConfig struct {
	Category          string
	ClientServiceName string
	ServerServiceName string
}

// nacosConfigDataId Get nacos DataId from environment variables
func nacosConfigDataId() string {
	dataId := os.Getenv(NACOS_ENV_CONFIG_DATA_ID)
	if len(dataId) == 0 {
		return NACOS_DEFAULT_DATA_ID
	}
	return dataId
}

// nacosConfigGroup Get nacos config group from environment variables
func nacosConfigGroup() string {
	configGroup := os.Getenv(NACOS_ENV_CONFIG_GROUP)
	if len(configGroup) == 0 {
		return NACOS_DEFAULT_CONFIG_GROUP
	}
	return configGroup
}

// nacosPort Get Nacos port from environment variables
func nacosPort() int64 {
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

// nacosAddr Get Nacos addr from environment variables
func nacosAddr() string {
	addr := os.Getenv(NACOS_ENV_SERVER_ADDR)
	if len(addr) == 0 {
		return NACOS_DEFAULT_SERVER_ADDR
	}
	return addr
}

// nacosNameSpaceId Get Nacos namespace id from environment variables
func nacosNameSpaceId() string {
	return os.Getenv(NACOS_ENV_NAMESPACE_ID)
}
