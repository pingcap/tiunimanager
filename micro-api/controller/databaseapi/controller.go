package databaseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiem/micro-api/controller"
	"github.com/pingcap/tiem/micro-cluster/client"
	cluster "github.com/pingcap/tiem/micro-cluster/proto"
	"net/http"
	"time"
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
		Filter: req.Filter,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetRespStatus()

		result := controller.BuildCommonResult(int(status.Code), status.Message, DataExportResp{
			RecordId: respDTO.GetRecordId(),
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
		FilePath: req.FilePath,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	} else {
		status := respDTO.GetRespStatus()

		result := controller.BuildCommonResult(int(status.Code), status.Message, DataImportResp{
			RecordId: respDTO.GetRecordId(),
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

	operator := controller.GetOperator(c)
	respDTO, err := client.ClusterClient.DescribeDataTransport(c, &cluster.DataTransportQueryRequest{
		Operator: operator.ConvertToDTO(),
		ClusterId: req.ClusterId,
		RecordId: req.RecordId,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	}else {
		status := respDTO.GetRespStatus()

		result := controller.BuildCommonResult(int(status.Code), status.Message, DataTransportInfo{
			RecordId: respDTO.GetTransportInfo().GetRecordId(),
			ClusterId: respDTO.GetTransportInfo().GetClusterId(),
			FilePath: respDTO.GetTransportInfo().GetFilePath(),
			Status: respDTO.GetTransportInfo().GetStatus(),
			TransportType: respDTO.GetTransportInfo().GetTransportType(),
			StartTime: time.Unix(respDTO.GetTransportInfo().GetStartTime(), 0),
			EndTime: time.Unix(respDTO.GetTransportInfo().GetEndTime(), 0),
		})

		c.JSON(http.StatusOK, result)
	}
}