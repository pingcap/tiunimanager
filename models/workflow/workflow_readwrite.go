/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/library/framework"
	dbCommon "github.com/pingcap/tiunimanager/models/common"
	"gorm.io/gorm"
)

type WorkFlowReadWrite struct {
	dbCommon.GormDB
}

func NewFlowReadWrite(db *gorm.DB) *WorkFlowReadWrite {
	m := &WorkFlowReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (m *WorkFlowReadWrite) CreateWorkFlow(ctx context.Context, flow *WorkFlow) (*WorkFlow, error) {
	return flow, m.DB(ctx).Create(flow).Error
}

func (m *WorkFlowReadWrite) UpdateWorkFlow(ctx context.Context, flowId string, status string, flowContext string) (err error) {
	if "" == flowId || "" == status {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "")
	}

	flow := &WorkFlow{}
	err = m.DB(ctx).First(flow, "id = ?", flowId).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_FLOW_NOT_FOUND, "flow %s not found", flowId)
	}

	db := m.DB(ctx).Model(flow)
	if "" != status {
		db.Update("status", status)
	}
	if "" != flowContext {
		db.Update("context", flowContext)
	}

	return db.Error
}

func (m *WorkFlowReadWrite) GetWorkFlow(ctx context.Context, flowId string) (flow *WorkFlow, err error) {
	if "" == flowId {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "flow id is required")
	}

	flow = &WorkFlow{}
	err = m.DB(ctx).First(flow, "id = ?", flowId).Error
	if err != nil {
		return nil, errors.NewErrorf(errors.TIUNIMANAGER_FLOW_NOT_FOUND, "flow %s not found", flowId)
	}
	return flow, nil
}

func (m *WorkFlowReadWrite) QueryWorkFlows(ctx context.Context, bizId, bizType, fuzzyName, status string, page int, pageSize int) (flows []*WorkFlow, total int64, err error) {
	flows = make([]*WorkFlow, 0)
	query := m.DB(ctx).Model(&WorkFlow{})
	if bizId != "" {
		query = query.Where("biz_id = ?", bizId)
	}
	if bizType != "" {
		query = query.Where("biz_type = ?", bizType)
	}
	if fuzzyName != "" {
		query = query.Where("name like '%" + fuzzyName + "%'")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err = query.Count(&total).Order("created_at desc").Offset(pageSize * (page - 1)).Limit(pageSize).Find(&flows).Error
	return flows, total, err
}

func (m *WorkFlowReadWrite) CreateWorkFlowNode(ctx context.Context, node *WorkFlowNode) (*WorkFlowNode, error) {
	return node, m.DB(ctx).Create(node).Error
}

func (m *WorkFlowReadWrite) UpdateWorkFlowNode(ctx context.Context, node *WorkFlowNode) (err error) {
	return m.DB(ctx).Model(node).Where("id = ?", node.ID).Updates(node).Error
}

func (m *WorkFlowReadWrite) GetWorkFlowNode(ctx context.Context, nodeId string) (node *WorkFlowNode, err error) {
	node = &WorkFlowNode{}
	return node, m.DB(ctx).Model(node).Where("id = ?", nodeId).First(node).Error
}

func (m *WorkFlowReadWrite) UpdateWorkFlowDetail(ctx context.Context, flow *WorkFlow, nodes []*WorkFlowNode) (err error) {
	return m.DB(ctx).Transaction(func(tx *gorm.DB) error {
		err = m.UpdateWorkFlow(dbCommon.CtxWithTransaction(ctx, tx), flow.ID, flow.Status, flow.Context)
		if err != nil {
			framework.LogWithContext(ctx).Errorf("update workflow %+v failed %s", flow, err.Error())
			tx.Rollback()
			return err
		}
		for _, node := range nodes {
			err = m.UpdateWorkFlowNode(dbCommon.CtxWithTransaction(ctx, tx), node)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("update workflow node %+v failed %s", node, err.Error())
				tx.Rollback()
				return err
			}
		}
		return nil
	})
}

func (m *WorkFlowReadWrite) QueryDetailWorkFlow(ctx context.Context, flowId string) (flow *WorkFlow, nodes []*WorkFlowNode, err error) {
	if "" == flowId {
		return nil, nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "empty flow id")
	}

	flow = &WorkFlow{}
	err = m.DB(ctx).First(flow, "id = ?", flowId).Error
	if err != nil {
		return nil, nil, errors.NewErrorf(errors.TIUNIMANAGER_FLOW_NOT_FOUND, "flow %s not found", flowId)
	}

	err = m.DB(ctx).Where("parent_id = ?", flowId).Find(&nodes).Error
	if err != nil && gorm.ErrRecordNotFound != err {
		return nil, nil, err
	}
	return flow, nodes, nil
}
