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
	"errors"
	"testing"

	"github.com/cloudwego/kitex/pkg/acl"
	"github.com/cloudwego/thriftgo/pkg/test"
)

var errFake = errors.New("fake error")

func invoke(ctx context.Context, request, response interface{}) error {
	return errFake
}

func TestNewContainer(t *testing.T) {
	container := NewDegradationContainer()
	aclMiddleware := acl.NewACLMiddleware([]acl.RejectFunc{container.GetACLRule()})
	test.Assert(t, errors.Is(aclMiddleware(invoke)(context.Background(), nil, nil), errFake))
	container.NotifyPolicyChange(&Config{Enable: false, Percentage: 100})
	test.Assert(t, errors.Is(aclMiddleware(invoke)(context.Background(), nil, nil), errFake))
	container.NotifyPolicyChange(&Config{Enable: true, Percentage: 100})
	test.Assert(t, errors.Is(aclMiddleware(invoke)(context.Background(), nil, nil), errorDegradation))
}
