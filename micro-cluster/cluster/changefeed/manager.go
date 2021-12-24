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
 ******************************************************************************/

package changefeed

import (
	"context"
	"encoding/json"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func copyDownstreamConfig(target *changefeed.ChangeFeedTask, changeFeedType string, content interface{}) error {
	target.Type = constants.DownstreamType(changeFeedType)
	config, err := json.Marshal(content)
	if err != nil {
		return err
	}
	target.Downstream, err = changefeed.UnmarshalDownstream(target.Type, string(config))
	return err
}

// Create
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter name
// @return string ID of ChangeFeedTask
// @return error
func (p *Manager) Create(ctx context.Context, request cluster.CreateChangeFeedTaskReq) (resp cluster.CreateChangeFeedTaskResp, err error) {
	task := &changefeed.ChangeFeedTask {
		Entity: dbCommon.Entity {
			TenantId: framework.GetTenantIDFromContext(ctx),
		},
		Name: request.Name,
		ClusterId: request.ClusterID,
		StartTS: request.StartTS,
		FilterRules: request.FilterRules,
	}

	err = copyDownstreamConfig(task, request.DownstreamType, request.Downstream)

	task, err = models.GetChangeFeedReaderWriter().Create(ctx, task)
	resp.ID = task.ID

	models.GetChangeFeedReaderWriter().LockStatus(ctx, task.ID)
	models.GetChangeFeedReaderWriter().UnlockStatus(ctx, task.ID, constants.ChangeFeedStatusNormal)

	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to create change feed task, %s", err.Error())
		return cluster.CreateChangeFeedTaskResp{}, framework.WrapError(common.TIEM_CHANGE_FEED_CREATE_ERROR, "failed to create change feed task", err)
	}


	return
}

func (p *Manager) Delete(ctx context.Context, request cluster.DeleteChangeFeedTaskReq) (resp cluster.DeleteChangeFeedTaskResp, err error) {
	err = models.GetChangeFeedReaderWriter().Delete(ctx, request.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to delete change feed task, %s", err.Error())
		return cluster.DeleteChangeFeedTaskResp{
			ID: request.ID,
		}, err
	}
	return
}

func (p *Manager) Pause(ctx context.Context, request cluster.PauseChangeFeedTaskReq) (resp cluster.PauseChangeFeedTaskResp, err error) {
	err = models.GetChangeFeedReaderWriter().LockStatus(ctx, request.ID)
	defer models.GetChangeFeedReaderWriter().UnlockStatus(ctx, request.ID, constants.ChangeFeedStatusStopped)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to delete change feed task, %s", err.Error())
		resp.Status = string(constants.ChangeFeedStatusStopped)
		return
	}

	return
}

func (p *Manager) Resume(ctx context.Context, request cluster.ResumeChangeFeedTaskReq) (resp cluster.ResumeChangeFeedTaskResp, err error) {
	err = models.GetChangeFeedReaderWriter().LockStatus(ctx, request.ID)
	defer models.GetChangeFeedReaderWriter().UnlockStatus(ctx, request.ID, constants.ChangeFeedStatusNormal)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to delete change feed task, %s", err.Error())
		resp.Status = string(constants.ChangeFeedStatusStopped)
		return
	}

	return
}

func (p *Manager) Update(ctx context.Context, request cluster.UpdateChangeFeedTaskReq) (resp cluster.UpdateChangeFeedTaskResp, err error) {
	models.GetChangeFeedReaderWriter().LockStatus(ctx, request.ID)
	models.GetChangeFeedReaderWriter().UnlockStatus(ctx, request.ID, constants.ChangeFeedStatusStopped)

	task := &changefeed.ChangeFeedTask {
		Entity: dbCommon.Entity {
			ID: request.ID,
			TenantId: framework.GetTenantIDFromContext(ctx),
		},
		Name: request.Name,
		FilterRules: request.FilterRules,
	}

	err = copyDownstreamConfig(task, request.DownstreamType, request.Downstream)

	models.GetChangeFeedReaderWriter().UpdateConfig(ctx, task)
	models.GetChangeFeedReaderWriter().LockStatus(ctx, request.ID)
	models.GetChangeFeedReaderWriter().UnlockStatus(ctx, request.ID, constants.ChangeFeedStatusNormal)

	return
}

func (p *Manager) Detail(ctx context.Context, request cluster.DetailChangeFeedTaskReq) (resp cluster.DetailChangeFeedTaskResp, err error) {
	task, err := models.GetChangeFeedReaderWriter().Get(ctx, request.ID)
	if err != nil {
		return
	}

	resp.ChangeFeedTaskInfo = parse(*task)

	return
}

func parse(task changefeed.ChangeFeedTask) cluster.ChangeFeedTaskInfo {
	info := cluster.ChangeFeedTaskInfo{
		ChangeFeedTask: cluster.ChangeFeedTask {
			ID:             task.ID,
			Name:           task.Name,
			ClusterID:      task.ClusterId,
			StartTS:        task.StartTS,
			FilterRules:    task.FilterRules,
			Status:         task.Status,
			DownstreamType: string(task.Type),
			Downstream: task.Downstream,
			CreateTime: task.CreatedAt,
			UpdateTime: task.UpdatedAt,
		},
		UnSteady: task.Locked(),
	}
	return info
}

func (p *Manager) Query(ctx context.Context, request cluster.QueryChangeFeedTaskReq) (resps []cluster.QueryChangeFeedTaskResp, total int, err error) {
	tasks, count, err := models.GetChangeFeedReaderWriter().QueryByClusterId(ctx, request.ClusterId, request.GetOffset(), request.PageSize)
	if err != nil {
		return
	}

	total = int(count)
	resps = make([]cluster.QueryChangeFeedTaskResp, 0)
	for _, task := range tasks {
		resp := cluster.QueryChangeFeedTaskResp{
			ChangeFeedTaskInfo: parse(*task),
		}
		resps = append(resps, resp)
	}
	return
}