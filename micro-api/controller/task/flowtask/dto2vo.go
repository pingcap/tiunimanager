
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

package flowtask

import (
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"strconv"
	"time"
)

func ParseFlowFromDTO (dto *clusterpb.FlowDTO) FlowWorkDisplayInfo {
	flow := FlowWorkDisplayInfo{
		Id:           uint(dto.Id),
		FlowWorkName: dto.FlowName,
		ClusterId:    dto.BizId,
		StatusInfo: controller.StatusInfo{
			CreateTime: time.Unix(dto.CreateTime, 0),
			UpdateTime: time.Unix(dto.UpdateTime, 0),
			DeleteTime: time.Unix(dto.DeleteTime, 0),
			StatusCode: strconv.Itoa(int(dto.Status)),
			StatusName: dto.StatusName,
		},
	}

	if dto.Operator != nil {
		flow.ManualOperator = true
		flow.OperatorName = dto.Operator.Name
		flow.OperatorId = dto.Operator.Id
		flow.TenantId = dto.Operator.TenantId
	}
	return flow
}

func ParseFlowWorkDetailInfoFromDTO (dto *clusterpb.FlowWithTaskDTO) FlowWorkDetailInfo {
	flow := FlowWorkDetailInfo{
		FlowWorkDisplayInfo: ParseFlowFromDTO(dto.Flow),
		FlowWorkDefineInfo: FlowWorkDefineInfo{
			TaskName: dto.TaskDef,
		},
		Tasks: make([]FlowWorkTaskInfo, 0),
	}

	if len(dto.Tasks) > 0 {
		for _, t := range dto.Tasks {
			flow.Tasks = append(flow.Tasks, ParseTaskFromDTO(t))
		}
	}

	return flow
}

func ParseTaskFromDTO (dto *clusterpb.TaskDTO) FlowWorkTaskInfo {
	task := FlowWorkTaskInfo {
		Id:         uint(dto.Id),
		TaskName:   dto.TaskName,
		Parameters: dto.Parameters,
		Result:     dto.Result,
		Status: int(dto.Status),
		StartTime: time.Unix(dto.StartTime, 0),
		EndTime: time.Unix(dto.EndTime, 0),
	}
	return task
}