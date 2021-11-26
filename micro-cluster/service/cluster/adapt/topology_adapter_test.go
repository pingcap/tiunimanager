/*******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 ******************************************************************************/

/*******************************************************************************
 * @File: topology_adapter_test.go
 * @Description:
 * @Author: wangyaozheng@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/17
*******************************************************************************/

package adapt

import (
	"context"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap/tiup/pkg/cluster/spec"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"testing"
)

func TestBuildComponents(t *testing.T)  {
	var planner DefaultTopologyPlanner
	knowledge.LoadKnowledge()
	got, err := planner.BuildComponents(
		context.TODO(),
		[]*domain.ClusterComponentDemand{
			{
				ComponentType: &knowledge.ClusterComponent{
					ComponentType: "TiDB",
					ComponentName: "TiDB",
				},
				TotalNodeCount: 1,
				DistributionItems: []*domain.ClusterNodeDistributionItem{
					{
						SpecCode: "4C8G",
						ZoneCode: "zone1",
						Count: 1,
					},
				},
			},
			{
				ComponentType: &knowledge.ClusterComponent{
					ComponentType: "TiKV",
					ComponentName: "TiKV",
				},
				TotalNodeCount: 1,
				DistributionItems: []*domain.ClusterNodeDistributionItem{
					{
						SpecCode: "8C32G",
						ZoneCode: "zone2",
						Count: 1,
					},
				},
			},
		},
		&domain.Cluster{
			TenantId: "testTenantId",
			Id: "testCluster",
			ClusterVersion: knowledge.ClusterVersion{Code: "v5.0.0", Name: "v5.0.0"},
			ClusterType: knowledge.ClusterType{Code: "TiDB", Name: "TiDB"},
		})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(got))
	assert.Equal(t, "TiDB", got[0].ComponentType.ComponentType)
	assert.Equal(t, "TiKV", got[1].ComponentType.ComponentType)
}

func TestAnalysisResourceRequest(t *testing.T)  {
	var planner DefaultTopologyPlanner
	knowledge.LoadKnowledge()
	group, err := planner.BuildComponents(
		context.TODO(),
		[]*domain.ClusterComponentDemand{
			{
				ComponentType: &knowledge.ClusterComponent{
					ComponentType: "TiDB",
					ComponentName: "TiDB",
				},
				TotalNodeCount: 1,
				DistributionItems: []*domain.ClusterNodeDistributionItem{
					{
						SpecCode: "4C8G",
						ZoneCode: "zone1",
						Count: 1,
					},
				},
			},
			{
				ComponentType: &knowledge.ClusterComponent{
					ComponentType: "TiKV",
					ComponentName: "TiKV",
				},
				TotalNodeCount: 1,
				DistributionItems: []*domain.ClusterNodeDistributionItem{
					{
						SpecCode: "8C32G",
						ZoneCode: "zone2",
						Count: 1,
					},
				},
			},
		},
		&domain.Cluster{
			TenantId: "testTenantId",
			Id: "testCluster",
			ClusterVersion: knowledge.ClusterVersion{Code: "v5.0.0", Name: "v5.0.0"},
			ClusterType: knowledge.ClusterType{Code: "TiDB", Name: "TiDB"},
		})
	assert.NoError(t, err)

	got, err := planner.AnalysisResourceRequest(context.TODO(), &domain.Cluster{Id: "testCluster", CpuArchitecture: "x86"}, group, false)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(got.BatchRequests[0].Requires))
	assert.Equal(t, "testCluster", got.BatchRequests[0].Applicant.HolderId)
	assert.Equal(t, "zone1", got.BatchRequests[0].Requires[0].Location.Zone)
	assert.Equal(t, "zone2", got.BatchRequests[0].Requires[1].Location.Zone)
}

func TestDefaultTopologyPlanner_GenerateTopologyConfig(t *testing.T) {
	var planner DefaultTopologyPlanner
	knowledge.LoadKnowledge()

	topology, err := planner.GenerateTopologyConfig(
		context.TODO(),
		[]*domain.ComponentGroup {
			{
				&knowledge.ClusterComponent {
					ComponentType: "TiDB",
					ComponentName: "TiDB",
				},
				[]*domain.ComponentInstance {
					{
						Host: "127.0.0.1",
						DiskPath: "/test",
						PortList: []int{ 4000, 10000 },
					},
					{
						Host: "127.0.0.3",
						DiskPath: "/test",
						PortList: []int{ 4001, 10002 },
					},
				},
			},

			{
				&knowledge.ClusterComponent {
					ComponentType: "PD",
					ComponentName: "PD",
				},
				[]*domain.ComponentInstance {
					{
						Host: "127.0.0.2",
						DiskPath: "/test",
						PortList: []int{ 4000, 10000, 10001, 10002, 10003, 10004 },
					},
				},
			},
		},
		&domain.Cluster{
			Id: "testCluster",
			CpuArchitecture: "arm64",
			Status: domain.ClusterStatusUnlined,
		},
	)

	assert.NoError(t, err)
	config := spec.Specification{}
	err = yaml.Unmarshal([]byte(topology), &config)
	assert.NoError(t, err)
}


