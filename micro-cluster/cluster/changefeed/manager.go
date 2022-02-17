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
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-cluster/cluster/management/meta"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/util/api/cdc"
	"strconv"
	"sync"
	"time"
)

var manager *Manager
var once sync.Once
var service Service
var serviceOnce sync.Once

type Manager struct{}

func GetManager() *Manager {
	once.Do(func() {
		if manager == nil {
			manager = &Manager{}
		}
	})
	return manager
}

// Create
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter name
// @return string ID of ChangeFeedTask
// @return error
func (p *Manager) Create(ctx context.Context, request cluster.CreateChangeFeedTaskReq) (resp cluster.CreateChangeFeedTaskResp, err error) {
	clusterMeta, err := meta.Get(ctx, request.ClusterID)
	if err != nil {
		return
	}

	cdcAddress := clusterMeta.GetCDCClientAddresses()
	if len(cdcAddress) == 0 {
		err = errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "CDC components required, cluster %s", clusterMeta.Cluster.ID)
		return
	}
	task := &changefeed.ChangeFeedTask{
		Entity: dbCommon.Entity{
			TenantId: framework.GetTenantIDFromContext(ctx),
			Status:   string(constants.ChangeFeedStatusInitial),
		},
		StartTS:     0,
		Name:        request.Name,
		ClusterId:   request.ClusterID,
		FilterRules: request.FilterRules,
	}

	if len(request.StartTS) > 0 {
		tso, parseError := strconv.ParseInt(request.StartTS, 10, 64)
		if parseError == nil {
			task.StartTS = tso
		} else {
			err = errors.NewError(errors.TIEM_PARAMETER_INVALID, parseError.Error())
			return
		}
	}

	if err = copyDownstreamConfig(task, request.DownstreamType, request.Downstream); err != nil {
		return
	}

	task, err = models.GetChangeFeedReaderWriter().Create(ctx, task)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("init change feed task failed, %s", err.Error())
		return
	} else {
		resp.ID = task.ID
	}

	err = p.createExecutor(ctx, clusterMeta, task)

	if err != nil {
		models.GetChangeFeedReaderWriter().Delete(ctx, task.ID)
	}
	return
}

func (p *Manager) Delete(ctx context.Context, request cluster.DeleteChangeFeedTaskReq) (resp cluster.DeleteChangeFeedTaskResp, err error) {
	task, err := models.GetChangeFeedReaderWriter().Get(ctx, request.ID)
	if err != nil {
		return
	} else {
		// return current task status
		resp.ID = task.ID
		resp.Status = task.Status
	}

	clusterMeta, err := meta.Get(ctx, task.ClusterId)
	if err != nil {
		return
	}
	cdcAddress := clusterMeta.GetCDCClientAddresses()
	if len(cdcAddress) == 0 {
		err = errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "CDC components required, cluster %s", clusterMeta.Cluster.ID)
		return
	}
	//ctx = framework.NewBackgroundMicroCtx(ctx, true)
	result, err := cdc.CDCService.DeleteChangeFeedTask(ctx, cdc.ChangeFeedDeleteReq{
		CDCAddress:   clusterMeta.GetCDCClientAddresses()[0].ToString(),
		ChangeFeedID: task.ID,
	})

	if err != nil || !result.Accepted {
		framework.LogWithContext(ctx).Errorf("failed to delete change feed task, err = %v, result = %v", err, result)
	}

	err = models.GetChangeFeedReaderWriter().Delete(ctx, request.ID)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to delete change feed task, %s", err.Error())
		return
	}
	return
}

func (p *Manager) Pause(ctx context.Context, request cluster.PauseChangeFeedTaskReq) (resp cluster.PauseChangeFeedTaskResp, err error) {
	task, err := models.GetChangeFeedReaderWriter().Get(ctx, request.ID)
	if err != nil {
		return
	} else {
		// return current task status
		resp.Status = task.Status
	}

	clusterMeta, err := meta.Get(ctx, task.ClusterId)
	if err != nil {
		return
	}
	cdcAddress := clusterMeta.GetCDCClientAddresses()
	if len(cdcAddress) == 0 {
		err = errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "CDC components required, cluster %s", clusterMeta.Cluster.ID)
		return
	}
	err = models.GetChangeFeedReaderWriter().LockStatus(ctx, request.ID)
	if err != nil {
		return
	}

	err = p.pauseExecutor(ctx, clusterMeta, task)

	return
}

func (p *Manager) Resume(ctx context.Context, request cluster.ResumeChangeFeedTaskReq) (resp cluster.ResumeChangeFeedTaskResp, err error) {
	task, err := models.GetChangeFeedReaderWriter().Get(ctx, request.ID)
	if err != nil {
		return
	} else {
		// return current task status
		resp.Status = task.Status
	}

	clusterMeta, err := meta.Get(ctx, task.ClusterId)
	if err != nil {
		return
	}

	cdcAddress := clusterMeta.GetCDCClientAddresses()
	if len(cdcAddress) == 0 {
		err = errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "CDC components required, cluster %s", clusterMeta.Cluster.ID)
		return
	}
	err = models.GetChangeFeedReaderWriter().LockStatus(ctx, request.ID)
	if err != nil {
		return
	}

	p.resumeExecutor(ctx, clusterMeta, task)

	return
}

// Update
// @Description: update change feed task config
// @Receiver p
// @Parameter ctx
// @Parameter request
// @return resp
// @return err
func (p *Manager) Update(ctx context.Context, request cluster.UpdateChangeFeedTaskReq) (resp cluster.UpdateChangeFeedTaskResp, err error) {
	task, err := models.GetChangeFeedReaderWriter().Get(ctx, request.ID)
	if err != nil {
		return
	} else {
		// return current task status
		resp.Status = task.Status
	}

	running := constants.ChangeFeedStatusNormal.ToString() == task.Status

	task.Name = request.Name
	task.FilterRules = request.FilterRules
	err = copyDownstreamConfig(task, request.DownstreamType, request.Downstream)
	if err != nil {
		return
	}

	clusterMeta, err := meta.Get(ctx, task.ClusterId)
	if err != nil {
		return
	}
	cdcAddress := clusterMeta.GetCDCClientAddresses()
	if len(cdcAddress) == 0 {
		err = errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "CDC components required, cluster %s", clusterMeta.Cluster.ID)
		return
	}
	// pause -> update -> resume
	if running {
		err = models.GetChangeFeedReaderWriter().LockStatus(ctx, request.ID)
		if err != nil {
			return
		}

		err = p.pauseExecutor(ctx, clusterMeta, task)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("update change feed task %s failed, step = pause, err = %s", request.ID, err.Error())
			return
		}
	}

	models.GetChangeFeedReaderWriter().UpdateConfig(ctx, task)
	err = p.updateExecutor(ctx, clusterMeta, task)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("update change feed task %s failed, step = update, err = %s", request.ID, err.Error())
		return
	}

	if running {
		err = models.GetChangeFeedReaderWriter().LockStatus(ctx, request.ID)
		if err != nil {
			return
		}
		err = p.resumeExecutor(ctx, clusterMeta, task)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("update change feed task %s failed, step = resume, err = %s", request.ID, err.Error())
			return
		}
	}
	return
}

func currentTSO() uint64 {
	return uint64((time.Now().UnixNano() / int64(time.Millisecond)) << 18)
}

func (p *Manager) Query(ctx context.Context, request cluster.QueryChangeFeedTaskReq) (resps []cluster.QueryChangeFeedTaskResp, total int, err error) {
	clusterMeta, err := meta.Get(ctx, request.ClusterId)
	if err != nil {
		return
	}

	cdcAddress := clusterMeta.GetCDCClientAddresses()
	if len(cdcAddress) == 0 {
		err = errors.NewErrorf(errors.TIEM_INVALID_TOPOLOGY, "CDC components required, cluster %s", clusterMeta.Cluster.ID)
		return
	}
	// remote
	result, err := cdc.CDCService.QueryChangeFeedTasks(ctx, cdc.ChangeFeedQueryReq{
		CDCAddress: cdcAddress[0].ToString(),
	})

	cdcTaskInstanceMap := make(map[string]cdc.ChangeFeedInfo)
	if err == nil {
		for _, t := range result.Tasks {
			cdcTaskInstanceMap[t.ChangeFeedID] = t
		}
	}

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

		resp.ChangeFeedTaskInfo.AcceptUpstreamUpdateTS(currentTSO())
		if t, ok := cdcTaskInstanceMap[task.ID]; ok {
			resp.AcceptDownstreamSyncTS(t.CheckPointTSO)
		} else {
			resp.AcceptDownstreamSyncTS(0)
		}
		resps = append(resps, resp)
	}

	return
}

func (p *Manager) createExecutor(ctx context.Context, clusterMeta *meta.ClusterMeta, task *changefeed.ChangeFeedTask) (err error) {
	ctx = framework.NewBackgroundMicroCtx(ctx, true)
	libResp, libError := cdc.CDCService.CreateChangeFeedTask(ctx, cdc.ChangeFeedCreateReq{
		CDCAddress:   clusterMeta.GetCDCClientAddresses()[0].ToString(),
		ChangeFeedID: task.ID,
		SinkURI:      task.Downstream.GetSinkURI(),
		StartTS:      uint64(task.StartTS),
		FilterRules:  task.FilterRules,
	})
	if libError != nil || !libResp.Accepted {
		errMsg := fmt.Sprintf("createExecutor change feed task failed, err = %v, resp = %v", libError, libResp)
		framework.LogWithContext(ctx).Errorf(errMsg)
		return errors.NewError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, errMsg)
	}

	if libResp.Succeed {
		models.GetChangeFeedReaderWriter().UnlockStatus(ctx, task.ID, constants.ChangeFeedStatusNormal)
	} else {
		framework.LogWithContext(ctx).Errorf("createExecutor change feed task faile, resp = %v", libResp)
		return errors.NewError(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, libResp.ErrorMsg)
	}

	return
}

func (p *Manager) pauseExecutor(ctx context.Context, clusterMeta *meta.ClusterMeta, task *changefeed.ChangeFeedTask) error {
	libResp, libError := cdc.CDCService.PauseChangeFeedTask(ctx, cdc.ChangeFeedPauseReq{
		CDCAddress:   clusterMeta.GetCDCClientAddresses()[0].ToString(),
		ChangeFeedID: task.ID,
	})

	if libError != nil || !libResp.Accepted || !libResp.Succeed {
		errMsg := fmt.Sprintf("pause change feed task failed, err = %v, resp = %v", libError, libResp)
		framework.LogWithContext(ctx).Errorf(errMsg)
		return errors.NewErrorf(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, errMsg)
	}

	return models.GetChangeFeedReaderWriter().UnlockStatus(ctx, task.ID, constants.ChangeFeedStatusStopped)
}

func (p *Manager) updateExecutor(ctx context.Context, clusterMeta *meta.ClusterMeta, task *changefeed.ChangeFeedTask) error {
	libResp, libError := cdc.CDCService.UpdateChangeFeedTask(ctx, cdc.ChangeFeedUpdateReq{
		CDCAddress:   clusterMeta.GetCDCClientAddresses()[0].ToString(),
		ChangeFeedID: task.ID,
		SinkURI:      task.Downstream.GetSinkURI(),
		TargetTS:     task.TargetTS,
		FilterRules:  task.FilterRules,
	})

	if libError != nil || !libResp.Accepted || !libResp.Succeed {
		errMsg := fmt.Sprintf("update change feed task failed, err = %v, resp = %v", libError, libResp)
		framework.LogWithContext(ctx).Errorf(errMsg)
		return errors.NewErrorf(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, errMsg)
	}
	return nil
}

func (p *Manager) resumeExecutor(ctx context.Context, clusterMeta *meta.ClusterMeta, task *changefeed.ChangeFeedTask) error {
	libResp, libError := cdc.CDCService.ResumeChangeFeedTask(ctx, cdc.ChangeFeedResumeReq{
		CDCAddress:   clusterMeta.GetCDCClientAddresses()[0].ToString(),
		ChangeFeedID: task.ID,
	})

	if libError != nil || !libResp.Accepted || !libResp.Succeed {
		errMsg := fmt.Sprintf("resume change feed task failed, err = %v, resp = %v", libError, libResp)
		framework.LogWithContext(ctx).Errorf(errMsg)
		return errors.NewErrorf(errors.TIEM_CHANGE_FEED_EXECUTE_ERROR, errMsg)
	}

	return models.GetChangeFeedReaderWriter().UnlockStatus(ctx, task.ID, constants.ChangeFeedStatusNormal)
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

func parse(task changefeed.ChangeFeedTask) cluster.ChangeFeedTaskInfo {
	info := cluster.ChangeFeedTaskInfo{
		ChangeFeedTask: cluster.ChangeFeedTask{
			ID:             task.ID,
			Name:           task.Name,
			ClusterID:      task.ClusterId,
			StartTS:        strconv.FormatInt(task.StartTS, 10),
			FilterRules:    task.FilterRules,
			Status:         task.Status,
			DownstreamType: string(task.Type),
			Downstream:     task.Downstream,
			CreateTime:     task.CreatedAt,
			UpdateTime:     task.UpdatedAt,
		},
		UnSteady: task.Locked(),
	}
	info.ConvertStartTS()
	return info
}
