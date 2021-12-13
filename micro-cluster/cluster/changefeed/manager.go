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
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
	dbCommon "github.com/pingcap-inc/tiem/models/common"
)

type Manager struct {}

func NewManager() *Manager {
	return &Manager{}
}

// Create
// @Description:
// @Receiver p
// @Parameter ctx
// @Parameter name
// @return string ID of ChangeFeedTask
// @return error
func (p *Manager) Create(ctx context.Context, request cluster.CreateChangeFeedTaskReq) (cluster.CreateChangeFeedTaskResp, error) {
	task := &changefeed.ChangeFeedTask{
		Entity: dbCommon.Entity{
			TenantId: "1111",
		},
		Name: request.Name,
		Type: constants.DownstreamType(request.DownstreamType),
		Downstream: changefeed.MysqlDownstream{
			WorkerCount: 3,
		},
	}

	task, err := models.GetChangeFeedReaderWriter().Create(ctx, task)

	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to create change feed task, %s", err.Error())
		return cluster.CreateChangeFeedTaskResp{}, framework.WrapError(common.TIEM_CHANGE_FEED_CREATE_ERROR, "failed to create change feed task", err)
	}

	return cluster.CreateChangeFeedTaskResp{
		ID: task.ID,
	}, nil
}

func (p *Manager) Delete(ctx context.Context, id string) error {

	err := models.GetChangeFeedReaderWriter().Delete(ctx, id)

	if err != nil {
		framework.LogWithContext(ctx).Errorf("failed to create change feed task, %s", err.Error())
		return framework.SimpleError(common.TIEM_CHANGE_FEED_CREATE_ERROR)
	}

	return nil
}