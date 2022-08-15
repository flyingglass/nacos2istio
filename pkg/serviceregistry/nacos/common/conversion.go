package common

import (
	nacosmodel "github.com/nacos-group/nacos-sdk-go/model"
	"strings"
)

func ConvertNacosServices(namespace, namespaceName string, services []nacosmodel.SubscribeService) ([]NacosServiceInstance, error) {
	instances := make([]NacosServiceInstance, 0)

	for _, service := range services {
		instances = append(instances, NacosServiceInstance{
			Ip:            service.Ip,
			Port:          uint32(service.Port),
			ServiceName:   service.ServiceName,
			Group:         convertGroup(service.InstanceId),
			Namespace:     namespace,
			NamespaceName: namespaceName,
			Metadata:      DeepCopyMap(service.Metadata),
		})
	}

	return instances, nil
}

func ConvertIstioNS(namespaceName string) string {
	if namespaceName == "" || namespaceName == "public" {
		return DefaultIstioNS
	}
	return namespaceName
}

//convertGroup "172.16.96.44#8899#DEFAULT#group@@service-name"
func convertGroup(instanceId string) string {
	if instanceId == "" {
		return DefaultNacosGroup
	}

	strArr := strings.Split(instanceId, "#")
	return strings.Split(strArr[len(strArr) - 1], "@@")[0]
}