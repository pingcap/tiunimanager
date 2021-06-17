package hostapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"net/http"
)
// Query 查询主机接口
// @Summary 查询主机接口
// @Description 查询主机
// @Tags 主机
// @Accept json
// @Produce json
// @Param Token header string true "登录token"
// @Param {object} body HostQuery false "查询请求"
// @Success 200 {object} controller.ResultWithPage{data=[]HostInfo}
// @Router /host/query [post]
func Query(c *gin.Context) {
	var req HostQuery
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	hostInfos := make([]HostInfo, 2 ,2)
	hostInfos[0] = HostInfo{HostId: "first",HostName: "hh",HostIp:"127.0.0.1:1111"}
	hostInfos[1] = HostInfo{HostId: "second",HostName: "hh",HostIp:"127.0.0.1:2222"}
	c.JSON(http.StatusOK, controller.SuccessWithPage(hostInfos, controller.Page{Page: 1, PageSize: 20, Total: 2}))
}