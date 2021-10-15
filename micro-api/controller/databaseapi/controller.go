package databaseapi

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"net/http"
	"time"
)

// ExportData
// @Summary export data from tidb cluster
// @Description export
// @Tags cluster export
// @Accept json
// @Produce json
// @Security ApiKeyAuth
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

	respDTO, err := client.ClusterClient.ExportData(framework.NewMicroCtxFromGinCtx(c), &clusterpb.DataExportRequest{
		Operator:        operator.ConvertToDTO(),
		ClusterId:       req.ClusterId,
		UserName:        req.UserName,
		Password:        req.Password,
		FileType:        req.FileType,
		Filter:          req.Filter,
		Sql:             req.Sql,
		FilePath:        req.FilePath,
		StorageType:     req.StorageType,
		BucketUrl:       req.BucketUrl,
		BucketRegion:    req.BucketRegion,
		EndpointUrl:     req.EndpointUrl,
		AccessKey:       req.AccessKey,
		SecretAccessKey: req.SecretAccessKey,
	}, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
	} else {
		status := respDTO.GetRespStatus()
		if common.TIEM_SUCCESS == status.GetCode() {
			result := controller.BuildCommonResult(int(status.Code), status.Message, DataExportResp{
				RecordId: respDTO.GetRecordId(),
			})
			c.JSON(http.StatusOK, result)
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}

// ImportData
// @Summary import data to tidb cluster
// @Description import
// @Tags cluster import
// @Accept json
// @Produce json
// @Security ApiKeyAuth
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

	respDTO, err := client.ClusterClient.ImportData(framework.NewMicroCtxFromGinCtx(c), &clusterpb.DataImportRequest{
		Operator:    operator.ConvertToDTO(),
		ClusterId:   req.ClusterId,
		UserName:    req.UserName,
		Password:    req.Password,
		FilePath:    req.FilePath,
		StorageType: req.StorageType,
	}, controller.DefaultTimeout)

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
	} else {
		status := respDTO.GetRespStatus()
		if common.TIEM_SUCCESS == status.GetCode() {
			result := controller.BuildCommonResult(int(status.Code), status.Message, DataImportResp{
				RecordId: respDTO.GetRecordId(),
			})
			c.JSON(http.StatusOK, result)
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}

// DescribeDataTransport
// @Summary query records of import and export
// @Description query records of import and export
// @Tags cluster data transport
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param clusterId path string true "cluster id"
// @Param dataTransportQueryReq query DataTransportQueryReq false "transport records query condition"
// @Success 200 {object} controller.CommonResult{data=[]DataTransportRecordQueryResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/{clusterId}/transport [get]
func DescribeDataTransport(c *gin.Context) {
	clusterId := c.Param("clusterId")
	var req DataTransportQueryReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)
	respDTO, err := client.ClusterClient.DescribeDataTransport(framework.NewMicroCtxFromGinCtx(c), &clusterpb.DataTransportQueryRequest{
		Operator:  operator.ConvertToDTO(),
		ClusterId: clusterId,
		RecordId:  req.RecordId,
		PageReq:   req.PageRequest.ConvertToDTO(),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, controller.Fail(http.StatusInternalServerError, err.Error()))
	} else {
		status := respDTO.GetRespStatus()
		if common.TIEM_SUCCESS == status.GetCode() {
			data := &DataTransportRecordQueryResp{
				TransportRecords: make([]*DataTransportInfo, len(respDTO.GetTransportInfos())),
			}
			for index, value := range respDTO.GetTransportInfos() {
				data.TransportRecords[index] = &DataTransportInfo{
					RecordId:      value.GetRecordId(),
					ClusterId:     value.GetClusterId(),
					TransportType: value.GetTransportType(),
					Status:        value.GetStatus(),
					FilePath:      value.GetFilePath(),
					StartTime:     time.Unix(value.GetStartTime(), 0),
					EndTime:       time.Unix(value.GetEndTime(), 0),
				}
			}
			result := controller.SuccessWithPage(data, *controller.ParsePageFromDTO(respDTO.PageReq))
			c.JSON(http.StatusOK, result)
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}
