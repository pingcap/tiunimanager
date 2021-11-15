/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package importexport

import (
	"github.com/pingcap-inc/tiem/micro-api/controller/cluster/management"
	"github.com/pingcap-inc/tiem/micro-api/interceptor"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-api/controller"
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
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "ExportData", int(status.GetCode()))

	var req DataExportReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
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
		StorageType:     req.StorageType,
		ZipName:         req.ZipName,
		BucketUrl:       req.BucketUrl,
		BucketRegion:    req.BucketRegion,
		EndpointUrl:     req.EndpointUrl,
		AccessKey:       req.AccessKey,
		SecretAccessKey: req.SecretAccessKey,
		Comment:         req.Comment,
	}, controller.DefaultTimeout)

	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
	} else {
		status = respDTO.GetRespStatus()
		if int32(common.TIEM_SUCCESS) == status.GetCode() {
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
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "ImportData", int(status.GetCode()))

	var req DataImportReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)

	respDTO, err := client.ClusterClient.ImportData(framework.NewMicroCtxFromGinCtx(c), &clusterpb.DataImportRequest{
		Operator:        operator.ConvertToDTO(),
		ClusterId:       req.ClusterId,
		UserName:        req.UserName,
		Password:        req.Password,
		RecordId:        req.RecordId,
		StorageType:     req.StorageType,
		BucketUrl:       req.BucketUrl,
		EndpointUrl:     req.EndpointUrl,
		AccessKey:       req.AccessKey,
		SecretAccessKey: req.SecretAccessKey,
		Comment:         req.Comment,
	}, controller.DefaultTimeout)

	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
	} else {
		status = respDTO.GetRespStatus()
		if int32(common.TIEM_SUCCESS) == status.GetCode() {
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
// @Param dataTransportQueryReq query DataTransportQueryReq true "transport records query condition"
// @Success 200 {object} controller.CommonResult{data=DataTransportRecordQueryResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/transport [get]
func DescribeDataTransport(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "DescribeDataTransport", int(status.GetCode()))

	var req DataTransportQueryReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		_ = c.Error(err)
		return
	}

	operator := controller.GetOperator(c)
	respDTO, err := client.ClusterClient.DescribeDataTransport(framework.NewMicroCtxFromGinCtx(c), &clusterpb.DataTransportQueryRequest{
		Operator:  operator.ConvertToDTO(),
		ClusterId: req.ClusterId,
		RecordId:  req.RecordId,
		ReImport:  req.ReImport,
		PageReq:   req.PageRequest.ConvertToDTO(),
	})

	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
	} else {
		status = respDTO.GetRespStatus()
		if int32(common.TIEM_SUCCESS) == status.GetCode() {
			data := &DataTransportRecordQueryResp{
				TransportRecords: make([]*DataTransportInfo, len(respDTO.GetTransportInfos())),
			}
			for index, value := range respDTO.GetTransportInfos() {
				data.TransportRecords[index] = &DataTransportInfo{
					RecordId:      value.GetRecordId(),
					ClusterId:     value.GetClusterId(),
					TransportType: value.GetTransportType(),
					FilePath:      value.GetFilePath(),
					StorageType:   value.GetStorageType(),
					Status:        *management.ParseStatusFromDTO(value.DisplayStatus),
					StartTime:     time.Unix(value.GetStartTime(), 0),
					EndTime:       time.Unix(value.GetEndTime(), 0),
					Comment:       value.GetComment(),
				}
			}
			result := controller.SuccessWithPage(data, *controller.ParsePageFromDTO(respDTO.PageReq))
			c.JSON(http.StatusOK, result)
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}

// DeleteDataTransportRecord
// @Summary delete data transport record
// @Description delete data transport record
// @Tags cluster data transport
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param recordId path int true "data transport recordId"
// @Param DataTransportDeleteReq body DataTransportDeleteReq true "data transport record delete request"
// @Success 200 {object} controller.CommonResult{data=int}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/transport/{recordId} [delete]
func DeleteDataTransportRecord(c *gin.Context) {
	var status *clusterpb.ResponseStatusDTO
	start := time.Now()
	defer interceptor.HandleMetrics(start, "DeleteDataTransportRecord", int(status.GetCode()))

	recordId, err := strconv.Atoi(c.Param("recordId"))
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	var req DataTransportDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	operator := controller.GetOperator(c)
	resp, err := client.ClusterClient.DeleteDataTransportRecord(framework.NewMicroCtxFromGinCtx(c), &clusterpb.DataTransportDeleteRequest{
		Operator:  operator.ConvertToDTO(),
		ClusterId: req.ClusterId,
		RecordId:  int64(recordId),
	}, controller.DefaultTimeout)
	if err != nil {
		status = &clusterpb.ResponseStatusDTO{Code: http.StatusBadRequest, Message: err.Error()}
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
	} else {
		status = resp.GetStatus()
		if common.TIEM_SUCCESS.Equal(status.GetCode()) {
			c.JSON(http.StatusOK, controller.Success(recordId))
		} else {
			c.JSON(http.StatusBadRequest, controller.Fail(int(status.GetCode()), status.GetMessage()))
		}
	}
}
