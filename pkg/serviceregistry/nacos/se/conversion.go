package se

import (
	"github.com/flyingglass/nacos2istio/pkg/serviceregistry/nacos/common"
	istio "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	"istio.io/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func ConvertServiceEntry(instances []common.NacosServiceInstance) map[string]*v1alpha3.ServiceEntry {
	serviceEntries := make(map[string]*v1alpha3.ServiceEntry)

	for _, instance := range instances {
		metadata := instance.Metadata
		host := convertServiceEntryHost(instance)

		serviceAccount := metadata["aeraki_meta_app_service_account"]
		if serviceAccount == "" {
			serviceAccount = common.DefaultServiceAccount
		}

		istioNS := common.ConvertIstioNS(instance.NamespaceName)

		// All the providers of a nacos service should be deployed in the same namespace
		if se, existed := serviceEntries[host]; existed {
			if istioNS != se.Namespace {
				log.Errorf("found provider in multiple namespaces: %s %s, ignore provider %v", se.Namespace, istioNS,
					metadata)
				continue
			}
		}

		// We assume that the port of all the provider instances should be the same. Is this a reasonable assumption?
		if se, existed := serviceEntries[host]; existed {
			if instance.Port != se.Spec.Ports[0].TargetPort {
				log.Errorf("found multiple ports for service %s : %v  %v, ignore provider %v", host,
					se.Spec.Ports[0].Number, instance.Port, metadata)
				continue
			}
		}

		locality := strings.ReplaceAll(metadata["aeraki_meta_locality"], "%2F", "/")

		labels := make(map[string]string)
		labels["app"] = instance.ServiceName
		labels["version"] = metadata["version"]

		serviceEntry, exist := serviceEntries[host]
		if !exist {
			serviceEntry = constructServiceEntry(instance)
			serviceEntries[host] = serviceEntry
		}
		serviceEntry.Spec.Endpoints = append(serviceEntry.Spec.Endpoints,
			constructWorkloadEntry(instance.Ip, serviceAccount, instance.Port, locality, labels))
	}

	return serviceEntries

}

func constructServiceEntry(instance common.NacosServiceInstance) *v1alpha3.ServiceEntry {
	host := convertServiceEntryHost(instance)
	spec := &istio.ServiceEntry{
		Hosts:      []string{host},
		Ports:      []*istio.Port{convertPort(instance.Port)},
		Resolution: istio.ServiceEntry_STATIC,
		Location:   istio.ServiceEntry_MESH_INTERNAL,
		Endpoints:  make([]*istio.WorkloadEntry, 0),
	}

	annotations := make(map[string]string)
	annotations["aeraki.net/nacosNamespaceName"] = instance.NamespaceName
	annotations["aeraki.net/nacosGroup"] = instance.Group
	annotations["aeraki.net/nacosService"] = instance.ServiceName

	serviceEntry := &v1alpha3.ServiceEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      convertServiceEntryName(instance),
			Namespace: common.ConvertIstioNS(instance.NamespaceName),
			Labels: map[string]string{
				"manager":  common.AerakiFieldManager,
				"registry": common.AerakiRegistry,
			},
			Annotations:annotations,
		},
		Spec: *spec,
	}
	return serviceEntry
}

// constructWorkloadEntry serviceAccount
func constructWorkloadEntry(ip string, serviceAccount string, port uint32, locality string,
	labels map[string]string) *istio.WorkloadEntry {
	return &istio.WorkloadEntry{
		Address:        ip,
		Ports:          map[string]uint32{common.HttpProtocol: port},
		//ServiceAccount: serviceAccount,
		Locality:       locality,
		Labels:         labels,
	}
}

func convertPort(port uint32) *istio.Port {
	return &istio.Port{
		Number:     80,
		Protocol:   common.HttpProtocol,
		Name:       common.HttpProtocol,
		TargetPort: port,
	}
}

// convertServiceEntryName constructs the service entry name for a given nacos service
func convertServiceEntryName(instance common.NacosServiceInstance) string {
	name := []string{common.AerakiFieldManager, instance.Group, instance.ServiceName}
	return strings.ReplaceAll(strings.ToLower(strings.Join(name, "-")), ".", "-")
}

func convertServiceEntryHost(instance common.NacosServiceInstance) string {
	host := []string{
		strings.ReplaceAll(instance.ServiceName, ".", "-"),
		common.ConvertIstioNS(instance.NamespaceName),
		"svc.cluster.local",
	}
	return strings.ToLower(strings.Join(host, "."))
}
