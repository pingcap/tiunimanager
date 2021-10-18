package backuprestore

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/library/client"
	"google.golang.org/grpc/codes"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/management"
)

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
				Status:       *management.ParseStatusFromDTO(resp.GetBackupRecord().DisplayStatus),
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
					Status:   *management.ParseStatusFromDTO(v.DisplayStatus),
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
	var req RestoreReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	baseInfo, demand := req.ConvertToDTO()

	reqDTO := &clusterpb.RecoverRequest{
		Operator: operator.ConvertToDTO(),
		Cluster:  baseInfo,
		Demands:  demand,
	}

	respDTO, err := client.ClusterClient.RecoverCluster(context.TODO(), reqDTO, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
	} else {
		status := respDTO.GetRespStatus()
		if common.TIEM_SUCCESS == status.GetCode() {
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
