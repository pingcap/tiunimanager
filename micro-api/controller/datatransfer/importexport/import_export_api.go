/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/client"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/micro-api/controller"
)

// ExportData
// @Summary export data from tidb cluster
// @Description export
// @Tags cluster export
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param dataExport body message.DataExportReq true "cluster info for data export"
// @Success 200 {object} controller.CommonResult{data=message.DataExportResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/export [post]
func ExportData(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.DataExportReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.ExportData, &message.DataExportResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// ImportData
// @Summary import data to tidb cluster
// @Description import
// @Tags cluster import
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param dataImport body message.DataImportReq true "cluster info for import data"
// @Success 200 {object} controller.CommonResult{data=message.DataImportResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/import [post]
func ImportData(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &message.DataImportReq{}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.ImportData, &message.DataImportResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// QueryDataTransport
// @Summary query records of import and export
// @Description query records of import and export
// @Tags cluster data transport
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param dataTransportQueryReq query message.QueryDataImportExportRecordsReq true "transport records query condition"
// @Success 200 {object} controller.CommonResult{data=message.QueryDataImportExportRecordsResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/transport [get]
func QueryDataTransport(c *gin.Context) {
	var request message.QueryDataImportExportRecordsReq

	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &request); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.QueryDataTransport, &message.QueryDataImportExportRecordsResp{},
			requestBody,
			controller.DefaultTimeout)
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
// @Param DataTransportDeleteReq body message.DeleteImportExportRecordReq true "data transport record delete request"
// @Success 200 {object} controller.CommonResult{data=message.DeleteImportExportRecordResp}
// @Failure 401 {object} controller.CommonResult
// @Failure 403 {object} controller.CommonResult
// @Failure 500 {object} controller.CommonResult
// @Router /clusters/transport/{recordId} [delete]
func DeleteDataTransportRecord(c *gin.Context) {
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &message.DeleteImportExportRecordReq{
		RecordID: c.Param("recordId"),
	}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteDataTransportRecord, &message.DeleteImportExportRecordResp{},
			requestBody,
			controller.DefaultTimeout)
	}
}
