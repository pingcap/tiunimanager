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
	apiModels "github.com/pingcap-inc/tiem/apimodels/datatransfer/changefeed"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// Create create a change feed task
// @Summary create a change feed task
// @Description create a change feed task
// @Tags change feed
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param changeFeedTask body apiModels.ChangeFeedTask true "change feed task request"
// @Success 200 {object} controller.CommonResult{data=apiModels.ChangeFeedTask{downstream=apiModels.MysqlDownstream}}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/ [post]
func Create(c *gin.Context) {
	var req apiModels.ChangeFeedTask

	err, requestBody := controller.HandleJsonRequestFromBody(c, &req)

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask,
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
// @Param queryReq query apiModels.QueryReq true "change feed tasks query condition"
// @Success 200 {object} controller.ResultWithPage{data=[]apiModels.ChangeFeedTaskDetail}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/ [get]
func Query(c *gin.Context) {
	var req apiModels.QueryReq

	err, requestBody := controller.HandleJsonRequestFromBody(c, &req)

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryChangeFeedTasks,
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
// @Success 200 {object} controller.CommonResult{data=apiModels.ChangeFeedTaskDetail}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId}/ [get]
func Detail(c *gin.Context) {
	err, requestBody := controller.HandleJsonRequestWithBuiltReq(c, &apiModels.DetailReq{
		Id: c.Param(paramNameOfChangeFeedTaskId),
	})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask,
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
// @Success 200 {object} controller.CommonResult{data=apiModels.ChangeFeedTaskDetail}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId}/pause [post]
func Pause(c *gin.Context) {
	err, requestBody := controller.HandleJsonRequestWithBuiltReq(c, &apiModels.PauseReq{
		Id: c.Param(paramNameOfChangeFeedTaskId),
	})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask,
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
// @Success 200 {object} controller.CommonResult{data=apiModels.ChangeFeedTaskDetail{downstream=apiModels.TiDBDownstream}}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId}/resume [post]
func Resume(c *gin.Context) {
	err, requestBody := controller.HandleJsonRequestWithBuiltReq(c, &apiModels.ResumeReq{
		Id: c.Param(paramNameOfChangeFeedTaskId),
	})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask,
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
// @Param task body apiModels.ChangeFeedTask true "change feed task"
// @Success 200 {object} controller.CommonResult{data=apiModels.ChangeFeedTaskDetail{downstream=apiModels.KafkaDownstream}}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId}/update [post]
func Update(c *gin.Context) {
	var req apiModels.UpdateReq

	err, requestBody := controller.HandleJsonRequestFromBody(c,
		&req,
		// append id in path to request
		func(c *gin.Context, req interface{}) error {
			req.(*apiModels.UpdateReq).Id = c.Param(paramNameOfChangeFeedTaskId)
			return nil
		})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask,
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
// @Success 200 {object} controller.CommonResult{data=apiModels.ChangeFeedTaskDetail}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /changefeeds/{changeFeedTaskId} [delete]
func Delete(c *gin.Context) {
	err, requestBody := controller.HandleJsonRequestWithBuiltReq(c, &apiModels.DeleteReq{
		Id: c.Param(paramNameOfChangeFeedTaskId),
	})

	if err == nil {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateChangeFeedTask,
			requestBody,
			controller.DefaultTimeout)
	}

}
