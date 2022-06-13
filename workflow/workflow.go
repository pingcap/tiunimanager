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

package workflow

/*
import (
	"context"
	"sync"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/models"
)

// WorkFlowService workflow interface
type WorkFlowService interface {
	// RegisterWorkFlow
	// @Description: register workflow define
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowName
	// @Parameter flowDefine
	RegisterWorkFlow(ctx context.Context, flowName string, flowDefine *WorkFlowDefine)

	// GetWorkFlowDefine
	// @Description: get workflow define by flowName
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowName
	// @Return *WorkFlowDefine
	// @Return error
	GetWorkFlowDefine(ctx context.Context, flowName string) (*WorkFlowDefine, error)

	// CreateWorkFlow
	// @Description: create new workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter bizId
	// @Parameter bizType
	// @Parameter flowName
	// @Return *WorkFlowAggregation
	// @Return error
	CreateWorkFlow(ctx context.Context, bizId string, bizType string, flowName string) (*WorkFlowAggregation, error)

	// ListWorkFlows
	// @Description: list workflows by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryWorkFlowsResp
	// @Return total
	// @Return error
	ListWorkFlows(ctx context.Context, request message.QueryWorkFlowsReq) (message.QueryWorkFlowsResp, structs.Page, error)

	// DetailWorkFlow
	// @Description: create new workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter request
	// @Return message.QueryWorkFlowDetailResp
	// @Return error
	DetailWorkFlow(ctx context.Context, request message.QueryWorkFlowDetailReq) (message.QueryWorkFlowDetailResp, error)

	// AddContext
	// @Description: add flow context for workflow
	// @Receiver m
	// @Parameter flow
	// @Parameter key
	// @Parameter value
	AddContext(flow *WorkFlowAggregation, key string, value interface{})

	// AsyncStart
	// @Description: async start workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Return error
	AsyncStart(ctx context.Context, flow *WorkFlowAggregation) error

	// Start
	// @Description: sync start workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Return error
	Start(ctx context.Context, flow *WorkFlowAggregation) error

	// Destroy
	// @Description: destroy workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Parameter reason
	// @Return error
	Destroy(ctx context.Context, flow *WorkFlowAggregation, reason string) error

	// Complete
	// @Description: complete workflow with result
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Parameter success
	Complete(ctx context.Context, flow *WorkFlowAggregation, success bool)
}

type WorkFlowManager struct {
	flowDefineMap sync.Map
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
	return &WorkFlowManager{}
}

func MockWorkFlowService(service WorkFlowService) {
	workflowService = service
}

func (mgr *WorkFlowManager) RegisterWorkFlow(ctx context.Context, flowName string, flowDefine *WorkFlowDefine) {
	mgr.flowDefineMap.Store(flowName, flowDefine)
	framework.LogWithContext(ctx).Infof("Register WorkFlow %s success, definition: %+v", flowDefine.FlowName, flowDefine)
}

func (mgr *WorkFlowManager) GetWorkFlowDefine(ctx context.Context, flowName string) (*WorkFlowDefine, error) {
	flowDefine, exist := mgr.flowDefineMap.Load(flowName)
	if !exist {
		framework.LogWithContext(ctx).Errorf("WorkFlow %s not exist", flowName)
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_WORKFLOW_DEFINE_NOT_FOUND, "%s workflow definion not exist", flowName)
	}
	return flowDefine.(*WorkFlowDefine), nil
}

func (mgr *WorkFlowManager) CreateWorkFlow(ctx context.Context, bizId string, bizType string, flowName string) (*WorkFlowAggregation, error) {
	flowDefine, exist := mgr.flowDefineMap.Load(flowName)
	if !exist {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_WORKFLOW_DEFINE_NOT_FOUND, "%s workflow definion not exist", flowName)
	}

	flow, err := createFlowWork(ctx, bizId, bizType, flowDefine.(*WorkFlowDefine))
	if err != nil {
		return nil, errors.WrapError(errors.TIUNIMANAGER_WORKFLOW_CREATE_FAILED, err.Error(), err)
	}
	return flow, nil
}

func (mgr *WorkFlowManager) ListWorkFlows(ctx context.Context, request message.QueryWorkFlowsReq) (resp message.QueryWorkFlowsResp, page structs.Page, err error) {
	flows, total, err := models.GetWorkFlowReaderWriter().QueryWorkFlows(ctx, request.BizID, request.BizType, request.FlowName, request.Status, request.Page, request.PageSize)
	if err != nil {
		return resp, page, errors.WrapError(errors.TIUNIMANAGER_WORKFLOW_QUERY_FAILED, err.Error(), err)
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
		return resp, errors.WrapError(errors.TIUNIMANAGER_WORKFLOW_DETAIL_FAILED, err.Error(), err)
	}

	define, err := mgr.GetWorkFlowDefine(ctx, flow.Name)
	if err != nil {
		return resp, errors.WrapError(errors.TIUNIMANAGER_WORKFLOW_DEFINE_NOT_FOUND, err.Error(), err)
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

func (mgr *WorkFlowManager) AddContext(flow *WorkFlowAggregation, key string, value interface{}) {
	flow.addContext(key, value)
}

func (mgr *WorkFlowManager) AsyncStart(ctx context.Context, flow *WorkFlowAggregation) error {
	framework.LogWithContext(ctx).Infof("Begin async start workflow name %s, workflowId %s, bizId: %s", flow.Flow.Name, flow.Flow.ID, flow.Flow.BizID)
	flow.asyncStart(ctx)
	return nil
}

func (mgr *WorkFlowManager) Start(ctx context.Context, flow *WorkFlowAggregation) error {
	framework.LogWithContext(ctx).Infof("Begin sync start workflow name %s, workflowId %s, bizId: %s", flow.Flow.Name, flow.Flow.ID, flow.Flow.BizID)
	flow.Context.Context = ctx
	flow.start(ctx)
	return nil
}

func (mgr *WorkFlowManager) Destroy(ctx context.Context, flow *WorkFlowAggregation, reason string) error {
	framework.LogWithContext(ctx).Infof("Begin destroy workflow name %s, workflowId %s, bizId: %s", flow.Flow.Name, flow.Flow.ID, flow.Flow.BizID)
	flow.destroy(ctx, reason)
	return nil
}

func (mgr *WorkFlowManager) Complete(ctx context.Context, flow *WorkFlowAggregation, success bool) {
	framework.LogWithContext(ctx).Infof("Begin complete workflow name %s, workflowId %s, bizId: %s", flow.Flow.Name, flow.Flow.ID, flow.Flow.BizID)
	flow.complete(success)
}
*/
