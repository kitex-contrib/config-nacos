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
	"sync"
	"testing"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
)

type fakeNacos struct {
	sync.RWMutex
	handlers map[configParam]callbackHandler
}

func (fn *fakeNacos) CloseClient() {
}

func (fn *fakeNacos) GetConfig(param vo.ConfigParam) (string, error) {
	return "", nil
}

func (fn *fakeNacos) PublishConfig(param vo.ConfigParam) (bool, error) {
	return false, nil
}

func (fn *fakeNacos) DeleteConfig(param vo.ConfigParam) (bool, error) {
	return false, nil
}

func (fn *fakeNacos) ListenConfig(params vo.ConfigParam) (err error) {
	fn.Lock()
	defer fn.Unlock()
	fn.handlers[configParamKey(params)] = params.OnChange
	return nil
}

func (fn *fakeNacos) CancelListenConfig(params vo.ConfigParam) (err error) {
	fn.Lock()
	defer fn.Unlock()
	delete(fn.handlers, configParamKey(params))
	return nil
}

func (fn *fakeNacos) SearchConfig(param vo.SearchConfigParam) (*model.ConfigPage, error) {
	return nil, nil
}

func (fn *fakeNacos) PublishAggr(param vo.ConfigParam) (published bool, err error) {
	return false, nil
}

func (fn *fakeNacos) change(cfg configParam, data string) {
	fn.Lock()
	defer fn.Unlock()

	handler, ok := fn.handlers[cfg]
	if !ok {
		return
	}
	fmt.Println("find handers ", fn.handlers, "cfg ", cfg, " data", data)
	handler("", cfg.Group, cfg.DataID, data)
}

func TestRegisterAndDeRegister(t *testing.T) {
	fake := &fakeNacos{
		handlers: map[configParam]callbackHandler{},
	}
	c := &client{
		ncli:     fake,
		handlers: map[configParam]map[int64]callbackHandler{},
	}

	var gotlock sync.Mutex
	gots := make(map[configParam]map[int64]string)
	key := configParam{
		Group:  "g1",
		DataID: "d1",
	}

	id1 := GetUniqueID()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		// register
		c.RegisterConfigCallback(vo.ConfigParam{
			DataId: "d1",
			Group:  "g1",
		}, func(s string, cp ConfigParser) {
			gotlock.Lock()
			defer gotlock.Unlock()
			ids, ok := gots[key]
			if !ok {
				ids = map[int64]string{}
				gots[key] = ids
			}
			ids[id1] = s
		}, id1)
	}()

	id2 := GetUniqueID()
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.RegisterConfigCallback(vo.ConfigParam{
			DataId: "d1",
			Group:  "g1",
		}, func(s string, cp ConfigParser) {
			gotlock.Lock()
			defer gotlock.Unlock()
			ids, ok := gots[key]
			if !ok {
				ids = map[int64]string{}
				gots[key] = ids
			}
			ids[id2] = s
		}, id2)
	}()
	wg.Wait()

	// first change
	fake.change(configParam{
		DataID: "d1",
		Group:  "g1",
	}, "first change")

	assert.Equal(t, map[configParam]map[int64]string{
		{
			Group:  "g1",
			DataID: "d1",
		}: {
			id1: "first change",
			id2: "first change",
		},
	}, gots)

	// second change
	c.DeregisterConfig(vo.ConfigParam{
		DataId: "d1",
		Group:  "g1",
	}, id2)

	fake.change(configParam{
		DataID: "d1",
		Group:  "g1",
	}, "second change")

	assert.Equal(t, map[configParam]map[int64]string{
		{
			Group:  "g1",
			DataID: "d1",
		}: {
			id1: "second change",
			id2: "first change",
		},
	}, gots)

	// third change
	c.DeregisterConfig(vo.ConfigParam{
		DataId: "d1",
		Group:  "g1",
	}, id1)

	fake.change(configParam{
		DataID: "d1",
		Group:  "g1",
	}, "third change")

	assert.Equal(t, map[configParam]map[int64]string{
		{
			Group:  "g1",
			DataID: "d1",
		}: {
			id1: "second change",
			id2: "first change",
		},
	}, gots)
}
