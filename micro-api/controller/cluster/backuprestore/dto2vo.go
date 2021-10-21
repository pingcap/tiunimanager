
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
