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
	"sync"
)

type FlowInterface interface {
	RegisterWorkFlow(ctx context.Context, defineName string, flowDefine *FlowWorkDefine)
	GetWorkFlowDefine(ctx context.Context, defineName string) (*FlowWorkDefine, error)
	CreateWorkFlow(ctx context.Context, bizId string, defineName string) (*FlowWorkAggregation, error)
	ListWorkFlows(ctx context.Context, bizId string, status string) ([]*FlowWorkEntity, error)
	DetailWorkFlow(ctx context.Context, flowId string) (*FlowWorkAggregation, error)
	AddContext(flow *FlowWorkAggregation, key string, value interface{})
	AsyncStart(flow *FlowWorkAggregation) error
	Start(flow *FlowWorkAggregation) error
	Destroy(flow *FlowWorkAggregation, reason string) error
	Complete(flow *FlowWorkAggregation, success bool)
}

type FlowManager struct {
	flowDefineMap sync.Map
	//todo db interface
}

var manager *FlowManager

func GetFlowManager() *FlowManager {
	if manager == nil {
		manager = &FlowManager{}
	}
	return manager
}

func (mgr *FlowManager) RegisterWorkFlow(ctx context.Context, defineName string, flowDefine *FlowWorkDefine) {
	mgr.flowDefineMap.Store(defineName, flowDefine)
	return
}

func (mgr *FlowManager) GetWorkFlowDefine(ctx context.Context, defineName string) (*FlowWorkDefine, error) {
	flowDefine, exist := mgr.flowDefineMap.Load(defineName)
	if !exist {
		return nil, fmt.Errorf("%s workflow definion not exist", defineName)
	}
	return flowDefine.(*FlowWorkDefine), nil
}

func (mgr *FlowManager) CreateWorkFlow(ctx context.Context, bizId string, defineName string) (*FlowWorkAggregation, error) {
	flowDefine, exist := mgr.flowDefineMap.Load(defineName)
	if !exist {
		return nil, fmt.Errorf("%s workflow definion not exist", defineName)
	}

	flow, err := createFlowWork(ctx, bizId, flowDefine.(*FlowWorkDefine))
	return flow, err
}

func (mgr *FlowManager) ListWorkFlows(ctx context.Context, bizId string, status string) ([]*FlowWorkEntity, error) {
	// todo: call db interface load FlowWorkEntity
	return nil, nil
}

func (mgr *FlowManager) DetailWorkFlow(ctx context.Context, flowId string) (*FlowWorkAggregation, error) {
	// todo: call db interface load FlowWorkAggregation
	return nil, nil
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
