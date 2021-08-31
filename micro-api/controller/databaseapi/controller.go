package databaseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	cluster "github.com/pingcap-inc/tiem/micro-cluster/proto"
	"net/http"
	"time"
)

// ExportData
// @Summary export data from tidb cluster
// @Description export
// @Tags cluster export
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param dataExport body DataExportReq true "cluster info for data export"
// @Success 200 {object} controller.CommonResult{data=DataExportResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/export [post]
func ExportData(c *gin.Context) {
	var req DataExportReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)
	
	respDTO, err := client.ClusterClient.ExportData(c,&cluster.DataExportRequest{
		Operator: operator.ConvertToDTO(),
		ClusterId: req.ClusterId,
		UserName: req.UserName,
		Password: req.Password,
		FileType: req.FileType,
		Filter: req.Filter,
	}, controller.DefaultTimeout)

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

// ImportData
// @Summary import data to tidb cluster
// @Description import
// @Tags cluster import
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param dataImport body DataImportReq true "cluster info for import data"
// @Success 200 {object} controller.CommonResult{data=DataImportResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/import [post]
func ImportData(c *gin.Context) {
	var req DataImportReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	respDTO, err := client.ClusterClient.ImportData(c,&cluster.DataImportRequest{
		Operator: operator.ConvertToDTO(),
		ClusterId: req.ClusterId,
		UserName: req.UserName,
		Password: req.Password,
		FilePath: req.FilePath,
	}, controller.DefaultTimeout)

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


// DescribeDataTransport
// @Summary query records of import and export
// @Description query records of import and export
// @Tags cluster data transport
// @Accept json
// @Produce json
// @Param Token header string true "token"
// @Param dataTransportQueryReq body DataTransportQueryReq true "cluster info for query records"
// @Success 200 {object} controller.CommonResult{data=[]DataTransportRecordQueryResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/transport [get]
func DescribeDataTransport(c *gin.Context) {
	clusterId := c.Query("clusterId")
	var req DataTransportQueryReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)
	respDTO, err := client.ClusterClient.DescribeDataTransport(c, &cluster.DataTransportQueryRequest{
		Operator: operator.ConvertToDTO(),
		ClusterId: clusterId,
		RecordId: req.RecordId,
		PageReq: req.PageRequest.ConvertToDTO(),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(500, err.Error()))
	}else {
		status := respDTO.GetRespStatus()
		data := &DataTransportRecordQueryResp{
			TransportRecords: make([]*DataTransportInfo, len(respDTO.GetTransportInfos())),
		}
		for index := 0; index < len(data.TransportRecords); index++ {
			data.TransportRecords[index] = &DataTransportInfo{
				RecordId: respDTO.GetTransportInfos()[index].GetRecordId(),
				ClusterId: respDTO.GetTransportInfos()[index].GetClusterId(),
				TransportType: respDTO.GetTransportInfos()[index].GetTransportType(),
				Status: respDTO.GetTransportInfos()[index].GetStatus(),
				FilePath: respDTO.GetTransportInfos()[index].GetFilePath(),
				StartTime: time.Unix(respDTO.GetTransportInfos()[index].GetStartTime(), 0),
				EndTime: time.Unix(respDTO.GetTransportInfos()[index].GetEndTime(), 0),
			}

		}

		result := controller.BuildResultWithPage(int(status.Code), status.Message, controller.ParsePageFromDTO(respDTO.PageReq), data)

		c.JSON(http.StatusOK, result)
	}
}