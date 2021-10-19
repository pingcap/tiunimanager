package backuprestore

import (
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/framework"
)

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
