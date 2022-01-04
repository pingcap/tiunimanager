/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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

package management

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type CreateClusterRsp struct {
	ClusterId string `json:"clusterId"`
	ClusterBaseInfo
	controller.StatusInfo
}

type PreviewClusterRsp struct {
	ClusterBaseInfo
	ClusterCommonDemand
	StockCheckResult  []StockCheckItem         `json:"stockCheckResult"`
	CapabilityIndexes []ServiceCapabilityIndex `json:"capabilityIndexes"`
}

type ServiceCapabilityIndex struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
	Unit        string      `json:"unit"`
}

type DeleteClusterRsp struct {
	ClusterId string `json:"clusterId"`
	controller.StatusInfo
}

type ResourceSpec struct {
	Zone struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"zone"`
	Spec struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"spec"`
	Count int `json:"count"`
}

type ClusterTopologyInfo struct {
	CpuArchitecture string `json:"cpuArchitecture" form:"cpuArchitecture"`
	Region          struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"region"`
	ComponentTopology []struct {
		ComponentType string `json:"componentType"`
		ResourceSpec  []ResourceSpec
	} `json:"componentTopology"`
}

type DetailClusterRsp struct {
	ClusterDisplayInfo
	ClusterTopologyInfo
	ClusterMaintenanceInfo
	Components []ComponentInstance `json:"components"`
}

type DescribeDashboardRsp struct {
	ClusterId string `json:"clusterId"`
	Url       string `json:"url"`
	Token     string `json:"token"`
}

type RestartClusterRsp struct {
	ClusterId string `json:"clusterId"`
	controller.StatusInfo
}

type StopClusterRsp struct {
	ClusterId string `json:"clusterId"`
	controller.StatusInfo
}

type DescribeMonitorRsp struct {
	ClusterId  string `json:"clusterId" example:"abc"`
	AlertUrl   string `json:"alertUrl" example:"http://127.0.0.1:9093"`
	GrafanaUrl string `json:"grafanaUrl" example:"http://127.0.0.1:3000"`
}