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
	"github.com/cloudwego-contrib/cwgo-pkg/config/nacos/v2/pkg/degradation"
)

// DegradationConfig is policy config of degradator.
// DON'T FORGET to update DeepCopy() and Equals() if you add new fields.
type Config = degradation.Config

type Container = degradation.Container

// GetDefaultDegradationConfig return defaultConfig of degradation.
func GetDefaultDegradationConfig() *Config {
	return degradation.GetDefaultDegradationConfig()
}

func NewDegradationContainer() *Container {
	return degradation.NewDegradationContainer()
}
