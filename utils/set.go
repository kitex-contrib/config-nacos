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

import "sync"

// ThreadSafeSet wrapper of Set.
type ThreadSafeSet struct {
	sync.RWMutex
	s Set
}

// DiffAndEmplace returns the keys that are not in other and emplace the old set.
func (ts *ThreadSafeSet) DiffAndEmplace(other Set) []string {
	ts.Lock()
	defer ts.Unlock()
	out := ts.s.Diff(other)
	ts.s = other
	return out
}

// Set map template.
type Set map[string]bool

// Diff returns the keys that are not in other
func (s Set) Diff(other Set) []string {
	out := make([]string, 0, len(s))
	for key := range s {
		if _, ok := other[key]; !ok {
			out = append(out, key)
		}
	}
	return out
}
