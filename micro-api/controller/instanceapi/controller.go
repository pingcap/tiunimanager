package instanceapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"net/http"
	"strconv"
)

// QueryParams 查询集群参数列表
// @Summary 查询集群参数列表
// @Description 查询集群参数列表
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param page query int false "page" default(1)
// @Param pageSize query int false "pageSize" default(20)
// @Param clusterId path string true "clusterId"
// @Success 200 {object} controller.ResultWithPage{data=[]ParamItem}
// @Router /params/{clusterId} [get]
func QueryParams(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	clusterId := c.Param("clusterId")
	c.JSON(http.StatusOK, controller.SuccessWithPage([]ParamItem{
		{Definition:Parameter{Name: clusterId}, CurrentValue: ParamValue{Value: clusterId}},
		{Definition:Parameter{Name: "p2"}, CurrentValue: ParamValue{Value: 2}},
	}, controller.Page{Page: page, PageSize: pageSize}))
}

// SubmitParams
// ParamUpdateReq
// ParamUpdateRsp
// SubmitParams 提交参数
// @Summary 提交参数
// @Description 提交参数
// @Tags cluster
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param request body ParamUpdateReq true "要提交的参数信息"
// @Success 200 {object} controller.CommonResult{data=ParamUpdateRsp}
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
