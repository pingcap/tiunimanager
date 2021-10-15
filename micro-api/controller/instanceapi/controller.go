
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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

package instanceapi

import (
	"context"
	"encoding/json"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"net/http"
	"strconv"
	"time"

	"github.com/pingcap-inc/tiem/library/client"
	"google.golang.org/grpc/codes"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/clusterapi"
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
	var req ParamQueryReq
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(err)
		return
	}
	clusterId := c.Param("clusterId")
	operator := controller.GetOperator(c)
	resp, err := client.ClusterClient.QueryParameters(context.TODO(), &clusterpb.QueryClusterParametersRequest{
		ClusterId: clusterId,
		Operator:  operator.ConvertToDTO(),
	}, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
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

		parameters := make([]ParamItem, len(knowledge.ParameterKnowledge.Parameters), len(knowledge.ParameterKnowledge.Parameters))

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
	var req ParamUpdateReq

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	clusterId := c.Param("clusterId")
	operator := controller.GetOperator(c)
	values := req.Values

	jsonByte, _ := json.Marshal(values)

	jsonContent := string(jsonByte)

	resp, err := client.ClusterClient.SaveParameters(context.TODO(), &clusterpb.SaveClusterParametersRequest{
		ClusterId:      clusterId,
		ParametersJson: jsonContent,
		Operator:       operator.ConvertToDTO(),
	}, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		c.JSON(http.StatusOK, controller.Success(ParamUpdateRsp{
			ClusterId: clusterId,
			TaskId:    uint(resp.DisplayInfo.InProcessFlowId),
		}))
	}
}

// Backup backup
// @Summary backup
// @Description backup
// @Tags cluster backup
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param backupReq body BackupReq true "backup request"
// @Success 200 {object} controller.CommonResult{data=BackupRecord}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups [post]
func Backup(c *gin.Context) {

	var req BackupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	operator := controller.GetOperator(c)

	resp, err := client.ClusterClient.CreateBackup(framework.NewMicroCtxFromGinCtx(c), &clusterpb.CreateBackupRequest{
		ClusterId:    req.ClusterId,
		BackupType:   req.BackupType,
		BackupMethod: req.BackupMethod,
		FilePath:     req.FilePath,
		Operator:     operator.ConvertToDTO(),
	}, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
	} else {
		status := resp.GetStatus()
		if common.TIEM_SUCCESS == status.GetCode() {
			c.JSON(http.StatusOK, controller.Success(BackupRecord{
				ID:           resp.GetBackupRecord().GetId(),
				ClusterId:    resp.GetBackupRecord().GetClusterId(),
				StartTime:    time.Unix(resp.GetBackupRecord().GetStartTime(), 0),
				EndTime:      time.Unix(resp.GetBackupRecord().GetEndTime(), 0),
				BackupType:   resp.GetBackupRecord().GetBackupType(),
				BackupMethod: resp.GetBackupRecord().GetBackupMethod(),
				BackupMode:   resp.GetBackupRecord().GetBackupMode(),
				FilePath:     resp.GetBackupRecord().GetFilePath(),
				Size:         resp.GetBackupRecord().GetSize(),
				Status:       *clusterapi.ParseStatusFromDTO(resp.GetBackupRecord().DisplayStatus),
			}))
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}

// QueryBackupStrategy show the backup strategy of a cluster
// @Summary show the backup strategy of a cluster
// @Description show the backup strategy of a cluster
// @Tags cluster backup
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=BackupStrategy}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/strategy/ [get]
func QueryBackupStrategy(c *gin.Context) {
	clusterId := c.Param("clusterId")
	operator := controller.GetOperator(c)

	resp, err := client.ClusterClient.GetBackupStrategy(framework.NewMicroCtxFromGinCtx(c), &clusterpb.GetBackupStrategyRequest{
		ClusterId: clusterId,
		Operator:  operator.ConvertToDTO(),
	}, controller.DefaultTimeout)
	if err != nil || resp == nil || resp.GetStrategy() == nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
	} else {
		status := resp.GetStatus()
		if common.TIEM_SUCCESS == status.GetCode() {
			c.JSON(http.StatusOK, controller.Success(BackupStrategy{
				ClusterId:      resp.GetStrategy().GetClusterId(),
				BackupDate:     resp.GetStrategy().GetBackupDate(),
				Period:         resp.GetStrategy().GetPeriod(),
				NextBackupTime: time.Unix(resp.GetStrategy().GetNextBackupTime(), 0),
			}))
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}

// SaveBackupStrategy save the backup strategy of a cluster
// @Summary save the backup strategy of a cluster
// @Description save the backup strategy of a cluster
// @Tags cluster backup
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Param updateReq body BackupStrategyUpdateReq true "备份策略信息"
// @Success 200 {object} controller.CommonResult{data=BackupStrategy}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/strategy [put]
func SaveBackupStrategy(c *gin.Context) {
	clusterId := c.Param("clusterId")

	var req BackupStrategyUpdateReq
	operator := controller.GetOperator(c)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	resp, err := client.ClusterClient.SaveBackupStrategy(framework.NewMicroCtxFromGinCtx(c), &clusterpb.SaveBackupStrategyRequest{
		Operator: operator.ConvertToDTO(),
		Strategy: &clusterpb.BackupStrategy{
			ClusterId:  clusterId,
			BackupDate: req.Strategy.BackupDate,
			Period:     req.Strategy.Period,
		},
	}, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
	} else {
		status := resp.GetStatus()
		if common.TIEM_SUCCESS == status.GetCode() {
			c.JSON(http.StatusOK, controller.Success(nil))
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}

// QueryBackup
// @Summary query backup records of a cluster
// @Description query backup records of a cluster
// @Tags cluster backup
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param backupRecordQuery query BackupRecordQueryReq true "backup records query condition"
// @Success 200 {object} controller.ResultWithPage{data=[]BackupRecord}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups [get]
func QueryBackup(c *gin.Context) {
	var queryReq BackupRecordQueryReq
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	operator := controller.GetOperator(c)
	reqDTO := &clusterpb.QueryBackupRequest{
		Operator:  operator.ConvertToDTO(),
		ClusterId: queryReq.ClusterId,
		Page:      queryReq.PageRequest.ConvertToDTO(),
		StartTime: queryReq.StartTime,
		EndTime:   queryReq.EndTime,
	}

	resp, err := client.ClusterClient.QueryBackupRecord(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
	} else {
		status := resp.GetStatus()
		if common.TIEM_SUCCESS == status.GetCode() {
			records := make([]BackupRecord, len(resp.BackupRecords))
			for i, v := range resp.BackupRecords {
				records[i] = BackupRecord{
					ID:           v.Id,
					ClusterId:    v.ClusterId,
					StartTime:    time.Unix(v.StartTime, 0),
					EndTime:      time.Unix(v.EndTime, 0),
					BackupType:   v.BackupType,
					BackupMethod: v.BackupMethod,
					BackupMode:   v.BackupMode,
					Operator: controller.Operator{
						ManualOperator: true,
						OperatorId:     v.Operator.Id,
						//OperatorName: v.Operator.Name,
						//TenantId: v.Operator.TenantId,
					},
					Size:     v.Size,
					Status:   *clusterapi.ParseStatusFromDTO(v.DisplayStatus),
					FilePath: v.FilePath,
				}
			}
			c.JSON(http.StatusOK, controller.SuccessWithPage(records, *controller.ParsePageFromDTO(resp.Page)))
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}

// DeleteBackup
// @Summary delete backup record
// @Description delete backup record
// @Tags cluster backup
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param backupId path int true "backup record id"
// @Param backupDeleteReq body BackupDeleteReq true "backup delete request"
// @Success 200 {object} controller.CommonResult{data=int}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups/{backupId} [delete]
func DeleteBackup(c *gin.Context) {
	backupId, err := strconv.Atoi(c.Param("backupId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}

	var req BackupDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(int(codes.InvalidArgument), err.Error()))
		return
	}
	operator := controller.GetOperator(c)
	resp, err := client.ClusterClient.DeleteBackupRecord(framework.NewMicroCtxFromGinCtx(c), &clusterpb.DeleteBackupRequest{
		BackupRecordId: int64(backupId),
		Operator:       operator.ConvertToDTO(),
		ClusterId:      req.ClusterId,
	}, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := resp.GetStatus()
		if common.TIEM_SUCCESS == status.GetCode() {
			c.JSON(http.StatusOK, controller.Success(backupId))
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}
