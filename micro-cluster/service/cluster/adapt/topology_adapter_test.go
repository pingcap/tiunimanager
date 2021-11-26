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
	"errors"
	"testing"

	"github.com/asim/go-micro/v3/client"
	"github.com/golang/mock/gomock"
	rpc_client "github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
	"github.com/pingcap-inc/tiem/test/mockdb"
	"github.com/stretchr/testify/assert"
)

func TestBuildComponents(t *testing.T) {
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
						Count:    1,
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
						Count:    1,
					},
				},
			},
		},
		&domain.Cluster{
			TenantId:       "testTenantId",
			Id:             "testCluster",
			ClusterVersion: knowledge.ClusterVersion{Code: "v5.0.0", Name: "v5.0.0"},
			ClusterType:    knowledge.ClusterType{Code: "TiDB", Name: "TiDB"},
		})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(got))
	assert.Equal(t, "TiDB", got[0].ComponentType.ComponentType)
	assert.Equal(t, "TiKV", got[1].ComponentType.ComponentType)
}

func TestAnalysisResourceRequest(t *testing.T) {
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
						Count:    1,
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
						Count:    1,
					},
				},
			},
		},
		&domain.Cluster{
			TenantId:       "testTenantId",
			Id:             "testCluster",
			ClusterVersion: knowledge.ClusterVersion{Code: "v5.0.0", Name: "v5.0.0"},
			ClusterType:    knowledge.ClusterType{Code: "TiDB", Name: "TiDB"},
		})
	assert.NoError(t, err)

	got, err := planner.AnalysisResourceRequest(context.TODO(), &domain.Cluster{Id: "testCluster", CpuArchitecture: "x86"}, group, false)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(got.BatchRequests[0].Requires))
	assert.Equal(t, "testCluster", got.BatchRequests[0].Applicant.HolderId)
	assert.Equal(t, "zone1", got.BatchRequests[0].Requires[0].Location.Zone)
	assert.Equal(t, "zone2", got.BatchRequests[0].Requires[1].Location.Zone)
}

func Test_GetClusterPort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mockdb.NewMockTiEMDBService(ctrl)
	mockClient.EXPECT().AllocResourcesInBatch(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, in *dbpb.DBBatchAllocRequest, opts ...client.CallOption) (*dbpb.DBBatchAllocResponse, error) {
			if len(in.BatchRequests) == 1 && len(in.BatchRequests[0].Requires) == 1 && in.BatchRequests[0].Requires[0].Strategy == int32(resource.ClusterPorts) {

				rsp := new(dbpb.DBBatchAllocResponse)
				rsp.Rs = new(dbpb.DBAllocResponseStatus)
				rsp.Rs.Code = int32(0)

				portResource := dbpb.DBPortResource{
					Start: in.BatchRequests[0].Requires[0].Require.PortReq[0].Start,
					End:   in.BatchRequests[0].Requires[0].Require.PortReq[0].End,
					Ports: []int32{in.BatchRequests[0].Requires[0].Require.PortReq[0].Start, in.BatchRequests[0].Requires[0].Require.PortReq[0].Start + 1},
				}

				hostResource := dbpb.DBHostResource{
					PortRes: []*dbpb.DBPortResource{
						&portResource,
					},
					ComputeRes: &dbpb.DBComputeRequirement{},
					DiskRes:    &dbpb.DBDiskResource{},
					Location:   &dbpb.DBLocation{},
				}

				response := dbpb.DBAllocResponse{
					Rs: &dbpb.DBAllocResponseStatus{Code: 0},
				}
				response.Results = append(response.Results, &hostResource)
				rsp.BatchResults = append(rsp.BatchResults, &response)

				return rsp, nil
			}
			return nil, errors.New("Bad Request")
		})
	rpc_client.DBClient = mockClient

	var planner DefaultTopologyPlanner
	requestId := "TEST_REQUEST_ID"
	cluster := &domain.Cluster{
		TenantId:       "testTenantId",
		Id:             "testCluster",
		ClusterVersion: knowledge.ClusterVersion{Code: "v5.0.0", Name: "v5.0.0"},
		ClusterType:    knowledge.ClusterType{Code: "TiDB", Name: "TiDB"},
	}
	ports, err := planner.getClusterPorts(context.TODO(), cluster, requestId)
	assert.Nil(t, err)
	clusterPortRange := knowledge.GetClusterPortRange(cluster.ClusterType.Code, cluster.ClusterVersion.Code)
	assert.Equal(t, clusterPortRange.Count, len(ports))
	assert.Equal(t, clusterPortRange.Start, ports[0])
	assert.Equal(t, clusterPortRange.Start+1, ports[1])
}
