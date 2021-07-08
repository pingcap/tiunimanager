package clusterapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/knowledge/models"
	"github.com/pingcap/ticp/micro-api/controller"
	"net/http"
	"strconv"
)

// Create 创建集群接口
// @Summary 创建集群接口
// @Description 创建集群接口
// @Tags cluster
// @Accept application/json
// @Produce application/json
// @Param Token header string true "token"
// @Param cluster body CreateReq true "创建参数"
// @Success 200 {object} controller.CommonResult{data=CreateClusterRsp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /cluster [post]
func Create(c *gin.Context) {
	var req CreateReq

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	rsp := CreateClusterRsp{ClusterId: "aaa"}
	rsp.DbPassword = req.DbPassword

	c.JSON(http.StatusOK, controller.Success(rsp))
}

// Query 查询集群列表
// @Summary 查询集群列表
// @Description 查询集群列表
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param page query int false "page" default(1)
// @Param pageSize query int false "pageSize" default(20)
// @Success 200 {object} controller.ResultWithPage{data=[]ClusterDisplayInfo}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /cluster/query [get]
func Query(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	clusters := []ClusterDisplayInfo{
		ClusterDisplayInfo{ClusterId: "cluster1"},
		ClusterDisplayInfo{ClusterId: "cluster2"},
	}

	c.JSON(http.StatusOK, controller.SuccessWithPage(clusters, controller.Page{Page: page, PageSize: pageSize}))
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
// @Router /cluster/{clusterId} [delete]
func Delete(c * gin.Context) {
	rsp := DeleteClusterRsp{}
	rsp.ClusterId = c.Param("clusterId")

	c.JSON(http.StatusOK, controller.Success(rsp))
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
// @Router /cluster/{clusterId} [get]
func Detail(c *gin.Context) {
	rsp := DetailClusterRsp{}

	rsp.ClusterId = c.Param("clusterId")
	rsp.DbPassword = "pass"
	c.JSON(http.StatusOK, controller.Success(rsp))
}

// ClusterKnowledge 查看集群基本知识
// @Summary 查看集群基本知识
// @Description 查看集群基本知识
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Success 200 {object} controller.CommonResult{data=[]models.ClusterTypeSpec}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /cluster/knowledge [get]
func ClusterKnowledge(c *gin.Context) {
	c.JSON(http.StatusOK, controller.Success([]models.ClusterTypeSpec{
		{
			models.ClusterType{
				Name: "what",
			},
			[]models.ClusterVersionSpec{

			},
		},
	}))
}
