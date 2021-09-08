package clusterapi

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
)

func (req *CreateReq) ConvertToDTO() (baseInfoDTO *cluster.ClusterBaseInfoDTO, demandsDTO []*cluster.ClusterNodeDemandDTO) {
	baseInfoDTO = req.ClusterBaseInfo.ConvertToDTO()

	demandsDTO = make([]*cluster.ClusterNodeDemandDTO, 0, len(req.NodeDemandList))

	for i,demand := range req.NodeDemandList {
		if demand.TotalNodeCount <= 0 {
			framework.Log().Infof("Skip empty demand for component %s", demand.ComponentType)
			continue
		}
		items := make([]*cluster.DistributionItemDTO, len(demand.DistributionItems), len(demand.DistributionItems))

		for j, item := range demand.DistributionItems {
			items[j] = &cluster.DistributionItemDTO{
				ZoneCode: item.ZoneCode,
				SpecCode: item.SpecCode,
				Count: int32(item.Count),
			}
		}

		demandsDTO[i] = &cluster.ClusterNodeDemandDTO{
			ComponentType:  demand.ComponentType,
			TotalNodeCount: int32(demand.TotalNodeCount),
			Items:          items,
		}
	}
	return
}

func (baseInfo *ClusterBaseInfo) ConvertToDTO() (dto *cluster.ClusterBaseInfoDTO) {
	dto = &cluster.ClusterBaseInfoDTO{
		ClusterName: baseInfo.ClusterName,

		DbPassword : baseInfo.DbPassword,
		ClusterType: controller.ConvertTypeDTO(baseInfo.ClusterType),
		ClusterVersion: controller.ConvertVersionDTO(baseInfo.ClusterVersion),
		Tags: baseInfo.Tags,
		Tls: baseInfo.Tls,
	}

	return
}

func (demand *ClusterNodeDemand) ConvertToDTO() (dto *cluster.ClusterNodeDemandDTO) {
	items := make([]*cluster.DistributionItemDTO, len(demand.DistributionItems), len(demand.DistributionItems))

	for index, item := range demand.DistributionItems {
		items[index] = item.ConvertToDTO()
	}

	dto = &cluster.ClusterNodeDemandDTO{
		ComponentType: demand.ComponentType,

		TotalNodeCount: int32(demand.TotalNodeCount),
		Items: items,
	}
	return
}

func (item DistributionItem) ConvertToDTO() (dto *cluster.DistributionItemDTO) {
	dto = &cluster.DistributionItemDTO{
		ZoneCode: item.ZoneCode,
		SpecCode: item.SpecCode,
		Count: int32(item.Count),
	}
	return
}
