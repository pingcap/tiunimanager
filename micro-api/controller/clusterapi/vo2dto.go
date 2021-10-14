package clusterapi

import (
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

func (req *CreateReq) ConvertToDTO() (baseInfoDTO *clusterpb.ClusterBaseInfoDTO, demandsDTO []*clusterpb.ClusterNodeDemandDTO) {
	baseInfoDTO = req.ClusterBaseInfo.ConvertToDTO()

	demandsDTO = make([]*clusterpb.ClusterNodeDemandDTO, 0, len(req.NodeDemandList))

	for _,demand := range req.NodeDemandList {
		if demand.TotalNodeCount <= 0 {
			framework.Log().Infof("Skip empty demand for component %s", demand.ComponentType)
			continue
		}
		items := make([]*clusterpb.DistributionItemDTO, len(demand.DistributionItems), len(demand.DistributionItems))

		for j, item := range demand.DistributionItems {
			items[j] = &clusterpb.DistributionItemDTO{
				ZoneCode: item.ZoneCode,
				SpecCode: item.SpecCode,
				Count: int32(item.Count),
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

	for _,demand := range req.NodeDemandList {
		if demand.TotalNodeCount <= 0 {
			framework.Log().Infof("Skip empty demand for component %s", demand.ComponentType)
			continue
		}
		items := make([]*clusterpb.DistributionItemDTO, len(demand.DistributionItems), len(demand.DistributionItems))

		for j, item := range demand.DistributionItems {
			items[j] = &clusterpb.DistributionItemDTO{
				ZoneCode: item.ZoneCode,
				SpecCode: item.SpecCode,
				Count: int32(item.Count),
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

		DbPassword : baseInfo.DbPassword,
		ClusterType: controller.ConvertTypeDTO(baseInfo.ClusterType),
		ClusterVersion: controller.ConvertVersionDTO(baseInfo.ClusterVersion),
		Tags: baseInfo.Tags,
		Tls: baseInfo.Tls,
		RecoverInfo: controller.ConvertRecoverInfoDTO(baseInfo.RecoverInfo.SourceClusterId, baseInfo.RecoverInfo.BackupRecordId),
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
		Items: items,
	}
	return
}

func (item DistributionItem) ConvertToDTO() (dto *clusterpb.DistributionItemDTO) {
	dto = &clusterpb.DistributionItemDTO{
		ZoneCode: item.ZoneCode,
		SpecCode: item.SpecCode,
		Count: int32(item.Count),
	}
	return
}
