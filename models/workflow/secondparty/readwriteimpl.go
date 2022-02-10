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

/*******************************************************************************
 * @File: ReadWriteImpl
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10
*******************************************************************************/

package secondparty

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type GormSecondPartyOperationReadWrite struct {
	common.GormDB
}

func NewGormSecondPartyOperationReadWrite(db *gorm.DB) *GormSecondPartyOperationReadWrite {
	m := &GormSecondPartyOperationReadWrite{
		common.WrapDB(db),
	}
	return m
}

func (m *GormSecondPartyOperationReadWrite) Create(ctx context.Context, operationType OperationType,
	workFlowNodeID string) (*SecondPartyOperation, error) {
	if "" == workFlowNodeID || "" == string(operationType) {
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "either workflownodeid(actual: %s) or "+
			"type(actual: %s) is nil", workFlowNodeID, operationType)
	}

	secondPartyOperation := &SecondPartyOperation{
		Type:           operationType,
		WorkFlowNodeID: workFlowNodeID,
	}

	return secondPartyOperation, m.DB(ctx).Create(secondPartyOperation).Error
}

func (m *GormSecondPartyOperationReadWrite) Update(ctx context.Context, updateTemplate *SecondPartyOperation) error {
	if "" == updateTemplate.ID {
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "id is nil for %+v", updateTemplate)
	}

	return m.DB(ctx).Omit(ColumnType, ColumnWorkFlowNodeID).
		Save(updateTemplate).Error
}

func (m *GormSecondPartyOperationReadWrite) Get(ctx context.Context, id string) (*SecondPartyOperation, error) {
	if "" == id {
		return nil, errors.NewError(errors.TIEM_PARAMETER_INVALID, "id required")
	}

	secondPartyOperation := &SecondPartyOperation{}
	err := m.DB(ctx).First(secondPartyOperation, "id = ?", id).Error

	if err != nil {
		return nil, common.WrapDBError(err)
	} else {
		return secondPartyOperation, nil
	}
}

func (m *GormSecondPartyOperationReadWrite) QueryByWorkFlowNodeID(ctx context.Context,
	workFlowNodeID string) (secondPartyOperation *SecondPartyOperation, err error) {
	if "" == workFlowNodeID {
		return nil, errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "node id is required")
	}

	var secondPartyOperations []SecondPartyOperation
	err = m.DB(ctx).Model(&SecondPartyOperation{}).
		Where(fmt.Sprintf("%s = ?", ColumnWorkFlowNodeID), workFlowNodeID).
		Find(&secondPartyOperations).Error

	if err != nil {
		return
	}

	errCt := 0
	errStatStr := ""
	processingCt := 0
	for _, operation := range secondPartyOperations {
		if operation.Status == OperationStatusFinished {
			return &operation, nil
		}
		if operation.Status == OperationStatusError {
			errStatStr = operation.ErrorStr
			errCt++
			continue
		}
		if operation.Status == OperationStatusProcessing {
			processingCt++
			continue
		}
	}
	if len(secondPartyOperations) == 0 {
		err = errors.NewError(errors.TIEM_PARAMETER_INVALID, "no match record was found")
		return
	}
	secondPartyOperation = &SecondPartyOperation{}
	if errCt >= len(secondPartyOperations) {
		secondPartyOperation.Status = OperationStatusError
		secondPartyOperation.ErrorStr = errStatStr
		return
	} else {
		if processingCt > 0 {
			secondPartyOperation.Status = OperationStatusProcessing
		} else {
			secondPartyOperation.Status = OperationStatusInit
		}
		return
	}
}
