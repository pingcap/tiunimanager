
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
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClusterAggregation_ExtractInstancesDTO(t *testing.T) {
	got := buildAggregation().ExtractInstancesDTO()
	assert.Equal(t, "127.0.0.1:4000", got.ExtranetConnectAddresses[0])
}

func buildAggregation() *ClusterAggregation {
	aggregation := &ClusterAggregation{
		Cluster: &Cluster{
			Id:             "111",
			TenantId:       "222",
			ClusterType:    *knowledge.ClusterTypeFromCode("TiDB"),
			ClusterVersion: *knowledge.ClusterVersionFromCode("v5.0.0"),
		},
		AvailableResources: &clusterpb.AllocHostResponse{
			TidbHosts: []*clusterpb.AllocHost{
				{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/"}},
			},
			TikvHosts: []*clusterpb.AllocHost{
				{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/"}},
			},
			PdHosts: []*clusterpb.AllocHost{
				{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/"}},
				{Ip: "127.0.0.1", Disk: &clusterpb.Disk{Path: "/"}},
			},
		},
	}
	aggregation.CurrentTopologyConfigRecord = &TopologyConfigRecord{
		TenantId:    aggregation.Cluster.TenantId,
		ClusterId:   aggregation.Cluster.Id,
		ConfigModel: convertConfig(aggregation.AvailableResources, aggregation.Cluster),
	}

	return aggregation
}

func TestClusterAggregation_ExtractComponentDTOs(t *testing.T) {
	got := buildAggregation().ExtractComponentDTOs()
	assert.Equal(t, "127.0.0.1", got[0].Nodes[1].Instance.HostId)
}
