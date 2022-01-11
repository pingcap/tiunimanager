/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

/*******************************************************************************
 * @File: readwriteimpl
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/11
*******************************************************************************/

package operation

import (
	"context"

	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type GormOperationReadWrite struct {
	common.GormDB
}

func NewGormOperationReadWrite(db *gorm.DB) *GormOperationReadWrite {
	m := &GormOperationReadWrite{
		common.WrapDB(db),
	}
	return m
}

func (m *GormOperationReadWrite) Create(ctx context.Context, operationType string,
	workFlowID string) (*Operation, error) {
	if "" == workFlowID || "" == string(operationType) {
		return nil, errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "either workflowid(actual: %s) or "+
			"type(actual: %s) is empty", workFlowID, operationType)
	}

	Operation := &Operation{
		Type:       operationType,
		WorkFlowID: workFlowID,
		Status:     Init,
	}

	return Operation, m.DB(ctx).Create(Operation).Error
}

func (m *GormOperationReadWrite) Update(ctx context.Context, updateTemplate *Operation) error {
	if "" == updateTemplate.ID {
		return errors.NewEMErrorf(errors.TIEM_PARAMETER_INVALID, "id is empty for %+v", updateTemplate)
	}

	return m.DB(ctx).Omit("type", "work_flow_node_id").
		Save(updateTemplate).Error
}

func (m *GormOperationReadWrite) Get(ctx context.Context, id string) (*Operation, error) {
	if "" == id {
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "id is empty")
	}

	Operation := &Operation{}
	err := m.DB(ctx).First(Operation, "id = ?", id).Error

	if err != nil {
		return nil, common.WrapDBError(err)
	} else {
		return Operation, nil
	}
}
