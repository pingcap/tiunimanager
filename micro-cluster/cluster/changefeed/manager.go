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
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/cluster/changefeed"
)

type ChangeFeedManager struct {
}

func NewChangeFeedManager() *ChangeFeedManager {
	return &ChangeFeedManager{}
}

func (p *ChangeFeedManager) Create(ctx context.Context, name string) (string, error) {
	result, err := models.GetChangeFeedReaderWriter().Create(ctx, &changefeed.ChangeFeedTask{
		Name: name,
		ClusterId: "111",
		Type: changefeed.DownstreamTypeMysql,
		StartTS: 0,
		Downstream: changefeed.MysqlDownstream{
			WorkerCount: 3,
		},
	})
	if err != nil {
		return "", framework.WrapError(common.TIEM_PARAMETER_INVALID, "create change feed task error", err)
	}

	return result.ID, nil
}
