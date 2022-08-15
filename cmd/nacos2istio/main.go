package main

import (
	"flag"
	"github.com/flyingglass/nacos2istio/pkg/serviceregistry/nacos/se"
	"github.com/flyingglass/nacos2istio/pkg/serviceregistry/nacos/we"
	"github.com/google/martian/log"
	"os"
	"os/signal"
	"syscall"
)

const (
	//defaultMode se: ServiceEntry mode, we: WorkloadEntry mode
	defaultMode = "se"
)

func main() {
	nacosAddr := flag.String("addr", "", "Nacos Address")
	nacosWatchNs := flag.String("ns", "public", "Nacos Namespace List")
	mode := flag.String("mode", defaultMode, "Nacos2Istio mode")
	flag.Parse()

	log.Infof("Nacos2Istio is starting..., Startup arguments, nacosAddress:%s, nacosWatchNs:%s, mode:%s",
		*nacosAddr, *nacosWatchNs, *mode)
	stopChan := make(chan struct{}, 1)

	if *mode == "se" {
		controller, err  := se.NewController(*nacosAddr, *nacosWatchNs)
		if err != nil {
			log.Errorf("Failed to run controller: %v", err)
		} else {
			controller.Run(stopChan)
		}
	} else if *mode == "we" {
		controller, err := we.NewController(*nacosAddr)
		if err != nil {
			log.Errorf("Failed to run controller: %v", err)
		} else {
			controller.Run(stopChan)
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<- signalChan
	stopChan <- struct{}{}
}
