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

	//todo: call cluster api
	resp,err := client.ClusterClient.ExportData(c,&cluster.DataExportRequest{
		TenantId: req.TenantId,
		ClusterId: req.ClusterId,
		UserName: req.UserName,
		Password: req.Password,
		FileType: req.FileType,
	})

	if err != nil {
		// 处理异常
	}

	c.JSON(http.StatusOK, controller.Success(resp.FlowId))
}

func ImportData(c *gin.Context) {
	var req DataImport
	err := c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	//todo: call cluster api

	if err != nil {
		// 处理异常
	}

	//c.JSON(http.StatusOK, controller.Success(CopyInstanceInfoFromDTO(resp.Cluster)))
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