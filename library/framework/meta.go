/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package framework

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type ServiceNameEnum string

const (
	MetaDBService  ServiceNameEnum = "em.db"
	ClusterService ServiceNameEnum = "em.cluster"
	ApiService     ServiceNameEnum = "em.api"
	FileService    ServiceNameEnum = "em.filemng"
)

func (s ServiceNameEnum) ServerName() string {
	switch s {
	case ClusterService:
		return "cluster-server"
	case ApiService:
		return "openapi-server"
	case FileService:
		return "file-server"
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
	ServiceAddress  string
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
		ServiceAddress:  args.Host + ":" + strconv.Itoa(args.Port),
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
