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
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnvFunc test env func
func TestEnvFunc(t *testing.T) {
	assert.Equal(t, int64(8848), nacosPort())
	assert.Equal(t, "127.0.0.1", nacosAddr())
	assert.Equal(t, "", nacosNameSpaceId())

	t.Setenv(NACOS_ENV_NAMESPACE_ID, "ns")
	t.Setenv(NACOS_ENV_SERVER_ADDR, "1.1.1.1")
	t.Setenv(NACOS_ENV_PORT, "80")
	t.Setenv(NACOS_ENV_CONFIG_DATA_ID, "{{.ClientServiceName}}")
	t.Setenv(NACOS_ENV_CONFIG_GROUP, "{{.Category}}")

	assert.Equal(t, int64(80), nacosPort())
	assert.Equal(t, "1.1.1.1", nacosAddr())
	assert.Equal(t, "ns", nacosNameSpaceId())
}
