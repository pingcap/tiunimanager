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
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// Create create a change feed task
// @Summary create a change feed task
// @Description create a change feed task
// @Tags change feed
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param changeFeedTask body cluster.CreateChangeFeedTaskReq true "change feed task request"
// @Success 200 {object} controller.CommonResult{data=cluster.CreateChangeFeedTaskResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/ [post]
func Create(c *gin.Context) {
	var req cluster.CreateChangeFeedTaskReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c, req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask, &cluster.CreateChangeFeedTaskResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Query
// @Summary query change feed tasks of a cluster
// @Description query change feed tasks of a cluster
// @Tags change feed
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryReq query cluster.QueryChangeFeedTaskReq true "change feed tasks query condition"
// @Success 200 {object} controller.ResultWithPage{data=[]cluster.QueryChangeFeedTaskResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/ [get]
func Query(c *gin.Context) {
	var req cluster.QueryChangeFeedTaskReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c, req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryChangeFeedTasks, make([][]cluster.QueryChangeFeedTaskResp, 0),
			requestBody,
			controller.DefaultTimeout)
	}
}

const paramNameOfChangeFeedTaskId = "changeFeedTaskId"

// Detail get change feed detail
// @Summary  get change feed detail
// @Description get change feed detail
// @Tags change feed
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param changeFeedTaskId path string true "changeFeedTaskId"
// @Success 200 {object} controller.CommonResult{data=cluster.DetailChangeFeedTaskResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId}/ [get]
func Detail(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, cluster.DetailChangeFeedTaskReq{
		ID: c.Param(paramNameOfChangeFeedTaskId),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask, &cluster.DetailChangeFeedTaskResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Pause pause a change feed task
// @Summary  pause a change feed task
// @Description pause a change feed task
// @Tags change feed
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param changeFeedTaskId path string true "changeFeedTaskId"
// @Success 200 {object} controller.CommonResult{data=cluster.PauseChangeFeedTaskResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId}/pause [post]
func Pause(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, cluster.PauseChangeFeedTaskReq{
		ID: c.Param(paramNameOfChangeFeedTaskId),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask, &cluster.PauseChangeFeedTaskResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Resume resume a change feed task
// @Summary  resume a change feed task
// @Description resume a change feed task
// @Tags change feed
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param changeFeedTaskId path string true "changeFeedTaskId"
// @Success 200 {object} controller.CommonResult{data=cluster.ResumeChangeFeedTaskResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId}/resume [post]
func Resume(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, cluster.ResumeChangeFeedTaskReq{
		Id: c.Param(paramNameOfChangeFeedTaskId),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask, &cluster.ResumeChangeFeedTaskResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Update update a change feed
// @Summary  resume a change feed
// @Description resume a change feed
// @Tags change feed
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param changeFeedTaskId path string true "changeFeedTaskId"
// @Param task body cluster.UpdateChangeFeedTaskReq true "change feed task"
// @Success 200 {object} controller.CommonResult{data=cluster.UpdateChangeFeedTaskResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId}/update [post]
func Update(c *gin.Context) {
	var req cluster.UpdateChangeFeedTaskReq

	if requestBody, ok := controller.HandleJsonRequestFromBody(c,
		req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*cluster.UpdateChangeFeedTaskReq).ID = c.Param(paramNameOfChangeFeedTaskId)
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask, &cluster.UpdateChangeFeedTaskResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Delete delete a change feed task
// @Summary  delete a change feed task
// @Description delete a change feed task
// @Tags change feed
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param changeFeedTaskId path string true "changeFeedTaskId"
// @Success 200 {object} controller.CommonResult{data=cluster.DeleteChangeFeedTaskResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId} [delete]
func Delete(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, cluster.DeleteChangeFeedTaskReq{
		ID: c.Param(paramNameOfChangeFeedTaskId),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask, &cluster.DeleteChangeFeedTaskResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Downstream
// @Summary unused, just display downstream config
// @Description show display config
// @Tags change feed
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param tidb body cluster.TiDBDownstream true "tidb"
// @Param mysql body cluster.MysqlDownstream true "mysql"
// @Param kafka body cluster.KafkaDownstream true "kafka"
// @Success 200 {object} controller.CommonResult
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /downstream/ [delete]
func Downstream(c *gin.Context) {
}
