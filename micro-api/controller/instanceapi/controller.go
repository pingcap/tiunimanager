package instanceapi

import (
	"context"
	"encoding/json"
	"github.com/asim/go-micro/v3/client"
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/micro-api/controller"
	"github.com/pingcap/tiem/micro-api/controller/clusterapi"
	cluster "github.com/pingcap/tiem/micro-cluster/client"
	proto "github.com/pingcap/tiem/micro-cluster/proto"
	"net/http"
	"time"
)

// QueryParams 查询集群参数列表
// @Summary 查询集群参数列表
// @Description 查询集群参数列表
// @Tags cluster params
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param queryReq body ParamQueryReq false "page" default(1)
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.ResultWithPage{data=[]ParamItem}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /params/{clusterId} [post]
func QueryParams(c *gin.Context) {
	var req ParamQueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}
	clusterId := c.Param("clusterId")
	operator := controller.GetOperator(c)
	resp, err := cluster.ClusterClient.QueryParameters(context.TODO(), &proto.QueryClusterParametersRequest{
		ClusterId: clusterId,
		Operator: operator.ConvertToDTO(),

	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		instances := make([]ParamInstance, 0)

		err = json.Unmarshal([]byte(resp.GetParametersJson()), &instances)

		instanceMap := make(map[string]interface{})
		if len(instances) > 0{
			for _,v := range instances {
				instanceMap[v.Name] = v.Value
			}
		}

		if err != nil  {
			c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
			return
		}

		parameters := make([]ParamItem, len(knowledge.ParameterKnowledge.Parameters), len(knowledge.ParameterKnowledge.Parameters))

		for i,v := range knowledge.ParameterKnowledge.Parameters {
			parameters[i] = ParamItem{
				Definition: *v,
				CurrentValue: ParamInstance{
					Name: v.Name,
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
// @Success 200 {object} controller.CommonResult{data=ParamUpdateRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /params/submit [post]
func SubmitParams(c *gin.Context) {
	var req ParamUpdateReq

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	clusterId := req.ClusterId
	operator := controller.GetOperator(c)
	values := req.Values

	jsonByte, _ := json.Marshal(values)

	jsonContent := string(jsonByte)

	resp, err := cluster.ClusterClient.SaveParameters(context.TODO(), &proto.SaveClusterParametersRequest{
		ClusterId: clusterId,
		ParametersJson: jsonContent,
		Operator: operator.ConvertToDTO(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		c.JSON(http.StatusOK, controller.Success(ParamUpdateRsp{
			ClusterId: req.ClusterId,
			TaskId: uint(resp.DisplayInfo.InProcessFlowId),
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
// @Param clusterId path string true "要备份的集群ID"
// @Success 200 {object} controller.CommonResult{data=BackupRecord}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backup/{clusterId} [post]
func Backup(c *gin.Context) {
	clusterId := c.Param("clusterId")

	operator := controller.GetOperator(c)

	resp, err := cluster.ClusterClient.CreateBackup(context.TODO(), &proto.CreateBackupRequest{
		ClusterId: clusterId,
		Operator: operator.ConvertToDTO(),
	}, func(o *client.CallOptions) {
		o.RequestTimeout = time.Second * 30
		o.DialTimeout = time.Second * 30
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		c.JSON(http.StatusOK, controller.Success(BackupRecord{
			ID: resp.BackupRecord.Id,
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
// @Router /backup/strategy/{clusterId} [get]
func QueryBackupStrategy(c *gin.Context) {
	clusterId := c.Param("clusterId")
	c.JSON(http.StatusOK, controller.Success([]BackupStrategy{
		{CronString: clusterId},
	}))
}

// SaveBackupStrategy
// @Summary 保存备份策略
// @Description 保存备份策略
// @Tags cluster backup
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param updateReq body BackupStrategyUpdateReq true "备份策略信息"
// @Success 200 {object} controller.CommonResult{data=BackupStrategy}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backup/strategy [post]
func SaveBackupStrategy(c *gin.Context) {
	var req BackupStrategyUpdateReq

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, controller.Success(req.BackupStrategy))
}

// QueryBackup
// @Summary 查询备份记录
// @Description 查询备份记录
// @Tags cluster backup
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param clusterId path string true "clusterId"
// @Param request body BackupRecordQueryReq false "page" default(1)
// @Success 200 {object} controller.ResultWithPage{data=[] BackupRecord}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backup/records/{clusterId} [post]
func QueryBackup(c *gin.Context) {
	clusterId := c.Param("clusterId")

	var queryReq BackupRecordQueryReq
	if err := c.ShouldBindJSON(&queryReq); err != nil {
		_ = c.Error(err)
		return
	}
	//operator := controller.GetOperator(c)

	reqDTO := &proto.QueryBackupRequest {
		ClusterId: clusterId,
		Page: queryReq.PageRequest.ConvertToDTO(),
	}

	resp, err := cluster.ClusterClient.QueryBackupRecord(context.TODO(), reqDTO)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		records := make([]BackupRecord, len(resp.BackupRecords), len(resp.BackupRecords))

		for i,v := range resp.BackupRecords {
			records[i] = BackupRecord{
				ID:        v.Id,
				ClusterId: v.ClusterId,
				StartTime: time.Unix(v.StartTime, 0),
				EndTime:   time.Unix(v.EndTime, 0),
				Range:     int(v.Range),
				Way: int(v.Way),
				Operator:  controller.Operator{
					ManualOperator: true,
					OperatorId: v.Operator.Id,
					//OperatorName: v.Operator.Name,
					//TenantId: v.Operator.TenantId,
				},
				Size:      v.Size,
				Status:    *clusterapi.ParseStatusFromDTO(v.DisplayStatus),
				FilePath:  v.FilePath,
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
// @Param request body BackupRecoverReq true "恢复备份请求"
// @Success 200 {object} controller.CommonResult{data=controller.StatusInfo}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backup/record/recover [post]
func RecoverBackup(c *gin.Context) {
	var req BackupRecoverReq

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, controller.Success(controller.StatusInfo{
		StatusName: "已完成",
	}))
}

// DeleteBackup
// @Summary 删除备份记录
// @Description 删除备份记录
// @Tags cluster backup
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param recordId path int true "删除备份ID"
// @Success 200 {object} controller.CommonResult{data=string}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backup/record/{recordId} [delete]
func DeleteBackup(c *gin.Context) {
	recordId := c.Param("recordId")

	c.JSON(http.StatusOK, controller.Success(recordId))
}