package instanceapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/micro-api/controller"
	"github.com/pingcap/tiem/micro-api/controller/clusterapi"
	cluster "github.com/pingcap/tiem/micro-cluster/client"
	proto "github.com/pingcap/tiem/micro-cluster/proto"
)

// QueryParams 查询集群参数列表
// @Summary 查询集群参数列表
// @Description 查询集群参数列表
// @Tags cluster params
// @Accept json
// @Produce json
// @Param Token header string true "token"
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
	resp, err := cluster.ClusterClient.QueryParameters(context.TODO(), &proto.QueryClusterParametersRequest{
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

// SubmitParams 提交参数
// @Summary 提交参数
// @Description 提交参数
// @Tags cluster params
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param updateReq body ParamUpdateReq true "要提交的参数信息"
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

	resp, err := cluster.ClusterClient.SaveParameters(context.TODO(), &proto.SaveClusterParametersRequest{
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
// @Param Token header string true "token"
// @Param backupReq body BackupReq true "要备份的集群信息"
// @Success 200 {object} controller.CommonResult{data=BackupRecord}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups [post]
func Backup(c *gin.Context) {

	var req BackupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	resp, err := cluster.ClusterClient.CreateBackup(context.TODO(), &proto.CreateBackupRequest{
		ClusterId: req.ClusterId,
		BackupType: req.BackupType,
		BackupRange: req.BackupRange,
		FilePath: req.FilePath,
		Operator:  operator.ConvertToDTO(),
	}, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		c.JSON(http.StatusOK, controller.Success(BackupRecord{
			ID:     		resp.GetBackupRecord().GetId(),
			ClusterId: 		resp.GetBackupRecord().GetClusterId(),
			StartTime:		time.Unix(resp.GetBackupRecord().GetStartTime(), 0),
			EndTime: 		time.Unix(resp.GetBackupRecord().GetEndTime(), 0),
			BackupRange:	resp.GetBackupRecord().GetRange(),
			BackupType: 	resp.GetBackupRecord().GetBackupType(),
			FilePath:  		resp.GetBackupRecord().GetFilePath(),
			Size:  			resp.GetBackupRecord().GetSize(),
			Status: *clusterapi.ParseStatusFromDTO(resp.GetBackupRecord().DisplayStatus),
		}))
	}
}

// QueryBackupStrategy
// @Summary 查询备份策略
// @Description 查询备份策略
// @Tags cluster backup
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.CommonResult{data=[]BackupStrategy}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/strategy/ [get]
func QueryBackupStrategy(c *gin.Context) {
	clusterId := c.Param("clusterId")
	operator := controller.GetOperator(c)

	resp, err := cluster.ClusterClient.GetBackupStrategy(context.TODO(), &proto.GetBackupStrategyRequest{
		ClusterId: clusterId,
		Operator:  operator.ConvertToDTO(),
	}, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		c.JSON(http.StatusOK, controller.Success(BackupStrategy{
			ClusterId: resp.GetStrategy().GetClusterId(),
			FilePath: resp.GetStrategy().GetFilePath(),
			BackupType: resp.GetStrategy().GetBackupType(),
			BackupRange: resp.GetStrategy().GetBackupRange(),
			BackupDate: resp.GetStrategy().GetBackupDate(),
			Period: resp.GetStrategy().GetPeriod(),
		}))
	}
}

// SaveBackupStrategy
// @Summary 保存备份策略
// @Description 保存备份策略
// @Tags cluster backup
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param clusterId path string true "clusterId"
// @Param updateReq body BackupStrategyUpdateReq true "备份策略信息"
// @Success 200 {object} controller.CommonResult{data=BackupStrategy}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/strategy [put]
func SaveBackupStrategy(c *gin.Context) {
	var req BackupStrategyUpdateReq
	operator := controller.GetOperator(c)

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}
	_, err := cluster.ClusterClient.SaveBackupStrategy(context.TODO(), &proto.SaveBackupStrategyRequest{
		Operator:  operator.ConvertToDTO(),
		Strategy:  &proto.BackupStrategy{
			ClusterId: req.strategy.ClusterId,
			BackupType: req.strategy.BackupType,
			BackupRange: req.strategy.BackupRange,
			BackupDate: req.strategy.BackupDate,
			Period: req.strategy.Period,
			FilePath: req.strategy.FilePath,
		},
	}, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		c.JSON(http.StatusOK, controller.Success(nil))
	}
}

// QueryBackup
// @Summary 查询备份记录
// @Description 查询备份记录
// @Tags cluster backup
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param clusterId query string true "clusterId"
// @Param request body BackupRecordQueryReq false "page" default(1)
// @Success 200 {object} controller.ResultWithPage{data=[]BackupRecord}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups [get]
func QueryBackup(c *gin.Context) {
	clusterId := c.Query("clusterId")

	var queryReq BackupRecordQueryReq
	if err := c.ShouldBindJSON(&queryReq); err != nil {
		_ = c.Error(err)
		return
	}
	//operator := controller.GetOperator(c)

	reqDTO := &proto.QueryBackupRequest{
		ClusterId: clusterId,
		Page:      queryReq.PageRequest.ConvertToDTO(),
	}

	resp, err := cluster.ClusterClient.QueryBackupRecord(context.TODO(), reqDTO, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		records := make([]BackupRecord, len(resp.BackupRecords), len(resp.BackupRecords))

		for i, v := range resp.BackupRecords {
			records[i] = BackupRecord{
				ID:        v.Id,
				ClusterId: v.ClusterId,
				StartTime: time.Unix(v.StartTime, 0),
				EndTime:   time.Unix(v.EndTime, 0),
				BackupRange:     v.Range,
				BackupType:      v.BackupType,
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
	}
}

// RecoverBackup
// @Summary 恢复备份
// @Description 恢复备份
// @Tags cluster backup
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param backupId path string true "backupId"
// @Param request body BackupRecoverReq true "恢复备份请求"
// @Success 200 {object} controller.CommonResult{data=controller.StatusInfo}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups/{backupId}/restore [post]
func RecoverBackup(c *gin.Context) {
	backupIdStr := c.Param("backupId")
	backupId, err := strconv.Atoi(backupIdStr)
	if err != nil {
		_ = c.Error(err)
		return
	}
	var req BackupRecoverReq

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	resp, err := cluster.ClusterClient.RecoverBackupRecord(context.TODO(), &proto.RecoverBackupRequest{
		ClusterId:      req.ClusterId,
		Operator:       operator.ConvertToDTO(),
		BackupRecordId: int64(backupId),
	}, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		c.JSON(http.StatusOK, controller.Success(*clusterapi.ParseStatusFromDTO(resp.GetRecoverRecord().DisplayStatus)))
	}
}

// DeleteBackup
// @Summary 删除备份记录
// @Description 删除备份记录
// @Tags cluster backup
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param backupId path int true "删除备份ID"
// @Success 200 {object} controller.CommonResult{data=int}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backups/{backupId} [delete]
func DeleteBackup(c *gin.Context) {
	backupId, _ := strconv.Atoi(c.Param("backupId"))
	operator := controller.GetOperator(c)

	_, err := cluster.ClusterClient.DeleteBackupRecord(context.TODO(), &proto.DeleteBackupRequest{
		BackupRecordId: int64(backupId),
		Operator:       operator.ConvertToDTO(),
	}, controller.DefaultTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		c.JSON(http.StatusOK, controller.Success(backupId))
	}
}
