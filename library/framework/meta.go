package framework

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ServiceNameEnum string

const (
	MetaDBService  ServiceNameEnum = "go.micro.tiem.db"
	ClusterService ServiceNameEnum = "go.micro.tiem.cluster"
	ApiService     ServiceNameEnum = "go.micro.tiem.api"
)

func (s ServiceNameEnum) ServerName() string {
	switch s {
	case MetaDBService:
		return "metadb-server"
	case ClusterService:
		return "cluster-server"
	case ApiService:
		return "openapi-server"
	default:
		log.Error("unexpected ServiceName")
		return ""
	}
}

type ServiceMeta struct {
	ServiceName     ServiceNameEnum
	RegistryAddress []string
	ServiceHost     string
	ServicePort     int
}

func (s *ServiceMeta) GetServiceAddress() string {
	return ":" + strconv.Itoa(s.ServicePort)
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
	if l := copy(registryAddresses, addresses); l != len(addresses) {
		log.Errorf("copy address failed, copied count(%d) not expected(%d)\n", l, len(addresses))
		return nil
	}
	return registryAddresses
}
