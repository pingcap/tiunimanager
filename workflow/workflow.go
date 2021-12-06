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

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/workflow"
	"sync"
)

type FlowInterface interface {
	// RegisterWorkFlow
	// @Description: register workflow define
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowName
	// @Parameter flowDefine
	RegisterWorkFlow(ctx context.Context, flowName string, flowDefine *FlowWorkDefine)

	// GetWorkFlowDefine
	// @Description: get workflow define by flowName
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowName
	// @Return *FlowWorkDefine
	// @Return error
	GetWorkFlowDefine(ctx context.Context, flowName string) (*FlowWorkDefine, error)

	// CreateWorkFlow
	// @Description: create new workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter bizId
	// @Parameter flowName
	// @Return *FlowWorkAggregation
	// @Return error
	CreateWorkFlow(ctx context.Context, bizId string, flowName string) (*FlowWorkAggregation, error)

	// ListWorkFlows
	// @Description: list workflows by condition
	// @Receiver m
	// @Parameter ctx
	// @Parameter bizId
	// @Parameter fuzzyName
	// @Parameter status
	// @Parameter page
	// @Parameter pageSize
	// @Return []*workflow.WorkFlow
	// @Return total
	// @Return error
	ListWorkFlows(ctx context.Context, bizId string, fuzzyName string, status string, page, pageSize int) ([]*workflow.WorkFlow, int64, error)

	// DetailWorkFlow
	// @Description: create new workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flowId
	// @Return *FlowWorkAggregation
	// @Return error
	DetailWorkFlow(ctx context.Context, flowId string) (*FlowWorkAggregation, error)

	// AddContext
	// @Description: add flow context for workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Parameter key
	// @Parameter value
	AddContext(flow *FlowWorkAggregation, key string, value interface{})

	// AsyncStart
	// @Description: async start workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Return error
	AsyncStart(flow *FlowWorkAggregation) error

	// Start
	// @Description: sync start workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Return error
	Start(flow *FlowWorkAggregation) error

	// Destroy
	// @Description: destroy workflow
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Parameter reason
	// @Return error
	Destroy(flow *FlowWorkAggregation, reason string) error

	// Complete
	// @Description: complete workflow with result
	// @Receiver m
	// @Parameter ctx
	// @Parameter flow
	// @Parameter success
	Complete(flow *FlowWorkAggregation, success bool)
}

type FlowManager struct {
	flowDefineMap sync.Map
}

var manager *FlowManager

func GetFlowManager() *FlowManager {
	if manager == nil {
		manager = &FlowManager{}
	}
	return manager
}

func (mgr *FlowManager) RegisterWorkFlow(ctx context.Context, flowName string, flowDefine *FlowWorkDefine) {
	mgr.flowDefineMap.Store(flowName, flowDefine)
	framework.LogWithContext(ctx).Infof("Register WorkFlow %s success, definition: %+v", flowName, flowDefine)
	return
}

func (mgr *FlowManager) GetWorkFlowDefine(ctx context.Context, flowName string) (*FlowWorkDefine, error) {
	flowDefine, exist := mgr.flowDefineMap.Load(flowName)
	if !exist {
		framework.LogWithContext(ctx).Errorf("WorkFlow %s not exist", flowName)
		return nil, fmt.Errorf("%s workflow definion not exist", flowName)
	}
	return flowDefine.(*FlowWorkDefine), nil
}

func (mgr *FlowManager) CreateWorkFlow(ctx context.Context, bizId string, flowName string) (*FlowWorkAggregation, error) {
	flowDefine, exist := mgr.flowDefineMap.Load(flowName)
	if !exist {
		return nil, fmt.Errorf("%s workflow definion not exist", flowName)
	}

	flow, err := createFlowWork(ctx, bizId, flowDefine.(*FlowWorkDefine))
	return flow, err
}

func (mgr *FlowManager) ListWorkFlows(ctx context.Context, bizId string, fuzzyName string, status string, page, pageSize int) ([]*workflow.WorkFlow, int64, error) {
	return models.GetWorkFlowReaderWriter().QueryWorkFlows(ctx, bizId, fuzzyName, status, page, pageSize)
}

func (mgr *FlowManager) DetailWorkFlow(ctx context.Context, flowId string) (*FlowWorkAggregation, error) {
	flow, nodes, err := models.GetWorkFlowReaderWriter().QueryDetailWorkFlow(ctx, flowId)
	if err != nil {
		return nil, err
	}

	define, err := mgr.GetWorkFlowDefine(ctx, flow.Name)
	if err != nil {
		return nil, err
	}

	return &FlowWorkAggregation{
		FlowWork: flow,
		Define:   define,
		Nodes:    nodes,
	}, nil
}

func (mgr *FlowManager) AddContext(flow *FlowWorkAggregation, key string, value interface{}) {
	flow.addContext(key, value)
	return
}

func (mgr *FlowManager) AsyncStart(flow *FlowWorkAggregation) error {
	flow.asyncStart()
	return nil
}

func (mgr *FlowManager) Start(flow *FlowWorkAggregation) error {
	flow.start()
	return nil
}

func (mgr *FlowManager) Destroy(flow *FlowWorkAggregation, reason string) error {
	flow.destroy(reason)
	return nil
}

func (mgr *FlowManager) Complete(flow *FlowWorkAggregation, success bool) {
	flow.complete(success)
	return
}
