package sdk

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"strconv"
	"strings"
)

// NewNacosNamingClient create a Nacos Naming Client
func NewNacosNamingClient(addr, namespace string) (naming_client.INamingClient, error) {
	host, port, err := parseNacosURL(addr)
	if err != nil {
		return nil, err
	}

	sc := []constant.ServerConfig{
		*constant.NewServerConfig(host, port),
	}

	cc := *constant.NewClientConfig(
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),

		constant.WithLogDir("logs"),
		constant.WithCacheDir("cache"),
		constant.WithLogLevel("info"),
		constant.WithNamespaceId(namespace),
	)

	return clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig: &cc,
		ServerConfigs: sc,
	})
}

func parseNacosURL(addr string) (string, uint64, error) {
	strArray := strings.Split(strings.TrimSpace(addr), ":")
	if len(strArray) < 2 {
		return "", 0, fmt.Errorf("invalid nacos address: %s, please specify a valid nacos address like \"nacos:8848\"", addr)
	}
	port, err := strconv.Atoi(strArray[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid nacos address: %s, please specify a valid nacos address like \"nacos:8848\"", addr)
	}
	return strArray[0], uint64(port), nil
}