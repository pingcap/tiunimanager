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

import "github.com/pingcap-inc/tiem/micro-api/controller"

type CreateReq struct {
	ChangeFeedTask
}

type CreateResp struct {
	ID string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
}

type QueryReq struct {
	ClusterId string `json:"clusterId" form:"clusterId" example:"CLUSTER_ID_IN_TIEM__22"`
	controller.Page
}

type QueryResp struct {
	ChangeFeedTask
}

type DetailReq struct {
	ID string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
}

type DetailResp struct {
	ChangeFeedTask
}

type PauseReq struct {
	ID string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
}

type PauseResp struct {
	Status int `json:"status" form:"status" example:"1" enums:"0,1,2,3,4,5"`
}

type ResumeReq struct {
	Id string `json:"id" form:"id" example:"CLUSTER_ID_IN_TIEM__22"`
}

type ResumeResp struct {
	Status int `json:"status" form:"status" example:"1" enums:"0,1,2,3,4,5"`
}

type UpdateReq struct {
	ChangeFeedTask
}

type UpdateResp struct {
	Status int `json:"status" form:"status" example:"1" enums:"0,1,2,3,4,5"`
}

type DeleteReq struct {
	ID string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
}

type DeleteResp struct {
	ID     string `json:"id" form:"id" example:"TASK_ID_IN_TIEM____22"`
	Status int    `json:"status" form:"status" example:"1" enums:"0,1,2,3,4,5"`
}

type RpcRequest struct {
	requestBody string  //
}

type ClusterOperateRequest struct {
	Operate string
	CommandContent interface{}
}

type PauseCommand struct {
	whatever string
}

type ResumeCommand struct {
	whatever string
}
