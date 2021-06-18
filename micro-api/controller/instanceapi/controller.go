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
// @Param page body int false "页码"
// @Param pageSize body int false "每页数量"
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
// @Param instanceName body string true "实例名"
// @Param instanceVersion body int true "实例版本"
// @Param dbPassword body string true "数据库密码"
// @Param pdCount body int true "PD数量"
// @Param tiDBCount body int true "TiDB数量"
// @Param tiKVCount body int true "TiKV数量"
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