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
	"encoding/json"
	"github.com/pingcap-inc/tiem/message/cluster"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
	"net/http"
	"time"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"

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
	var request cluster.BackupClusterDataReq
	if err := c.ShouldBindQuery(&request); err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controller.InvokeRpcMethod(c, client.ClusterClient.CreateBackup, &cluster.BackupClusterDataResp{}, string(body), controller.DefaultTimeout)
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
	var request cluster.GetBackupStrategyReq
	request.ClusterID = c.Param("clusterId")
	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controller.InvokeRpcMethod(c, client.ClusterClient.GetBackupStrategy, &cluster.GetBackupStrategyResp{}, string(body), controller.DefaultTimeout)
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
	var request cluster.SaveBackupStrategyReq
	if err := c.ShouldBindQuery(&request); err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.ClusterID = c.Param("clusterId")
	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controller.InvokeRpcMethod(c, client.ClusterClient.SaveBackupStrategy, &cluster.SaveBackupStrategyResp{}, string(body), controller.DefaultTimeout)
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
	if err := c.ShouldBindQuery(&request); err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controller.InvokeRpcMethod(c, client.ClusterClient.QueryBackupRecords, &cluster.QueryBackupRecordsResp{}, string(body), controller.DefaultTimeout)
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
	var request cluster.DeleteBackupDataReq
	if err := c.ShouldBindQuery(&request); err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.BackupID = c.Param("backupId")
	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controller.InvokeRpcMethod(c, client.ClusterClient.DeleteBackupRecord, &cluster.DeleteBackupDataResp{}, string(body), controller.DefaultTimeout)
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
	start := time.Now()
	defer interceptor.HandleMetrics(start, "Restore", int(status.GetCode()))

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
