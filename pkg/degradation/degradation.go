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

package degradation

import (
	"context"
	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/cloudwego/kitex/pkg/acl"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
)

var errorDegradation = errors.New("rejected by client degradation config")

// DegradationConfig is policy config of degradator.
// DON'T FORGET to update DeepCopy() and Equals() if you add new fields.
type Config struct {
	Enable     bool `json:"enable"`
	Percentage int  `json:"percentage"`
}

type Container struct {
	sync.RWMutex
	dgConfig atomic.Value
}

var defaultDGConfig = &Config{Enable: false, Percentage: 0}

// GetDefaultDGConfig return defaultConfig of degradation.
func GetDefaultDGConfig() *Config {
	return defaultDGConfig
}

func NewDeGradationContainer() *Container {
	dgContainer := &Container{}
	dgContainer.dgConfig.Store(GetDefaultDGConfig())
	return dgContainer
}

func (s *Container) NotifyPolicyChange(cfg *Config) {
	s.dgConfig.Store(cfg)
}

func (s *Container) GetACLRule() acl.RejectFunc {
	return func(ctx context.Context, request interface{}) (reason error) {
		config := s.dgConfig.Load().(*Config)
		if !config.Enable {
			return nil
		}
		if fastrand.Intn(100) < config.Percentage {
			return errorDegradation
		}
		return nil
	}
}

// DeepCopy returns a full copy of DGConfig.
func (c *Config) DeepCopy() *Config {
	if c == nil {
		return nil
	}
	return &Config{
		Enable:     c.Enable,
		Percentage: c.Percentage,
	}
}

func (c *Config) Equals(other *Config) bool {
	if c == nil && other == nil {
		return true
	}
	if c == nil || other == nil {
		return false
	}
	return c.Enable == other.Enable && c.Percentage == other.Percentage
}
