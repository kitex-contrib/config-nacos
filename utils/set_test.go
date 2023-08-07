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

package utils

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	ts := &ThreadSafeSet{}
	m1 := Set{
		"h1": true,
		"h2": true,
	}
	got := ts.DiffAndEmplace(m1)
	assert.Equal(t, []string([]string{}), got)
	assert.Equal(t, m1, ts.s)

	m2 := Set{
		"h3": true,
		"h4": true,
	}
	got = ts.DiffAndEmplace(m2)
	sort.Strings(got)
	assert.Equal(t, []string([]string{"h1", "h2"}), got)
	assert.Equal(t, m2, ts.s)
}
