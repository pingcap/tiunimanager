package instanceapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"net/http"
)

// Query 查询实例接口
// @Summary 查询实例接口
// @Description 查询实例
// @Tags instance
// @Accept application/json
// @Produce application/json
// @Param Token header string true "登录token"
// @Param {object} body InstanceQuery true "查询请求"
// @Success 200 {object} controller.ResultWithPage{data=[]InstanceInfo}
// @Router /instance/query [post]
func Query(c *gin.Context) {
	//var req InstanceQuery
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	_ = c.Error(err)
	//	return
	//}

	instanceInfos := make([]InstanceInfo, 2 ,2)
	instanceInfos[0] = InstanceInfo{InstanceName: "instance1"}
	instanceInfos[1] = InstanceInfo{InstanceName: "instance2"}
	c.JSON(http.StatusOK, controller.SuccessWithPage(instanceInfos, controller.Page{Page: 1, PageSize: 20, Total: 2}))
}

// Create 创建实例接口
// @Summary 创建实例接口
// @Description 创建实例
// @Tags instance
// @Accept application/json
// @Produce application/json
// @Param Token header string true "登录token"
// @Param {object} body InstanceCreate true "创建请求"
// @Success 200 {object} controller.CommonResult{data=InstanceInfo}
// @Router /instance/create [post]
func Create(c *gin.Context) {
	//var req InstanceCreate
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	_ = c.Error(err)
	//	return
	//}
	c.JSON(http.StatusOK, controller.Success(InstanceInfo{InstanceName: "newInstance"}))
}