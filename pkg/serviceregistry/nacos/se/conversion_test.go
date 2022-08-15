package se

import (
	"github.com/flyingglass/nacos2istio/pkg/serviceregistry/nacos/common"
	"istio.io/pkg/log"
	"testing"
)

func TestConversion(t *testing.T) {
	instances := make([]common.NacosServiceInstance, 0)
	instances = append(instances,
		common.NacosServiceInstance{
			Ip:            "172.16.16.6",
			Port:          8899,
			ServiceName:   "provider-api",
			Group:         "group",
			Metadata: map[string]string{
				"zone": "default",
				"region": "default",
				"group": "group",
				"env": "dev",
				"version": "v1",
			},
			Namespace:     "",
			NamespaceName: "public",
		},
		common.NacosServiceInstance{
			Ip:            "172.16.96.32",
			Port:          8899,
			ServiceName:   "provider-api",
			Group:         "group",
			Metadata: map[string]string{
				"zone": "default",
				"region": "default",
				"group": "group",
				"env": "dev",
				"version": "v1",
			},
			Namespace:     "",
			NamespaceName: "public",
		})
	serviceEntries := ConvertServiceEntry(instances)
	for _, se := range serviceEntries {
		log.Info(se.Spec.Endpoints)
	}

	// again
	serviceEntries = ConvertServiceEntry(instances)
	for _, se := range serviceEntries {
		log.Info(se.Spec.Endpoints)
	}
}
