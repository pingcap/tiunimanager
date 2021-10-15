package taskapi

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"net/http"
	"strconv"
	"time"
)

// Query query flow works
// @Summary query flow works
// @Description query flow works
// @Tags task
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param queryReq query QueryReq false "query request"
// @Success 200 {object} controller.ResultWithPage{data=[]FlowWorkDisplayInfo}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /flowworks [get]
func Query(c *gin.Context) {
	var queryReq QueryReq
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		_ = c.Error(err)
		return
	}

	reqDTO := &cluster.ListFlowsRequest{
		BizId:   queryReq.ClusterId,
		Keyword: queryReq.Keyword,
		Status: int64(queryReq.Status),
		Page: queryReq.PageRequest.ConvertToDTO(),
	}

	respDTO, err := client.ClusterClient.ListFlows(context.TODO(), reqDTO, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetStatus()

		flows := make([]FlowWorkDisplayInfo, len(respDTO.Flows), len(respDTO.Flows))

		for i, v := range respDTO.Flows {
			flows[i] = FlowWorkDisplayInfo{
				Id: uint(v.Id),
				FlowWorkName: v.FlowName,
				ClusterId:    v.BizId,
				StatusInfo: controller.StatusInfo{
					CreateTime:	time.Unix(v.CreateTime, 0),
					UpdateTime:	time.Unix(v.UpdateTime, 0),
					DeleteTime:	time.Unix(v.DeleteTime, 0),
					StatusCode: strconv.Itoa(int(v.Status)),
					StatusName: v.StatusName,
				},
			}
			if v.Operator != nil {
				flows[i].ManualOperator = true
				flows[i].OperatorName = v.Operator.Name
				flows[i].OperatorId = v.Operator.Id
				flows[i].TenantId = v.Operator.TenantId
			}
		}

		result := controller.BuildResultWithPage(int(status.Code), status.Message, controller.ParsePageFromDTO(respDTO.Page), flows)

		c.JSON(http.StatusOK, result)
	}
}

// Detail show details of a flow work
// @Summary show details of a flow work
// @Description show details of a flow work
// @Tags task
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param flowWorkId path string true "flow work id"
// @Success 200 {object} controller.CommonResult{data=FlowWorkDetailInfo}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /flowwork/{flowWorkId} [get]
func Detail(c *gin.Context) {

}
