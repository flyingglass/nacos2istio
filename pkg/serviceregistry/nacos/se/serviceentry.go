package se

import (
	"context"
	"github.com/flyingglass/nacos2istio/pkg/serviceregistry/nacos/common"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	"istio.io/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SyncServices2IstioUntilMaxRetries will try to synchronize service entries to the Istio until max retry number
func SyncServices2IstioUntilMaxRetries(new *v1alpha3.ServiceEntry, ic *istioclient.Clientset) {

	err := syncService2Istio(new, ic)
	retries := 0
	for err != nil {
		if common.IsRetryableError(err) && retries < common.MaxRetries {
			log.Errorf("Failed to synchronize nacos services to Istio, error: %v,  retrying %v ...", err, retries)
			err = syncService2Istio(new, ic)
			retries++
		} else {
			log.Errorf("Failed to synchronize nacos services to Istio: %v", err)
			err = nil
		}
	}
}

func syncService2Istio(new *v1alpha3.ServiceEntry, ic *istioclient.Clientset) error {
	existingServiceEntry, err := ic.NetworkingV1alpha3().ServiceEntries(new.Namespace).Get(context.TODO(), new.Name,
		metav1.GetOptions{},
	)

	if common.IsRealError(err) {
		return err
	} else if common.IsNotFound(err) {
		return createServiceEntry(new, ic)
	} else {
		return updateServiceEntry(new, existingServiceEntry, ic)
	}
}

func createServiceEntry(serviceEntry *v1alpha3.ServiceEntry, ic *istioclient.Clientset) error {
	_, err := ic.NetworkingV1alpha3().ServiceEntries(serviceEntry.Namespace).Create(context.TODO(), serviceEntry,
		metav1.CreateOptions{FieldManager: common.AerakiFieldManager})
	if err == nil {
		log.Infof("service entry %s has been created: %s", serviceEntry.Name, common.Struct2JSON(serviceEntry))
	}
	return err
}

func updateServiceEntry(new *v1alpha3.ServiceEntry,
	old *v1alpha3.ServiceEntry, ic *istioclient.Clientset) error {
	//new.Spec.Ports = old.Spec.Ports
	new.ResourceVersion = old.ResourceVersion
	_, err := ic.NetworkingV1alpha3().ServiceEntries(new.Namespace).Update(context.TODO(), new,
		metav1.UpdateOptions{FieldManager: common.AerakiFieldManager})
	if err == nil {
		log.Infof("service entry %s has been updated: %s", new.Name, common.Struct2JSON(new))
	}
	return err
}

func mergeServiceEntryEndpoints(new *v1alpha3.ServiceEntry, old *v1alpha3.ServiceEntry) {
	if old == nil || old.Spec.WorkloadSelector != nil {
		return
	}
	endpoints := new.Spec.Endpoints
	for _, ep := range old.Spec.Endpoints {
		endpoints = append(endpoints, ep)
	}
	new.Spec.Endpoints = endpoints
}
