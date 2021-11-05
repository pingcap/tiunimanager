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
	AddCluster(ctx context.Context, cluster *Cluster) error

	Persist(ctx context.Context, aggregation *ClusterAggregation) error

	Load(ctx context.Context, id string) (cluster *ClusterAggregation, err error)

	Query(ctx context.Context, clusterId, clusterName, clusterType, clusterStatus, clusterTag string, page, pageSize int) ([]*ClusterAggregation, int, error)
}

type ClusterAccessProxy interface {
	QueryParameterJson(ctx context.Context, clusterId string) (string, error)
}

type MetadataManager interface {
	FetchFromRemoteCluster(ctx context.Context, request *clusterpb.ClusterTakeoverReqDTO) (spec.Metadata, error)
	RebuildMetadataFromComponents(ctx context.Context, cluster *Cluster, components []*ComponentGroup) (spec.Metadata, error)
	ParseComponentsFromMetaData(ctx context.Context, meta spec.Metadata) ([]*ComponentGroup, error)
	ParseClusterInfoFromMetaData(ctx context.Context, meta spec.BaseMeta) (clusterType string, user string, group string, version string)
}

type ClusterTopologyPlanner interface {
	BuildComponents(ctx context.Context, cluster *Cluster, demands []*ClusterComponentDemand) ([]*ComponentGroup, error)
	AnalysisResourceRequest(ctx context.Context, cluster *Cluster, components []*ComponentGroup) (*clusterpb.BatchAllocRequest, error)
	ApplyResourceToComponents(ctx context.Context, components []*ComponentGroup, response *clusterpb.BatchAllocResponse) error
}

type TaskRepository interface {
	AddFlowWork(ctx context.Context, flowWork *FlowWorkEntity) error
	AddFlowTask(ctx context.Context, task *TaskEntity, flowId uint) error
	AddCronTask(ctx context.Context, cronTask *CronTaskEntity) error

	Persist(ctx context.Context, flowWork *FlowWorkAggregation) error

	LoadFlowWork(ctx context.Context, id uint) (*FlowWorkEntity, error)
	Load(ctx context.Context, id uint) (flowWork *FlowWorkAggregation, err error)

	QueryCronTask(ctx context.Context, bizId string, cronTaskType int) (cronTask *CronTaskEntity, err error)
	PersistCronTask(ctx context.Context, cronTask *CronTaskEntity) (err error)

	ListFlows(ctx context.Context, bizId, keyword string, status int, page int, pageSize int) ([]*FlowWorkEntity, int, error)
}

type ComponentParser interface {
	GetComponent() *knowledge.ClusterComponent
	ParseComponent(spec *spec.Specification) *ComponentGroup
}
