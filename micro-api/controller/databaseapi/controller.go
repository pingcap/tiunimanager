package databaseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/ticp/micro-api/controller"
	"github.com/pingcap/ticp/micro-cluster/client"
	cluster "github.com/pingcap/ticp/micro-cluster/proto"
	"net/http"
)

func ExportData(c *gin.Context) {
	var req DataExport
	err := c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	//todo: call cluster api
	respDTO, err := client.ClusterClient.ExportData(c,&cluster.DataExportRequest{
		Operator: operator.ConvertToDTO(),
		ClusterId: req.ClusterId,
		UserName: req.UserName,
		Password: req.Password,
		FileType: req.FileType,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetRespStatus()

		result := controller.BuildCommonResult(int(status.Code), status.Message, DataExportResp{
			FlowId: respDTO.GetFlowId(),
		})

		c.JSON(http.StatusOK, result)
	}
}

func ImportData(c *gin.Context) {
	var req DataImport
	err := c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	//todo: call cluster api
	respDTO, err := client.ClusterClient.ImportData(c,&cluster.DataImportRequest{
		Operator: operator.ConvertToDTO(),
		ClusterId: req.ClusterId,
		UserName: req.UserName,
		Password: req.Password,
		DataDir: req.DataDir,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetRespStatus()

		result := controller.BuildCommonResult(int(status.Code), status.Message, DataImportResp{
			FlowId: respDTO.GetFlowId(),
		})

		c.JSON(http.StatusOK, result)
	}
}

func DescribeDataTransport(c *gin.Context) {
	var req DataTransportQuery
	err := c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	//todo: call cluster api

	if err != nil {
		// 处理异常
	}

	//c.JSON(http.StatusOK, controller.SuccessWithPage(instanceInfos, controller.Page{Page: req.Page, PageSize: req.PageSize, Total: len(instanceInfos)}))
}