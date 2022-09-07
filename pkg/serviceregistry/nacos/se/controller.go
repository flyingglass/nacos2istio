package se

import (
	"github.com/flyingglass/nacos2istio/pkg/serviceregistry/nacos/common"
	nacos "github.com/flyingglass/nacos2istio/pkg/serviceregistry/nacos/sdk"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	"istio.io/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
	"sync"
	"time"
)

type Controller struct {
	mutex         sync.Mutex
	nc            naming_client.INamingClient
	ncAddr        string
	ncNeedWatchNS map[string]bool
	ncWatchedNS   map[string]bool
	ic            *istioclient.Clientset
	eventChan     chan []common.NacosServiceInstance
}

func NewController(ncAddr string, ncWatchNs string) (*Controller, error) {
	nc, err := nacos.NewNacosNamingClient(ncAddr, "")
	if err != nil {
		log.Errorf("failed to new nacos client consumer client: %v", err)
		return nil, err
	}
	ic, err := getIstioClient()
	if err != nil {
		log.Errorf("failed to create istio client: %v", err)
		return nil, err
	}

	return &Controller{
		ic:            ic,
		nc:            nc,
		ncAddr:        ncAddr,
		ncNeedWatchNS: convertNeedWatchedNS(ncWatchNs),
		ncWatchedNS:   make(map[string]bool),
		eventChan:     make(chan []common.NacosServiceInstance),
	}, nil
}

func (c *Controller) Run(stop <-chan struct{}) {
	go c.watchNacos(stop)
	go c.watchServices(stop)

}

func (c *Controller) watchNacos(stop <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			nameSpaces, err := c.nc.GetAllNamespaces()
			if err != nil {
				log.Errorf("failed to get all namespaces: %v", err)
			}
			for _, ns := range nameSpaces {
				if c.ncNeedWatchNS[ns.NamespaceShowName] && !c.ncWatchedNS[ns.Namespace] {
					namespaceWatcher, err := NewNamespaceWatcher(c.ncAddr, ns.Namespace, ns.NamespaceShowName, c.eventChan)
					if err != nil {
						log.Errorf("failed to watch namespace %s", ns.Namespace, err)
					} else {
						go namespaceWatcher.Run(stop)
						c.ncWatchedNS[ns.Namespace] = true
						log.Infof("start watching namespace %s", ns.Namespace)
					}
				}
			}
		case <-stop:
			return
		}
	}
}

func (c *Controller) watchServices(stop <-chan struct{}) {
	for {
		select {
		case services := <-c.eventChan:
			c.mutex.Lock()

			serviceEntries := ConvertServiceEntry(services)
			for _, se := range serviceEntries {
				SyncServices2IstioUntilMaxRetries(se, c.ic)
			}

			c.mutex.Unlock()
		case <-stop:
			return
		}
	}
}

func getIstioClient() (*istioclient.Clientset, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	ic, err := istioclient.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return ic, nil
}

func convertNeedWatchedNS(ns string) map[string]bool {
	needWatchNS := make(map[string]bool)
	for _, s := range strings.Split(ns, ",") {
		needWatchNS[s] = true
	}
	return needWatchNS
}
