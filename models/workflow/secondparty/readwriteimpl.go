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
	"errors"
	"fmt"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"

	dbCommon "github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
)

type GormSecondPartyOperationReadWrite struct {
	dbCommon.GormDB
}

func NewGormSecondPartyOperationReadWrite(db *gorm.DB) *GormSecondPartyOperationReadWrite {
	m := &GormSecondPartyOperationReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (m *GormSecondPartyOperationReadWrite) Create(ctx context.Context, operationType OperationType,
	workFlowNodeID string) (*SecondPartyOperation, error) {
	if "" == workFlowNodeID || "" == string(operationType) {
		return nil, framework.NewTiEMErrorf(common.TIEM_PARAMETER_INVALID, "either workflownodeid(actual: %s) or "+
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
		return framework.NewTiEMErrorf(common.TIEM_PARAMETER_INVALID, "id is nil for %+v", updateTemplate)
	}

	return m.DB(ctx).Omit(Column_Type, Column_WorkFlowNodeID).
		Save(updateTemplate).Error
}

func (m *GormSecondPartyOperationReadWrite) Get(ctx context.Context, id string) (*SecondPartyOperation, error) {
	if "" == id {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	secondPartyOperation := &SecondPartyOperation{}
	err := m.DB(ctx).First(secondPartyOperation, "id = ?", id).Error

	if err != nil {
		return nil, framework.NewTiEMError(common.TIEM_SECOND_PARTY_OPERATION_NOT_FOUND, err.Error())
	} else {
		return secondPartyOperation, nil
	}
}

func (m *GormSecondPartyOperationReadWrite) QueryByWorkFlowNodeID(ctx context.Context,
	workFlowNodeID string) (secondPartyOperation *SecondPartyOperation, err error) {
	if "" == workFlowNodeID {
		return nil, framework.SimpleError(common.TIEM_PARAMETER_INVALID)
	}

	var secondPartyOperations []SecondPartyOperation
	err = m.DB(ctx).Model(&SecondPartyOperation{}).
		Where(fmt.Sprintf("%s = ?", Column_WorkFlowNodeID), workFlowNodeID).
		Find(&secondPartyOperations).Error

	if err != nil {
		return
	}

	errCt := 0
	errStatStr := ""
	processingCt := 0
	for _, operation := range secondPartyOperations {
		if operation.Status == OperationStatus_Finished {
			return &operation, nil
		}
		if operation.Status == OperationStatus_Error {
			errStatStr = operation.ErrorStr
			errCt++
			continue
		}
		if operation.Status == OperationStatus_Processing {
			processingCt++
			continue
		}
	}
	if len(secondPartyOperations) == 0 {
		err = errors.New("no match record was found")
		return
	}
	secondPartyOperation = &SecondPartyOperation{}
	if errCt >= len(secondPartyOperations) {
		secondPartyOperation.Status = OperationStatus_Error
		secondPartyOperation.ErrorStr = errStatStr
		return
	} else {
		if processingCt > 0 {
			secondPartyOperation.Status = OperationStatus_Processing
		} else {
			secondPartyOperation.Status = OperationStatus_Init
		}
		return
	}
}
