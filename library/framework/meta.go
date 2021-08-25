package framework

import (
	"strconv"
	"strings"
)

type ServiceNameEnum string

const (
	MetaDBService  ServiceNameEnum = "go.micro.tiem.db"
	ClusterService ServiceNameEnum = "go.micro.tiem.cluster"
	ApiService     ServiceNameEnum = "go.micro.tiem.api"
)

func (s ServiceNameEnum) ServerName() string {
	switch s {
	case MetaDBService: return "metadb-server"
	case ClusterService:	return "cluster-server"
	case ApiService:	return "openapi-server"
	default:
		panic("unexpected")
	}
}

type ServiceMeta struct {
	ServiceName     ServiceNameEnum
	RegistryAddress []string
	ServiceHost     string
	ServicePort     int
}

func (s *ServiceMeta) GetServiceAddress() string {
	return s.ServiceHost + ":" + strconv.Itoa(s.ServicePort)
}

func NewServiceMetaFromArgs(serviceName ServiceNameEnum, args *ClientArgs) *ServiceMeta {
	return &ServiceMeta{
		ServiceName:     serviceName,
		RegistryAddress: splitRegistryAddress(args.RegistryAddress),
		ServicePort:     args.Port,
		ServiceHost:     args.Host,
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

