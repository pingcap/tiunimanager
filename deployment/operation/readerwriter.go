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
 * @File: readerwriter
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/11
*******************************************************************************/

package operation

import "context"

type ReaderWriter interface {
	// Create
	// @Description: create new operation record
	// @Receiver m
	// @Parameter ctx
	// @Parameter operationType
	// @Parameter workFlowID
	// @Return *Operation
	// @Return error
	Create(ctx context.Context, operationType string, workFlowID string) (*Operation, error)

	// Update
	// @Description: update operation record
	// @Receiver m
	// @Parameter ctx
	// @Parameter updateTemplate skip fields below : Type、WorkFlowID
	// @Return error
	Update(ctx context.Context, updateTemplate *Operation) error

	// Get
	// @Description: get operation record for given operation id
	// @Receiver m
	// @Parameter ctx
	// @Parameter id
	// @Return *SecondPartyOperation
	// @Return error
	Get(ctx context.Context, id string) (*Operation, error)
}
