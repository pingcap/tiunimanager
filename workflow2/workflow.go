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

package workflow2

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/pingcap-inc/tiem/models/workflow"
	"sync"
	"time"
)

type WorkFlowManager struct {
	flowDefineMap    sync.Map //key: flowName, value: flowDefine
	nodeGoroutineMap sync.Map //key: flowId, value: stopChannel
	watchInterval    time.Duration
	retry            int //todo: node retry times
}

var workflowService WorkFlowService
var once sync.Once

func GetWorkFlowService() WorkFlowService {
	once.Do(func() {
		if workflowService == nil {
			workflowService = NewWorkFlowManager()
		}
	})
	return workflowService
}

func NewWorkFlowManager() WorkFlowService {
	mgr := &WorkFlowManager{
		watchInterval: 5 * time.Second,
		retry:         1,
	}
	go mgr.watchLoop(context.Background())
	return mgr
}

func MockWorkFlowService(service WorkFlowService) {
	workflowService = service
}

func (mgr *WorkFlowManager) watchLoop(ctx context.Context) {
	ticker := time.NewTicker(mgr.watchInterval)
	for range ticker.C {
		framework.LogWithContext(ctx).Infof("begin workflow watchLoop every %+v", mgr.watchInterval)
		//handle processing workflow last
		mgr.handleUnFinishedWorkFlow(ctx, constants.WorkFlowStatusCanceling)
		mgr.handleUnFinishedWorkFlow(ctx, constants.WorkFlowStatusStopped)
		mgr.handleUnFinishedWorkFlow(ctx, constants.WorkFlowStatusProcessing)
	}
}

func (mgr *WorkFlowManager) handleUnFinishedWorkFlow(ctx context.Context, status string) {
	//todo: recover
	for page, pageSize := 1, defaultPageSize; ; page++ {
		flows, _, err := models.GetWorkFlowReaderWriter().QueryWorkFlows(ctx, "", "", "", status, page, pageSize)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("query workflows by status %s failed %+v", status, err)
			continue
		}
		if len(flows) == 0 {
			return
		}
		for _, flow := range flows {
			switch status {
			case constants.WorkFlowStatusProcessing:
				_, exist := mgr.nodeGoroutineMap.Load(flow.ID)
				if !exist {
					//workflow has no processing goroutine
					flowMeta, err := NewWorkFlowMeta(ctx, flow.ID)
					if err != nil {
						framework.LogWithContext(ctx).Errorf("build workflow meta by flow id %s failed %s", flow.ID, err.Error())
						return
					}
					newCtx, cancelFunc := context.WithCancel(flowMeta.Context)
					flowMeta.Context = NewFlowContext(newCtx, flowMeta.Context.FlowData)
					mgr.nodeGoroutineMap.Store(flow.ID, cancelFunc)
					go func() {
						//todo: recover
						defer mgr.nodeGoroutineMap.Delete(flow.ID) //clean node go routine map whether end or stop
						//load workflow, call executor and handle polling, restore workflow
						flowMeta.Execute()
						return
					}()
				} else {
					//workflow has processing goroutine
					continue
				}
			case constants.WorkFlowStatusStopped:
				cancelFunc, exist := mgr.nodeGoroutineMap.Load(flow.ID)
				if !exist {
					//workflow has no processing goroutine
					continue
				} else {
					//workflow has processing goroutine
					if cancelFunc != nil {
						cancelFunc.(context.CancelFunc)()
						framework.LogWithContext(ctx).Infof("close workflow id %s, name %s stop channel success", flow.ID, flow.Name)
					}
					mgr.nodeGoroutineMap.Delete(flow.ID)
					framework.LogWithContext(ctx).Infof("stop workflow id %s, name %s success", flow.ID, flow.Name)
				}
			case constants.WorkFlowStatusCanceling:
				cancelFunc, exist := mgr.nodeGoroutineMap.Load(flow.ID)
				if exist {
					//workflow has processing goroutine
					if cancelFunc != nil {
						cancelFunc.(context.CancelFunc)()
						framework.LogWithContext(ctx).Infof("close workflow id %s, name %s stop channel success", flow.ID, flow.Name)
					}
					mgr.nodeGoroutineMap.Delete(flow.ID)
				}
				//load workflow, cancel flow and node status, update workflow
				flowMeta, err := NewWorkFlowMeta(ctx, flow.ID)
				if err != nil {
					framework.LogWithContext(ctx).Errorf("build workflow meta by flow id %s failed %s", flow.ID, err.Error())
					return
				}
				if flowMeta.CurrentNode != nil {
					flowMeta.CurrentNode.Status = constants.WorkFlowStatusCanceled
				}
				flowMeta.Flow.Status = constants.WorkFlowStatusCanceled
				flowMeta.Restore()
				framework.LogWithContext(ctx).Infof("cancel workflow id %s, name %s success", flow.ID, flow.Name)
			}
		}
	}
}

func (mgr *WorkFlowManager) RegisterWorkFlow(ctx context.Context, flowName string, flowDefine *WorkFlowDefine) {
	mgr.flowDefineMap.Store(flowName, flowDefine)
	framework.LogWithContext(ctx).Infof("Register WorkFlow %s success, definition: %+v", flowDefine.FlowName, flowDefine)
}

func (mgr *WorkFlowManager) GetWorkFlowDefine(ctx context.Context, flowName string) (*WorkFlowDefine, error) {
	flowDefine, exist := mgr.flowDefineMap.Load(flowName)
	if !exist {
		framework.LogWithContext(ctx).Errorf("WorkFlow %s not exist", flowName)
		return nil, errors.NewErrorf(errors.TIEM_WORKFLOW_DEFINE_NOT_FOUND, "%s workflow definion not exist", flowName)
	}
	return flowDefine.(*WorkFlowDefine), nil
}

func (mgr *WorkFlowManager) CreateWorkFlow(ctx context.Context, bizId string, bizType string, flowName string) (string, error) {
	flowDefine, exist := mgr.flowDefineMap.Load(flowName)
	if !exist {
		return "", errors.NewErrorf(errors.TIEM_WORKFLOW_DEFINE_NOT_FOUND, "%s workflow definion not exist", flowName)
	}

	dataMap := map[string]string{
		framework.TiEM_X_TRACE_ID_KEY:  framework.GetTraceIDFromContext(ctx),
		framework.TiEM_X_USER_ID_KEY:   framework.GetUserIDFromContext(ctx),
		framework.TiEM_X_TENANT_ID_KEY: framework.GetTenantIDFromContext(ctx),
	}
	flow, err := models.GetWorkFlowReaderWriter().CreateWorkFlow(ctx, &workflow.WorkFlow{
		Name:    flowDefine.(*WorkFlowDefine).FlowName,
		BizID:   bizId,
		BizType: bizType,
		Entity: common.Entity{
			Status:   constants.WorkFlowStatusInitializing,
			TenantId: framework.GetTenantIDFromContext(ctx),
		},
		Context: NewFlowContext(ctx, dataMap).GetContextString(),
	})
	return flow.ID, err
}

func (mgr *WorkFlowManager) ListWorkFlows(ctx context.Context, request message.QueryWorkFlowsReq) (resp message.QueryWorkFlowsResp, page structs.Page, err error) {
	flows, total, err := models.GetWorkFlowReaderWriter().QueryWorkFlows(ctx, request.BizID, request.BizType, request.FlowName, request.Status, request.Page, request.PageSize)
	if err != nil {
		return resp, page, errors.WrapError(errors.TIEM_WORKFLOW_QUERY_FAILED, err.Error(), err)
	}

	flowInfos := make([]*structs.WorkFlowInfo, len(flows))
	for index, flow := range flows {
		flowInfos[index] = &structs.WorkFlowInfo{
			ID:         flow.ID,
			Name:       flow.Name,
			BizID:      flow.BizID,
			BizType:    flow.BizType,
			Status:     flow.Status,
			CreateTime: flow.CreatedAt,
			UpdateTime: flow.UpdatedAt,
			DeleteTime: flow.DeletedAt.Time,
		}
	}
	return message.QueryWorkFlowsResp{
			WorkFlows: flowInfos,
		}, structs.Page{
			Page:     request.Page,
			PageSize: request.PageSize,
			Total:    int(total),
		}, nil
}

func (mgr *WorkFlowManager) DetailWorkFlow(ctx context.Context, request message.QueryWorkFlowDetailReq) (resp message.QueryWorkFlowDetailResp, err error) {
	flow, nodes, err := models.GetWorkFlowReaderWriter().QueryDetailWorkFlow(ctx, request.WorkFlowID)
	if err != nil {
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_DETAIL_FAILED, err.Error(), err)
	}

	define, err := mgr.GetWorkFlowDefine(ctx, flow.Name)
	if err != nil {
		return resp, errors.WrapError(errors.TIEM_WORKFLOW_DEFINE_NOT_FOUND, err.Error(), err)
	}

	resp = message.QueryWorkFlowDetailResp{
		Info: &structs.WorkFlowInfo{
			ID:         flow.ID,
			Name:       flow.Name,
			BizID:      flow.BizID,
			BizType:    flow.BizType,
			Status:     flow.Status,
			CreateTime: flow.CreatedAt,
			UpdateTime: flow.UpdatedAt,
			DeleteTime: flow.DeletedAt.Time,
		},
		NodeInfo:  make([]*structs.WorkFlowNodeInfo, 0),
		NodeNames: define.getNodeNameList(),
	}
	for _, node := range nodes {
		resp.NodeInfo = append(resp.NodeInfo, &structs.WorkFlowNodeInfo{
			ID:         node.ID,
			Name:       node.Name,
			Parameters: node.Parameters,
			Result:     node.Result,
			Status:     node.Status,
			StartTime:  node.StartTime,
			EndTime:    node.EndTime,
		})
		if node.Status == constants.WorkFlowStatusError {
			break
		}
	}

	return resp, nil
}

func (mgr *WorkFlowManager) InitContext(ctx context.Context, flowId string, key string, value interface{}) error {
	meta, err := NewWorkFlowMeta(ctx, flowId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("build workflow meta by flow id %s failed %s", flowId, err.Error())
		return err
	}

	if err = meta.Context.SetData(key, value); err != nil {
		framework.LogWithContext(ctx).Errorf("set workflow context of flow id %s failed %s", flowId, err.Error())
		return err
	}

	meta.Restore()
	return nil
}

func (mgr *WorkFlowManager) Start(ctx context.Context, flowId string) error {
	framework.LogWithContext(ctx).Infof("Begin async start workflow Id %s", flowId)
	flow, err := models.GetWorkFlowReaderWriter().GetWorkFlow(ctx, flowId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get workflow by workflow Id %s, failed %s", flowId, err.Error())
		return errors.NewErrorf(errors.TIEM_WORKFLOW_QUERY_FAILED, err.Error(), err)
	}
	if flow.Finished() {
		framework.LogWithContext(ctx).Infof("workflow Id %s is finished", flowId)
		return errors.NewErrorf(errors.TIEM_WORKFLOW_START_FAILED, err.Error(), err)
	}
	return models.GetWorkFlowReaderWriter().UpdateWorkFlow(ctx, flowId, constants.WorkFlowStatusProcessing, "")
}

func (mgr *WorkFlowManager) Stop(ctx context.Context, flowId string) error {
	framework.LogWithContext(ctx).Infof("Begin stop workflow Id %s", flowId)
	flow, err := models.GetWorkFlowReaderWriter().GetWorkFlow(ctx, flowId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get workflow by workflow Id %s, failed %s", flowId, err.Error())
		return errors.NewErrorf(errors.TIEM_WORKFLOW_QUERY_FAILED, err.Error(), err)
	}
	if flow.Finished() {
		framework.LogWithContext(ctx).Infof("workflow Id %s is finished", flowId)
		return errors.NewErrorf(errors.TIEM_WORKFLOW_STOP_FAILED, err.Error(), err)
	}
	return models.GetWorkFlowReaderWriter().UpdateWorkFlow(ctx, flowId, constants.WorkFlowStatusStopped, "")
}

func (mgr *WorkFlowManager) Cancel(ctx context.Context, flowId string, reason string) error {
	framework.LogWithContext(ctx).Infof("Begin cancel workflow Id %s", flowId)
	flow, err := models.GetWorkFlowReaderWriter().GetWorkFlow(ctx, flowId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get workflow by workflow Id %s, failed %s", flowId, err.Error())
		return errors.NewErrorf(errors.TIEM_WORKFLOW_QUERY_FAILED, err.Error(), err)
	}
	if flow.Finished() || !flow.Stopped() {
		framework.LogWithContext(ctx).Infof("workflow Id %s is finished", flowId)
		return errors.NewErrorf(errors.TIEM_WORKFLOW_CANCEL_FAILED, err.Error(), err)
	}
	return models.GetWorkFlowReaderWriter().UpdateWorkFlow(ctx, flowId, constants.WorkFlowStatusCanceling, "")
}
