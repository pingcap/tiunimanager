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
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
)

type FlowReadWrite struct {
}

func NewFlowReadWrite() *FlowReadWrite {
	m := new(FlowReadWrite)
	return m
}

func (m *FlowReadWrite) CreateWorkFlow(ctx context.Context, flow *WorkFlow) (*WorkFlow, error) {
	return flow, models.DB(ctx).Create(flow).Error
}

func (m *FlowReadWrite) UpdateWorkFlowStatus(ctx context.Context, flowId string, status string) (err error) {
	if "" == flowId || "" == status {
		return framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	flow := &WorkFlow{}
	err = models.DB(ctx).First(flow, "id = ?", flowId).Error
	if err != nil {
		return framework.SimpleError(common.TIEM_FLOW_NOT_FOUND)
	}

	return models.DB(ctx).Model(flow).
		Update("status", status).Error
}

func (m *FlowReadWrite) GetWorkFlow(ctx context.Context, flowId string) (flow *WorkFlow, err error) {
	if "" == flowId {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	flow = &WorkFlow{}
	err = models.DB(ctx).First(flow, "id = ?", flowId).Error
	if err != nil {
		return nil, framework.SimpleError(common.TIEM_FLOW_NOT_FOUND)
	}
	return flow, nil
}

func (m *FlowReadWrite) QueryWorkFlows(ctx context.Context, bizId, fuzzyName, status string, page int, pageSize int) (flows []*WorkFlow, total int64, err error) {
	flows = make([]*WorkFlow, pageSize)
	query := models.DB(ctx).Model(&WorkFlow{})
	if bizId != "" {
		query = query.Where("biz_id = ?", bizId)
	}
	if fuzzyName != "" {
		query = query.Where("name like '%" + fuzzyName + "%'")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err = query.Count(&total).Order("id desc").Offset(pageSize * (page - 1)).Limit(pageSize).Find(&flows).Error
	return flows, total, err
}

func (m *FlowReadWrite) CreateWorkFlowNode(ctx context.Context, node *WorkFlowNode) (*WorkFlowNode, error) {
	return node, models.DB(ctx).Create(node).Error
}

func (m *FlowReadWrite) UpdateWorkFlowNode(ctx context.Context, node *WorkFlowNode) (err error) {
	return models.DB(ctx).Model(node).Where("id = ?", node.ID).Updates(node).Error
}

func (m *FlowReadWrite) GetWorkFlowNode(ctx context.Context, nodeId string) (node *WorkFlowNode, err error) {
	node = &WorkFlowNode{}
	return node, models.DB(ctx).Model(node).Where("id = ?", nodeId).First(node).Error
}

func (m *FlowReadWrite) DetailWorkFlow(ctx context.Context, flowId string) (flow *WorkFlow, nodes []*WorkFlowNode, err error) {
	if "" == flowId {
		return nil, nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	flow = &WorkFlow{}
	err = models.DB(ctx).First(flow, "id = ?", flowId).Error
	if err != nil {
		return nil, nil, framework.SimpleError(common.TIEM_FLOW_NOT_FOUND)
	}

	err = models.DB(ctx).Where("parent_id = ?", flowId).Find(&nodes).Error
	if err != nil {
		return nil, nil, err
	}
	return flow, nodes, nil
}
