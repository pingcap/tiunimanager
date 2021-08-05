package instanceapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/library/knowledge"
	"github.com/pingcap/tiem/micro-api/controller"
	"net/http"
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
	c.JSON(http.StatusOK, controller.SuccessWithPage([]ParamItem{
		{Definition: knowledge.Parameter{Name: clusterId}, CurrentValue: ParamInstance{Value: clusterId}},
		{Definition: knowledge.Parameter{Name: "p2"}, CurrentValue: ParamInstance{Value: 2}},
	}, controller.Page{Page: req.Page, PageSize: req.PageSize}))
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

	c.JSON(http.StatusOK, controller.Success(ParamUpdateRsp{
		ClusterId: req.ClusterId,
		TaskId: 111,
	}))
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
	c.JSON(http.StatusOK, controller.Success(BackupRecord{ClusterId: clusterId}))
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
// @Success 200 {object} controller.ResultWithPage{data=[]BackupRecord}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /backup/records/{clusterId} [post]
func QueryBackup(c *gin.Context) {
	clusterId := c.Param("clusterId")
	c.JSON(http.StatusOK, controller.Success([]BackupRecord{
		{ClusterId: clusterId, ID: 111},
	}))
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

