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
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap/tiup/pkg/cluster/spec"
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
	}
	aggregation.CurrentTopologyConfigRecord = &TopologyConfigRecord{
		TenantId:    aggregation.Cluster.TenantId,
		ClusterId:   aggregation.Cluster.Id,
	}

	return aggregation
}

func TestClusterAggregation_ExtractComponentDTOs(t *testing.T) {
	aggregation := &ClusterAggregation{
		Cluster: &Cluster{
			ClusterType: *knowledge.ClusterTypeFromCode("TiDB"),
			ClusterVersion: *knowledge.ClusterVersionFromCode("v4.0.12"),
		},
		CurrentTopologyConfigRecord: &TopologyConfigRecord{
			ConfigModel: &spec.Specification{
				TiDBServers: []*spec.TiDBSpec{
					{Host: "127.0.0.1"},
					{Host: "127.0.0.2"},
				},
				TiKVServers: []*spec.TiKVSpec{
					{Host: "127.0.0.1"},
					{Host: "127.0.0.2"},
				},
				PDServers: []*spec.PDSpec{
					{Host: "127.0.0.1"},
					{Host: "127.0.0.2"},
				},
				TiFlashServers: []*spec.TiFlashSpec{
					{Host: "127.0.0.1"},
					{Host: "127.0.0.2"},
				},
			},
		},
	}

	components := aggregation.ExtractComponentDTOs()

	assert.Equal(t, 4, len(components))
	assert.Equal(t, "127.0.0.2", components[2].Nodes[1].NodeId)

}