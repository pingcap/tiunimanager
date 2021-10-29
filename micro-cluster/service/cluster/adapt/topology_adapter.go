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

package adapt

import (
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"github.com/pingcap-inc/tiem/micro-cluster/service/cluster/domain"
)

type DefaultTopologyPlanner struct {
}

func (d DefaultTopologyPlanner) BuildComponents(cluster *domain.Cluster, demands []*domain.ClusterComponentDemand) ([]*domain.ComponentGroup, error) {
	panic("implement me")
}

func (d DefaultTopologyPlanner) AnalysisResourceRequest(cluster *domain.Cluster, components []*domain.ComponentGroup) (*clusterpb.BatchAllocRequest, error) {
	requirementList := make([]*clusterpb.AllocRequirement, 0)

	for _, component := range components {
		for _, instance := range component.Nodes {
			portRequirementList := make([]*clusterpb.PortRequirement, 0)
			for _, port := range instance.PortList {
				portRequirementList = append(portRequirementList, &clusterpb.PortRequirement{
					Start:   int32(port),
					End:     int32(port + 1),
					PortCnt: 1,
				})
			}
			requirementList = append(requirementList, &clusterpb.AllocRequirement{
				Location: &clusterpb.Location{Host: instance.Host},
				Require: &clusterpb.Requirement{
					PortReq:    portRequirementList,
					DiskReq:    &clusterpb.DiskRequirement{NeedDisk: false},
					ComputeReq: &clusterpb.ComputeRequirement{CpuCores: 0, Memory: 0},
				},
				Count:      1,
				HostFilter: &clusterpb.Filter{},
				Strategy:   int32(resource.UserSpecifyHost),
			})
		}
	}

	allocReq := &clusterpb.BatchAllocRequest{
		BatchRequests: []*clusterpb.AllocRequest{
			{
				Applicant: &clusterpb.Applicant{
					HolderId:  cluster.Id,
					RequestId: uuidutil.GenerateID(),
				},

				Requires: requirementList,
			},
		},
	}

	return allocReq, nil
}

func (d DefaultTopologyPlanner) ApplyResourceToComponents(components []*domain.ComponentGroup, response *clusterpb.BatchAllocResponse) error {
	panic("implement me")
}
