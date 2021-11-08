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
 *                                                                            *
 ******************************************************************************/

package parameter

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"

	"github.com/pingcap-inc/tiem/library/client"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
)

// QueryParams query params of a cluster
// @Summary query params of a cluster
// @Description query params of a cluster
// @Tags cluster params
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryReq query ParamQueryReq false "page" default(1)
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.ResultWithPage{data=[]ParamItem}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/params [get]
func QueryParams(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "QueryParams", int(status.GetCode()))

	var req ParamQueryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		_ = c.Error(err)
		return
	}
	clusterId := c.Param("clusterId")
	operator := controller.GetOperator(c)
	resp, err := client.ClusterClient.QueryParameters(framework.NewMicroCtxFromGinCtx(c), &clusterpb.QueryClusterParametersRequest{
		ClusterId: clusterId,
		Operator:  operator.ConvertToDTO(),
	}, controller.DefaultTimeout)

	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status = resp.GetStatus()
		instances := make([]ParamInstance, 0)

		err = json.Unmarshal([]byte(resp.GetParametersJson()), &instances)

		instanceMap := make(map[string]interface{})
		if len(instances) > 0 {
			for _, v := range instances {
				instanceMap[v.Name] = v.Value
			}
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
			return
		}

		parameters := make([]ParamItem, len(knowledge.ParameterKnowledge.Parameters))

		for i, v := range knowledge.ParameterKnowledge.Parameters {
			parameters[i] = ParamItem{
				Definition: *v,
				CurrentValue: ParamInstance{
					Name:  v.Name,
					Value: instanceMap[v.Name],
				},
			}
		}

		c.JSON(http.StatusOK, controller.SuccessWithPage(parameters, controller.Page{Page: req.Page, PageSize: req.PageSize, Total: len(parameters)}))
	}
}

// SubmitParams submit params
// @Summary submit params
// @Description submit params
// @Tags cluster params
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param updateReq body ParamUpdateReq true "update params request"
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=ParamUpdateRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/params [post]
func SubmitParams(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "SubmitParams", int(status.GetCode()))

	var req ParamUpdateReq

	if err := c.ShouldBindJSON(&req); err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		_ = c.Error(err)
		return
	}

	clusterId := c.Param("clusterId")
	operator := controller.GetOperator(c)
	values := req.Values

	jsonByte, _ := json.Marshal(values)

	jsonContent := string(jsonByte)

	resp, err := client.ClusterClient.SaveParameters(framework.NewMicroCtxFromGinCtx(c), &clusterpb.SaveClusterParametersRequest{
		ClusterId:      clusterId,
		ParametersJson: jsonContent,
		Operator:       operator.ConvertToDTO(),
	}, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusInternalServerError, Message: err.Error()}
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status = resp.GetStatus()
		c.JSON(http.StatusOK, controller.Success(ParamUpdateRsp{
			ClusterId: clusterId,
			TaskId:    uint(resp.DisplayInfo.InProcessFlowId),
		}))
	}
}
