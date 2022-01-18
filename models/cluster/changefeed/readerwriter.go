/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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
 ******************************************************************************/

package changefeed

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
)

type ReaderWriter interface {
	// Create
	// @Description: create a change feed task with default status lock
	// @Receiver m
	// @Parameter ctx
	// @Parameter task
	// @return *ChangeFeedTask
	// @return error
	Create(ctx context.Context, task *ChangeFeedTask) (*ChangeFeedTask, error)

	// Delete
	// @Description: delete a change feed
	// @Receiver m
	// @Parameter ctx
	// @Parameter taskId
	// @return err if task non-existent
	Delete(ctx context.Context, taskId string) (err error)

	// Get
	// @Description: get from id
	// @Receiver m
	// @Parameter ctx
	// @Parameter taskId
	// @return *ChangeFeedTask
	// @return error if task non-existent
	Get(ctx context.Context, taskId string) (*ChangeFeedTask, error)

	// QueryByClusterId
	// @Description: query tasks
	// @Receiver m
	// @Parameter ctx
	// @Parameter clusterId
	// @Parameter offset
	// @Parameter length
	// @return tasks
	// @return total
	// @return err
	QueryByClusterId(ctx context.Context, clusterId string, offset int, length int) (tasks []*ChangeFeedTask, total int64, err error)

	//
    // Query
    // @Description:
    // @param ctx
    // @param clusterId
    // @param taskType
    // @param status
    // @param offset
    // @param length
    // @return tasks
    // @return total
    // @return err
    //
	Query(ctx context.Context, clusterId string, taskType []constants.DownstreamType, status []constants.ChangeFeedStatus, offset int, length int) (tasks []*ChangeFeedTask, total int64, err error)

	// LockStatus
	// @Description: block changing status until task unlocked or lock expired
	// @Receiver m
	// @Parameter ctx
	// @Parameter taskId
	// @return error if task non-existent or locked
	LockStatus(ctx context.Context, taskId string) error

	// UnlockStatus
	// @Description: update status to target and release status lock
	// @Receiver m
	// @Parameter ctx
	// @Parameter taskId
	// @Parameter targetStatus
	// @return error if task non-existent or unlocked
	UnlockStatus(ctx context.Context, taskId string, targetStatus constants.ChangeFeedStatus) error

	// UpdateConfig
	// @Description: update task config with a template
	// @Receiver m
	// @Parameter ctx
	// @Parameter updateTemplate skip fields below : ChangeFeedStatus、StatusLock、ClusterId, StartTS
	// @return error if task non-existent
	UpdateConfig(ctx context.Context, updateTemplate *ChangeFeedTask) error
}
