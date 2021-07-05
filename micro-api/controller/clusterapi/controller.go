package clusterapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"net/http"
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
// @Router /cluster [post]
func Create(c *gin.Context) {

}

// Query 查询集群列表
// @Summary 查询集群列表
// @Description 查询集群列表
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param page query int false "page" default(1)
// @Param pageSize query int false "page" default(1)
// @Success 200 {object} controller.ResultWithPage{data=[]ClusterDisplayInfo}
// @Router /clusters [get]
func Query(c *gin.Context) {

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
// @Router /cluster/ [delete]
func Delete(c * gin.Context) {

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
// @Router /cluster/{clusterId} [get]
func Detail(c *gin.Context) {
	rsp := DetailClusterRsp{}
	rsp.ClusterId = c.GetString("clusterId")
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
// @Success 200 {object} controller.CommonResult{data=ClusterKnowledgeRsp}
// @Router /knowledge [get]
func ClusterKnowledge(c *gin.Context) {

}
