
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

package domain

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap/tiup/pkg/cluster/spec"
)

var ClusterRepo ClusterRepository

var TaskRepo TaskRepository

var RemoteClusterProxy ClusterAccessProxy

var MetadataMgr MetadataManager

var TopologyPlanner ClusterTopologyPlanner

type ClusterRepository interface {
	AddCluster (cluster *Cluster) error

	Persist(aggregation *ClusterAggregation) error

	Load (id string) (cluster *ClusterAggregation, err error)

	Query (clusterId, clusterName, clusterType, clusterStatus, clusterTag string, page, pageSize int) ([]*ClusterAggregation, int, error)

}

type ClusterAccessProxy interface {
	QueryParameterJson(clusterId string) (string, error)
}

type MetadataManager interface {
	FetchFromRemoteCluster(ctx context.Context, request *clusterpb.ClusterTakeoverReqDTO) (spec.Metadata, error)
	RebuildMetadataFromComponents(cluster *Cluster, components []*ComponentGroup) (spec.Metadata, error)
	ParseComponentsFromMetaData(spec.Metadata) ([]*ComponentGroup, error)
	ParseClusterInfoFromMetaData(meta spec.BaseMeta) (user string, group string, version string)
}

type ClusterTopologyPlanner interface {
	BuildComponents(cluster *Cluster, demands []*ClusterComponentDemand) ([]*ComponentGroup, error)
	AnalysisResourceRequest(components []*ComponentGroup) (*clusterpb.BatchAllocRequest, error)
	ApplyResourceToComponents(components []*ComponentGroup, response *clusterpb.BatchAllocResponse) error
}

type TaskRepository interface {
	AddFlowWork(flowWork *FlowWorkEntity) error
	AddFlowTask(task *TaskEntity, flowId uint) error
	AddCronTask(cronTask *CronTaskEntity) error

	Persist(flowWork *FlowWorkAggregation) error

	LoadFlowWork(id uint) (*FlowWorkEntity, error)
	Load(id uint) (flowWork *FlowWorkAggregation, err error)

	QueryCronTask(bizId string, cronTaskType int) (cronTask *CronTaskEntity, err error)
	PersistCronTask(cronTask *CronTaskEntity) (err error)

	ListFlows(bizId, keyword string, status int, page int, pageSize int) ([]*FlowWorkEntity, int, error)
}

type ComponentParser interface {
	GetComponent() *knowledge.ClusterComponent
	ParseComponent(spec *spec.Specification) *ComponentGroup
}

