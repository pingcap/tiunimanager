package servicemeta

import (
	"github.com/pingcap-inc/tiem/library/framework/args"
	"strconv"
	"strings"
)

type ServiceNameEnum string

const (
	MetaDBService  ServiceNameEnum = "go.micro.tiem.db"
	ClusterService ServiceNameEnum = "go.micro.tiem.cluster"
	ApiService     ServiceNameEnum = "go.micro.tiem.api"
)

type ServiceMeta struct {
	ServiceName     ServiceNameEnum
	RegistryAddress []string
	ServiceHost     string
	ServicePort     int
}

func (s *ServiceMeta) GetServiceAddress() string {
	return s.ServiceHost + ":" + strconv.Itoa(s.ServicePort)
}

func NewServiceMetaFromArgs(serviceName ServiceNameEnum, args *args.ClientArgs) *ServiceMeta {
	return &ServiceMeta{
		ServiceName: serviceName,
		RegistryAddress: splitRegistryAddress(args.RegistryAddress),
		ServicePort: args.Port,
		ServiceHost: args.Host,
	}
}

func splitRegistryAddress(argAddress string) []string {
	addresses := strings.Split(argAddress, ",")
	registryAddresses := make([]string, len(addresses))
	for i, addr := range addresses {
		registryAddresses[i] = addr
	}
	return registryAddresses
}

