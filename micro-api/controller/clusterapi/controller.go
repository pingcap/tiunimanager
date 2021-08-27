package clusterapi

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pingcap-inc/tiem/library/knowledge"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"github.com/pingcap-inc/tiem/micro-cluster/proto"
)

// Create 创建集群接口
// @Summary 创建集群接口
// @Description 创建集群接口
// @Tags cluster
// @Accept application/json
// @Produce application/json
// @Param Token header string true "token"
// @Param createReq body CreateReq true "创建参数"
// @Success 200 {object} controller.CommonResult{data=CreateClusterRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters [post]
func Create(c *gin.Context) {
	var req CreateReq

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	baseInfo, demand := req.ConvertToDTO()

	reqDTO := &cluster.ClusterCreateReqDTO{
		Operator: operator.ConvertToDTO(),
		Cluster:  baseInfo,
		Demands:  demand,
	}

	respDTO, err := client.ClusterClient.CreateCluster(context.TODO(), reqDTO, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetRespStatus()

		result := controller.BuildCommonResult(int(status.Code), status.Message, CreateClusterRsp{
			ClusterId:       respDTO.GetClusterId(),
			ClusterBaseInfo: *ParseClusterBaseInfoFromDTO(respDTO.GetBaseInfo()),
			StatusInfo:      *ParseStatusFromDTO(respDTO.GetClusterStatus()),
		})

		c.JSON(http.StatusOK, result)
	}
}

// Query 查询集群列表
// @Summary 查询集群列表
// @Description 查询集群列表
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param queryReq query QueryReq false "page" default(1)
// @Success 200 {object} controller.ResultWithPage{data=[]ClusterDisplayInfo}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters [get]
func Query(c *gin.Context) {

	var queryReq QueryReq

	if err := c.ShouldBindQuery(&queryReq); err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	reqDTO := &cluster.ClusterQueryReqDTO{
		Operator:      operator.ConvertToDTO(),
		PageReq:       queryReq.PageRequest.ConvertToDTO(),
		ClusterId:     queryReq.ClusterId,
		ClusterType:   queryReq.ClusterType,
		ClusterName:   queryReq.ClusterName,
		ClusterTag:    queryReq.ClusterTag,
		ClusterStatus: queryReq.ClusterStatus,
	}

	respDTO, err := client.ClusterClient.QueryCluster(context.TODO(), reqDTO, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetRespStatus()

		clusters := make([]ClusterDisplayInfo, len(respDTO.Clusters), len(respDTO.Clusters))

		for i, v := range respDTO.Clusters {
			clusters[i] = *ParseDisplayInfoFromDTO(v)
		}

		result := controller.BuildResultWithPage(int(status.Code), status.Message, controller.ParsePageFromDTO(respDTO.Page), clusters)

		c.JSON(http.StatusOK, result)
	}
}

// Delete 删除集群
// @Summary 删除集群
// @Description 删除集群
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param clusterId path string true "待删除的集群ID"
// @Success 200 {object} controller.CommonResult{data=DeleteClusterRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId} [delete]
func Delete(c *gin.Context) {

	operator := controller.GetOperator(c)

	reqDTO := &cluster.ClusterDeleteReqDTO{
		Operator:  operator.ConvertToDTO(),
		ClusterId: c.Param("clusterId"),
	}

	respDTO, err := client.ClusterClient.DeleteCluster(context.TODO(), reqDTO, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetRespStatus()

		result := controller.BuildCommonResult(int(status.Code), status.Message, DeleteClusterRsp{
			ClusterId:  respDTO.GetClusterId(),
			StatusInfo: *ParseStatusFromDTO(respDTO.GetClusterStatus()),
		})

		c.JSON(http.StatusOK, result)
	}
}

// Detail 查看集群详情
// @Summary 查看集群详情
// @Description 查看集群详情
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param clusterId path string true "集群ID"
// @Success 200 {object} controller.CommonResult{data=DetailClusterRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId} [get]
func Detail(c *gin.Context) {
	operator := controller.GetOperator(c)

	reqDTO := &cluster.ClusterDetailReqDTO{
		Operator:  operator.ConvertToDTO(),
		ClusterId: c.Param("clusterId"),
	}

	respDTO, err := client.ClusterClient.DetailCluster(context.TODO(), reqDTO, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetRespStatus()

		display := respDTO.GetDisplayInfo()
		maintenance := respDTO.GetMaintenanceInfo()
		components := respDTO.GetComponents()

		componentInstances := make([]ComponentInstance, len(components), len(components))
		for i, v := range components {
			componentInstances[i] = *ParseComponentInfoFromDTO(v)
		}

		result := controller.BuildCommonResult(int(status.Code), status.Message, DetailClusterRsp{
			ClusterDisplayInfo:     *ParseDisplayInfoFromDTO(display),
			ClusterMaintenanceInfo: *ParseMaintenanceInfoFromDTO(maintenance),
			Components:             componentInstances,
		})

		c.JSON(http.StatusOK, result)
	}
}

// ClusterKnowledge 查看集群基本知识
// @Summary 查看集群基本知识
// @Description 查看集群基本知识
// @Tags knowledge
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Success 200 {object} controller.CommonResult{data=[]knowledge.ClusterTypeSpec}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /knowledges [get]
func ClusterKnowledge(c *gin.Context) {
	c.JSON(http.StatusOK, controller.Success(knowledge.SpecKnowledge.Specs))
}


// DescribeDashboard
// @Summary 查看集群dashboard信息
// @Description 查看集群dashboard信息
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Success 200 {object} controller.CommonResult{data=DescribeDashboardRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /cluster/dashboard [get]
func DescribeDashboard(c *gin.Context) {
	operator := controller.GetOperator(c)
	reqDTO := &cluster.DescribeDashboardRequest{
		Operator: operator.ConvertToDTO(),
		ClusterId: c.Param("clusterId"),
	}
	respDTO, err := client.ClusterClient.DescribeDashboard(context.TODO(), reqDTO, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetStatus()
		result := controller.BuildCommonResult(int(status.Code), status.Message, DescribeDashboardRsp{
			ClusterId: respDTO.GetClusterId(),
			Url: respDTO.GetUrl(),
			ShareCode: respDTO.GetShareCode(),
		})

		c.JSON(http.StatusOK, result)
	}
}