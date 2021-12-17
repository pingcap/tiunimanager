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

package backuprestore

import (
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message/cluster"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

// Backup
// @Summary backup
// @Description backup
// @Tags cluster backup
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param backupReq body cluster.BackupClusterDataReq true "backup request"
// @Success 200 {object} controller.CommonResult{data=cluster.BackupClusterDataResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups/ [post]
func Backup(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &cluster.BackupClusterDataReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateBackup, &cluster.BackupClusterDataResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// GetBackupStrategy show the backup strategy of a cluster
// @Summary show the backup strategy of a cluster
// @Description show the backup strategy of a cluster
// @Tags cluster backup
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=cluster.GetBackupStrategyResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/strategy/ [get]
func GetBackupStrategy(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &cluster.GetBackupStrategyReq{
		ClusterID: c.Param("clusterId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.GetBackupStrategy, &cluster.GetBackupStrategyResp{},
			requestBody,
			controller.DefaultTimeout)
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
// @Param updateReq body cluster.SaveBackupStrategyReq true "backup strategy request"
// @Success 200 {object} controller.CommonResult{data=cluster.SaveBackupStrategyResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/strategy [put]
func SaveBackupStrategy(c *gin.Context) {
	req := cluster.SaveBackupStrategyReq{
		ClusterID: c.Param("clusterId"),
	}

	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.SaveBackupStrategy, &cluster.SaveBackupStrategyResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryBackupRecords
// @Summary query backup records of a cluster
// @Description query backup records of a cluster
// @Tags cluster backup
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param backupRecordQuery query cluster.QueryBackupRecordsReq true "backup records query condition"
// @Success 200 {object} controller.ResultWithPage{data=cluster.QueryBackupRecordsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups/ [get]
func QueryBackupRecords(c *gin.Context) {
	var request cluster.QueryBackupRecordsReq

	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &request); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryBackupRecords, &cluster.QueryBackupRecordsResp{},
			requestBody,
			controller.DefaultTimeout)
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
// @Param backupDeleteReq body cluster.DeleteBackupDataReq true "backup delete request"
// @Success 200 {object} controller.CommonResult{data=cluster.DeleteBackupDataResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups/{backupId} [delete]
func DeleteBackup(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &cluster.DeleteBackupDataReq{
		BackupID: c.Param("backupId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteBackupRecords, &cluster.DeleteBackupDataResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Restore
// @Summary restore a new cluster by backup record
// @Description restore a new cluster by backup record
// @Tags cluster
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body RestoreReq true "restore request"
// @Success 200 {object} controller.CommonResult{data=controller.StatusInfo}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/restore [post]
func Restore(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	var req RestoreReq
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	baseInfo, commonDemand, demand := req.ConvertToDTO()

	reqDTO := &clusterpb.RecoverRequest{
		Operator:     operator.ConvertToDTO(),
		Cluster:      baseInfo,
		Demands:      demand,
		CommonDemand: commonDemand,
	}

	respDTO, err := client.ClusterClient.RecoverCluster(framework.NewMicroCtxFromGinCtx(c), reqDTO, controller.DefaultTimeout)

	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
	} else {
		status = respDTO.GetRespStatus()
		if common.TIEM_SUCCESS.Equal(status.GetCode()) {
			result := controller.BuildCommonResult(int(status.Code), status.Message, RecoverClusterRsp{
				ClusterId:       respDTO.GetClusterId(),
				ClusterBaseInfo: *ParseClusterBaseInfoFromDTO(respDTO.GetBaseInfo()),
				StatusInfo:      *ParseStatusFromDTO(respDTO.GetClusterStatus()),
			})
			c.JSON(http.StatusOK, result)
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}
