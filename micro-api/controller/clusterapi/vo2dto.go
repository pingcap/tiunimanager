
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package clusterapi

import (
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

func (req *CreateReq) ConvertToDTO() (baseInfoDTO *clusterpb.ClusterBaseInfoDTO, demandsDTO []*clusterpb.ClusterNodeDemandDTO) {
	baseInfoDTO = req.ClusterBaseInfo.ConvertToDTO()

	demandsDTO = make([]*clusterpb.ClusterNodeDemandDTO, 0, len(req.NodeDemandList))

	for _, demand := range req.NodeDemandList {
		if demand.TotalNodeCount <= 0 {
			framework.Log().Infof("Skip empty demand for component %s", demand.ComponentType)
			continue
		}
		items := make([]*clusterpb.DistributionItemDTO, len(demand.DistributionItems), len(demand.DistributionItems))

		for j, item := range demand.DistributionItems {
			items[j] = &clusterpb.DistributionItemDTO{
				ZoneCode: item.ZoneCode,
				SpecCode: item.SpecCode,
				Count:    int32(item.Count),
			}
		}

		demandsDTO = append(demandsDTO, &clusterpb.ClusterNodeDemandDTO{
			ComponentType:  demand.ComponentType,
			TotalNodeCount: int32(demand.TotalNodeCount),
			Items:          items,
		})
	}
	return
}

func (req *RestoreReq) ConvertToDTO() (baseInfoDTO *clusterpb.ClusterBaseInfoDTO, demandsDTO []*clusterpb.ClusterNodeDemandDTO) {
	baseInfoDTO = req.ClusterBaseInfo.ConvertToDTO()

	demandsDTO = make([]*clusterpb.ClusterNodeDemandDTO, 0, len(req.NodeDemandList))

	for _, demand := range req.NodeDemandList {
		if demand.TotalNodeCount <= 0 {
			framework.Log().Infof("Skip empty demand for component %s", demand.ComponentType)
			continue
		}
		items := make([]*clusterpb.DistributionItemDTO, len(demand.DistributionItems), len(demand.DistributionItems))

		for j, item := range demand.DistributionItems {
			items[j] = &clusterpb.DistributionItemDTO{
				ZoneCode: item.ZoneCode,
				SpecCode: item.SpecCode,
				Count:    int32(item.Count),
			}
		}

		demandsDTO = append(demandsDTO, &clusterpb.ClusterNodeDemandDTO{
			ComponentType:  demand.ComponentType,
			TotalNodeCount: int32(demand.TotalNodeCount),
			Items:          items,
		})
	}
	return
}

func (baseInfo *ClusterBaseInfo) ConvertToDTO() (dto *clusterpb.ClusterBaseInfoDTO) {
	dto = &clusterpb.ClusterBaseInfoDTO{
		ClusterName: baseInfo.ClusterName,

		DbPassword:     baseInfo.DbPassword,
		ClusterType:    controller.ConvertTypeDTO(baseInfo.ClusterType),
		ClusterVersion: controller.ConvertVersionDTO(baseInfo.ClusterVersion),
		Tags:           baseInfo.Tags,
		Tls:            baseInfo.Tls,
		RecoverInfo:    controller.ConvertRecoverInfoDTO(baseInfo.RecoverInfo.SourceClusterId, baseInfo.RecoverInfo.BackupRecordId),
	}

	return
}

func (demand *ClusterNodeDemand) ConvertToDTO() (dto *clusterpb.ClusterNodeDemandDTO) {
	items := make([]*clusterpb.DistributionItemDTO, len(demand.DistributionItems), len(demand.DistributionItems))

	for index, item := range demand.DistributionItems {
		items[index] = item.ConvertToDTO()
	}

	dto = &clusterpb.ClusterNodeDemandDTO{
		ComponentType: demand.ComponentType,

		TotalNodeCount: int32(demand.TotalNodeCount),
		Items:          items,
	}
	return
}

func (item DistributionItem) ConvertToDTO() (dto *clusterpb.DistributionItemDTO) {
	dto = &clusterpb.DistributionItemDTO{
		ZoneCode: item.ZoneCode,
		SpecCode: item.SpecCode,
		Count:    int32(item.Count),
	}
	return
}
