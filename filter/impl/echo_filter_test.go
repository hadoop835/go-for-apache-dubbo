// Copyright 2016-2019 Yincheng Fang
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package impl

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

import (
	"github.com/dubbo/go-for-apache-dubbo/common"
	"github.com/dubbo/go-for-apache-dubbo/protocol"
	"github.com/dubbo/go-for-apache-dubbo/protocol/invocation"
)

func TestEchoFilter_Invoke(t *testing.T) {
	filter := GetFilter()
	result := filter.Invoke(protocol.NewBaseInvoker(common.URL{}),
		invocation.NewRPCInvocationForProvider("$echo", []interface{}{"OK"}, nil))
	assert.Equal(t, "OK", result.Result())

	result = filter.Invoke(protocol.NewBaseInvoker(common.URL{}),
		invocation.NewRPCInvocationForProvider("MethodName", []interface{}{"OK"}, nil))
	assert.Nil(t, result.Error())
	assert.Nil(t, result.Result())
}
