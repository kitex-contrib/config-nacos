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
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"sigs.k8s.io/yaml"
)

const (
	NacosDefaultServerAddr   = "127.0.0.1"
	NacosDefaultPort         = 8848
	NacosDefaultGrpcPorc     = 9848
	NacosDefaultConfigGroup  = "DEFAULT_GROUP"
	NacosDefaultClientDataID = "{{.ClientServiceName}}.{{.ServerServiceName}}.{{.Category}}"
	NacosDefaultServerDataID = "{{.ServerServiceName}}.{{.Category}}"
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

var _ ConfigParser = &parser{}

// ConfigParser the parser for nacos config.
type ConfigParser interface {
	Decode(kind, data string, config interface{}) error
}

type parser struct{}

// Decode decodes the data to struct in specified format.
func (p *parser) Decode(kind, data string, config interface{}) error {
	switch kind {
	case "yaml", "json":
		// since YAML is a superset of JSON, it can parse JSON using a YAML parser
		return yaml.Unmarshal([]byte(data), config)
	default:
		return fmt.Errorf("unsupported config data type %s", kind)
	}
}

// DefaultConfigParse default nacos config parser.
func defaultConfigParse() ConfigParser {
	return &parser{}
}
