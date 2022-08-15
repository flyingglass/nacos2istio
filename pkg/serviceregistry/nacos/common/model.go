package common

//NacosServiceInstance contains the info of a nacos service
type NacosServiceInstance struct {
	Ip        string `json:"ip"`
	Port        uint32            `json:"port"`
	ServiceName string            `json:"serviceName"`
	Group string `json:"group"`
	Metadata    map[string]string `json:"metadata"`
	Namespace     string `json:"namespace"`
	NamespaceName string `json:"namespace_name"`
}

type NacosInfo struct {
	ServiceName string
	Namespace string
	NamespaceName string
	Group string
}