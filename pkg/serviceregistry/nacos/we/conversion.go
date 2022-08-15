package we

import (
	"fmt"
	"github.com/flyingglass/nacos2istio/pkg/serviceregistry/nacos/common"
	istio "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

func ConvertNacosInfo(annotations map[string]string)(*common.NacosInfo, error) {
	serviceName, existed := annotations["aeraki.net/nacosService"]
	if !existed {
		return nil, fmt.Errorf("polaris info should have [annotation]: aeraki.net/nacosService")
	}
	namespace, existed := annotations["aeraki.net/nacosNamespace"]
	if !existed {
		namespace = ""
	}

	namespaceName, existed := annotations["aeraki.net/nacosNamespaceName"]
	if !existed {
		namespaceName = common.DefaultIstioNS
	}
	group, existed := annotations["aeraki.net/nacosGroup"]
	if !existed {
		group = "DEFAULT_GROUP"
	}
	return &common.NacosInfo{
		ServiceName: serviceName,
		Namespace:   namespace,
		NamespaceName: namespaceName,
		Group:       group,
	}, nil
}

//ConvertWorkloadEntry serviceAccount and metadata todo
func ConvertWorkloadEntry(instance common.NacosServiceInstance) *v1alpha3.WorkloadEntry{
	labels := make(map[string]string)
	if app, existed:= instance.Metadata["aeraki_meta_we_label_app"]; !existed {
		labels["app"] = instance.ServiceName + "-deploy"
	} else {
		labels["app"] = app
	}
	if version, existed := instance.Metadata["aeraki_meta_we_label_version"]; !existed {
		labels["version"] = "v1"
	} else {
		labels["version"] = version
	}
	spec := istio.WorkloadEntry{
		Address:        instance.Ip,
		Ports:          map[string]uint32{common.HttpProtocol: instance.Port},
		Labels:         labels,
	}

	annotations := make(map[string]string)
	annotations["aeraki.net/nacosNamespace"] = instance.Namespace
	annotations["aeraki.net/nacosService"] = instance.ServiceName


	return &v1alpha3.WorkloadEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConvertWorkloadEntryName(instance),
			Namespace: common.ConvertIstioNS(instance.NamespaceName),
			Labels: map[string]string{
				"manager":  common.AerakiFieldManager,
				"registry": "nacos",
			},
			Annotations: annotations,
		},
		Spec:       spec,
	}
}

func ConvertWorkloadEntryName(instance common.NacosServiceInstance) string {
	validName := []string{
		common.AerakiFieldManager,
		instance.Ip, strconv.Itoa(int(instance.Port)),
		instance.ServiceName, instance.Group,
	}

	return strings.ToLower(strings.Join(validName, "-"))
}

func ConvertServiceEntryHost(instance common.NacosServiceInstance) string {
	host := strings.ReplaceAll(instance.ServiceName, ".", "-") + instance.Namespace + "svc.svc.cluster.local"
	return strings.ToLower(host)
}

func constructWorkloadEntry(ip string, serviceAccount string, port uint32, locality string,
	labels map[string]string) *istio.WorkloadEntry {
	return &istio.WorkloadEntry{
		Address:        ip,
		Ports:          map[string]uint32{common.HttpProtocol: port},
		ServiceAccount: serviceAccount,
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

// constructServiceEntryName constructs the service entry name for a given nacos service
func constructServiceEntryName(service string) string {
	validDNSName := strings.ReplaceAll(strings.ToLower(service), ".", "-")
	return common.AerakiFieldManager + "-" + validDNSName
}