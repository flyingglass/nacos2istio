package we

import (
	"context"
	"github.com/flyingglass/nacos2istio/pkg/serviceregistry/nacos/common"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	"istio.io/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SyncServices2IstioUntilMaxRetries(instances []common.NacosServiceInstance, ic *istioclient.Clientset)  {
	err := syncService2Istio(instances, ic)
	retries := 0

	for err != nil {
		if common.IsRetryableError(err) && retries < common.MaxRetries {
			log.Errorf("Failed to synchronize nacos services to Istio, error: %v,  retrying %v ...", err, retries)
			err = syncService2Istio(instances, ic)
			retries++
		} else {
			log.Errorf("Failed to synchronize nacos services to Istio: %v", err)
			err = nil
		}
	}
}

func syncService2Istio(instances []common.NacosServiceInstance, ic *istioclient.Clientset) error {
	// 1. 创建ServiceEntry(外部负责对齐)
	// 2. WorkloadEntry create/update/delete
	for _, nsInstance := range instances {
		weList, err := getAerakiWorkloadEntryList(nsInstance.Namespace, ic)

		if err != nil {
			return err
		}
		newInstances := make([]common.NacosServiceInstance, 0)
		removeWeList := make([]*v1alpha3.WorkloadEntry, 0)
		// newList
		for _, instance := range instances {
			host, found := ConvertWorkloadEntryName(instance), false
			for _, we := range weList.Items {
				if we.Name == host {
					found = true
				}
			}

			if !found {
				newInstances = append(newInstances, instance)
			}
			// update?
		}

		for _, instance := range newInstances {
			_ = createWorkloadEntry(ConvertWorkloadEntry(instance), ic)
		}

		// delete
		for _, we := range weList.Items {
			found := false
			for _, instance := range instances {
				host := ConvertWorkloadEntryName(instance)
				if we.Name == host {
					found = true
				}
			}

			if !found {
				removeWeList = append(removeWeList, we)
			}
		}

		for _, se := range removeWeList {
			_ = deleteWorkloadEntry(se.Namespace, se.Name, ic)
		}

		log.Infof("new: %v, \nremove: %v", newInstances, removeWeList)
	}


	return nil
}

func createWorkloadEntry(workloadEntry *v1alpha3.WorkloadEntry, ic *istioclient.Clientset) error {
	_, err := ic.NetworkingV1alpha3().WorkloadEntries(workloadEntry.Namespace).Create(context.TODO(), workloadEntry,
		metav1.CreateOptions{FieldManager: common.AerakiFieldManager})
	if err == nil {
		log.Infof("service entry %s has been created: %s", workloadEntry.Name, common.Struct2JSON(workloadEntry))
	} else {
		log.Errorf("failed to create workload entry: %s for err: %v", common.Struct2JSON(workloadEntry), err)
	}
	return err
}

func deleteWorkloadEntry(namespace, name string, ic *istioclient.Clientset) error {
	err := ic.NetworkingV1alpha3().WorkloadEntries(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err == nil {
		log.Infof("service entry %s has been deleted: %s", name)
	} else {
		log.Errorf("failed to delete workload entry %s on %s namespace for err: %v", name, namespace, err)
	}
	return err
}

func updateWorkloadEntry(new *v1alpha3.WorkloadEntry, old *v1alpha3.WorkloadEntry, ic *istioclient.Clientset) error {
	new.Spec.Ports = old.Spec.Ports
	new.Spec.Address = old.Spec.Address
	new.ResourceVersion = old.ResourceVersion
	_, err := ic.NetworkingV1alpha3().WorkloadEntries(new.Namespace).Update(context.TODO(), new,
		metav1.UpdateOptions{FieldManager: common.AerakiFieldManager})
	if err == nil {
		log.Infof("workload entry %s has been updated: %s", new.Name, common.Struct2JSON(new))
	}
	return err
}



func getAerakiWorkloadEntryList(ns string, ic *istioclient.Clientset) (*v1alpha3.WorkloadEntryList, error) {
	weList, err := ic.NetworkingV1alpha3().WorkloadEntries(ns).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "manager=" + common.AerakiFieldManager + ", registry=" + common.AerakiRegistry,
	})
	if err != nil {
		log.Errorf("Error list services entry: %v", err)
		return nil, err
	}
	return weList, err
}

