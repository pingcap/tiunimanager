package backuprestore

import (
	"time"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/management"
)

func ParseClusterBaseInfoFromDTO(dto *clusterpb.ClusterBaseInfoDTO) (baseInfo *management.ClusterBaseInfo) {
	baseInfo = &management.ClusterBaseInfo{
		ClusterName:    dto.ClusterName,
		DbPassword:     dto.DbPassword,
		ClusterType:    dto.ClusterType.Name,
		ClusterVersion: dto.ClusterVersion.Name,
		Tags:           dto.Tags,
		Tls:            dto.Tls,
	}

	return
}

func ParseStatusFromDTO(dto *clusterpb.DisplayStatusDTO) (statusInfo *controller.StatusInfo) {
	statusInfo = &controller.StatusInfo{
		StatusCode:      dto.StatusCode,
		StatusName:      dto.StatusName,
		InProcessFlowId: int(dto.InProcessFlowId),
		CreateTime:      time.Unix(dto.CreateTime, 0),
		UpdateTime:      time.Unix(dto.UpdateTime, 0),
		DeleteTime:      time.Unix(dto.DeleteTime, 0),
	}

	return
}
