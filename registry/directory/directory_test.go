// Copyright 2016-2019 hxmhlt
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

package directory

import (
	"context"
	"net/url"
	"strconv"
	"testing"
	"time"
)
import (
	"github.com/stretchr/testify/assert"
)
import (
	"github.com/dubbo/go-for-apache-dubbo/cluster/cluster_impl"
	"github.com/dubbo/go-for-apache-dubbo/common"
	"github.com/dubbo/go-for-apache-dubbo/common/constant"
	"github.com/dubbo/go-for-apache-dubbo/common/extension"
	"github.com/dubbo/go-for-apache-dubbo/protocol/invocation"
	"github.com/dubbo/go-for-apache-dubbo/protocol/protocolwrapper"
	"github.com/dubbo/go-for-apache-dubbo/registry"
)

func TestSubscribe(t *testing.T) {
	registryDirectory, _ := normalRegistryDir()

	time.Sleep(1e9)
	assert.Len(t, registryDirectory.cacheInvokers, 3)
}

func TestSubscribe_Delete(t *testing.T) {
	registryDirectory, mockRegistry := normalRegistryDir()
	time.Sleep(1e9)
	assert.Len(t, registryDirectory.cacheInvokers, 3)
	mockRegistry.MockEvent(&registry.ServiceEvent{Action: registry.ServiceDel, Service: *common.NewURLWithOptions("TEST0", common.WithProtocol("dubbo"))})
	time.Sleep(1e9)
	assert.Len(t, registryDirectory.cacheInvokers, 2)

}
func TestSubscribe_InvalidUrl(t *testing.T) {
	url, _ := common.NewURL(context.TODO(), "mock://127.0.0.1:1111")
	mockRegistry, _ := registry.NewMockRegistry(&common.URL{})
	_, err := NewRegistryDirectory(&url, mockRegistry)
	assert.Error(t, err)
}

func TestSubscribe_Group(t *testing.T) {
	extension.SetProtocol(protocolwrapper.FILTER, protocolwrapper.NewMockProtocolFilter)
	extension.SetCluster("mock", cluster_impl.NewMockCluster)

	regurl, _ := common.NewURL(context.TODO(), "mock://127.0.0.1:1111")
	suburl, _ := common.NewURL(context.TODO(), "dubbo://127.0.0.1:20000")
	suburl.Params.Set(constant.CLUSTER_KEY, "mock")
	regurl.SubURL = &suburl
	mockRegistry, _ := registry.NewMockRegistry(&common.URL{})
	registryDirectory, _ := NewRegistryDirectory(&regurl, mockRegistry)

	go registryDirectory.Subscribe(*common.NewURLWithOptions("testservice"))

	//for group1
	urlmap := url.Values{}
	urlmap.Set(constant.GROUP_KEY, "group1")
	urlmap.Set(constant.CLUSTER_KEY, "failover") //to test merge url
	for i := 0; i < 3; i++ {
		mockRegistry.(*registry.MockRegistry).MockEvent(&registry.ServiceEvent{Action: registry.ServiceAdd, Service: *common.NewURLWithOptions("TEST"+strconv.FormatInt(int64(i), 10), common.WithProtocol("dubbo"),
			common.WithParams(urlmap))})
	}
	//for group2
	urlmap2 := url.Values{}
	urlmap2.Set(constant.GROUP_KEY, "group2")
	urlmap2.Set(constant.CLUSTER_KEY, "failover") //to test merge url
	for i := 0; i < 3; i++ {
		mockRegistry.(*registry.MockRegistry).MockEvent(&registry.ServiceEvent{Action: registry.ServiceAdd, Service: *common.NewURLWithOptions("TEST"+strconv.FormatInt(int64(i), 10), common.WithProtocol("dubbo"),
			common.WithParams(urlmap2))})
	}

	time.Sleep(1e9)
	assert.Len(t, registryDirectory.cacheInvokers, 2)
}

func Test_Destroy(t *testing.T) {
	registryDirectory, _ := normalRegistryDir()

	time.Sleep(1e9)
	assert.Len(t, registryDirectory.cacheInvokers, 3)
	assert.Equal(t, true, registryDirectory.IsAvailable())

	registryDirectory.Destroy()
	assert.Len(t, registryDirectory.cacheInvokers, 0)
	assert.Equal(t, false, registryDirectory.IsAvailable())
}

func Test_List(t *testing.T) {
	registryDirectory, _ := normalRegistryDir()

	time.Sleep(1e9)
	assert.Len(t, registryDirectory.List(&invocation.RPCInvocation{}), 3)
	assert.Equal(t, true, registryDirectory.IsAvailable())

}

func normalRegistryDir() (*registryDirectory, *registry.MockRegistry) {
	extension.SetProtocol(protocolwrapper.FILTER, protocolwrapper.NewMockProtocolFilter)

	url, _ := common.NewURL(context.TODO(), "mock://127.0.0.1:1111")
	suburl, _ := common.NewURL(context.TODO(), "dubbo://127.0.0.1:20000")
	url.SubURL = &suburl
	mockRegistry, _ := registry.NewMockRegistry(&common.URL{})
	registryDirectory, _ := NewRegistryDirectory(&url, mockRegistry)

	go registryDirectory.Subscribe(*common.NewURLWithOptions("testservice"))
	for i := 0; i < 3; i++ {
		mockRegistry.(*registry.MockRegistry).MockEvent(&registry.ServiceEvent{Action: registry.ServiceAdd, Service: *common.NewURLWithOptions("TEST"+strconv.FormatInt(int64(i), 10), common.WithProtocol("dubbo"))})
	}
	return registryDirectory, mockRegistry.(*registry.MockRegistry)
}

func TestMergeUrl(t *testing.T) {
	referenceUrlParams := url.Values{}
	referenceUrlParams.Set(constant.CLUSTER_KEY, "random")
	referenceUrlParams.Set("test3", "1")
	serviceUrlParams := url.Values{}
	serviceUrlParams.Set("test2", "1")
	serviceUrlParams.Set(constant.CLUSTER_KEY, "roundrobin")
	referenceUrl, _ := common.NewURL(context.TODO(), "mock1://127.0.0.1:1111", common.WithParams(referenceUrlParams))
	serviceUrl, _ := common.NewURL(context.TODO(), "mock2://127.0.0.1:20000", common.WithParams(serviceUrlParams))

	mergedUrl := mergeUrl(serviceUrl, &referenceUrl)
	assert.Equal(t, "random", mergedUrl.GetParam(constant.CLUSTER_KEY, ""))
	assert.Equal(t, "1", mergedUrl.GetParam("test2", ""))
	assert.Equal(t, "1", mergedUrl.GetParam("test3", ""))
}
