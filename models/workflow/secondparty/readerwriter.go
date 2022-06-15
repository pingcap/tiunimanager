/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
 * @File: ReaderWriter
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/9
*******************************************************************************/

package secondparty

import (
	"context"
)

type ReaderWriter interface {
	// Create
	// @Description: create new second party operation record
	// @Receiver m
	// @Parameter ctx
	// @Parameter operationType
	// @Parameter workFlowNodeID
	// @Return *SecondPartyOperation
	// @Return error
	Create(ctx context.Context, operationType OperationType, workFlowNodeID string) (*SecondPartyOperation, error)

	// Update
	// @Description: update second party operation record
	// @Receiver m
	// @Parameter ctx
	// @Parameter updateTemplate skip fields below : Type、WorkFlowNodeID
	// @Return error
	Update(ctx context.Context, updateTemplate *SecondPartyOperation) error

	// Get
	// @Description: get second party operation record for given record id
	// @Receiver m
	// @Parameter ctx
	// @Parameter id
	// @Return *SecondPartyOperation
	// @Return error
	Get(ctx context.Context, id string) (*SecondPartyOperation, error)

	// QueryByWorkFlowNodeID
	// @Description: get second party operation record for given WorkFlowNode ID
	// @Receiver m
	// @Parameter ctx
	// @Parameter workFlowNodeID
	// @Return *SecondPartyOperation
	// @Return error
	QueryByWorkFlowNodeID(ctx context.Context, workFlowNodeID string) (*SecondPartyOperation, error)
}
