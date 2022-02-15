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

/*******************************************************************************
 * @File: cdc_manager_api.go
 * @Description: interface of cdc_manager_api
 * @Author: hansen@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/15 11:30
*******************************************************************************/

package switchover

import (
	"context"

	"github.com/pingcap-inc/tiem/message/cluster"
)

type CDCManagerAPI interface {
	Create(ctx context.Context, request cluster.CreateChangeFeedTaskReq) (resp cluster.CreateChangeFeedTaskResp, err error)
	Delete(ctx context.Context, request cluster.DeleteChangeFeedTaskReq) (resp cluster.DeleteChangeFeedTaskResp, err error)
	Pause(ctx context.Context, request cluster.PauseChangeFeedTaskReq) (resp cluster.PauseChangeFeedTaskResp, err error)
	Resume(ctx context.Context, request cluster.ResumeChangeFeedTaskReq) (resp cluster.ResumeChangeFeedTaskResp, err error)
	Detail(ctx context.Context, request cluster.DetailChangeFeedTaskReq) (resp cluster.DetailChangeFeedTaskResp, err error)
	Query(ctx context.Context, request cluster.QueryChangeFeedTaskReq) (resps []cluster.QueryChangeFeedTaskResp, total int, err error)
}
