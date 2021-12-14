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
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"net/http"
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
	var request message.DataExportReq

	if err := c.ShouldBindQuery(&request); err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controller.InvokeRpcMethod(c, client.ClusterClient.ExportData, &message.DataExportResp{}, string(body), controller.DefaultTimeout)
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
	var request message.DataImportReq

	if err := c.ShouldBindQuery(&request); err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controller.InvokeRpcMethod(c, client.ClusterClient.ImportData, &message.DataImportResp{}, string(body), controller.DefaultTimeout)
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

	if err := c.ShouldBindQuery(&request); err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controller.InvokeRpcMethod(c, client.ClusterClient.QueryDataTransport, &message.QueryDataImportExportRecordsResp{}, string(body), controller.DefaultTimeout)
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
	var request message.DeleteImportExportRecordReq

	if err := c.ShouldBindQuery(&request); err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request.RecordID = c.Param("recordId")
	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(c).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	controller.InvokeRpcMethod(c, client.ClusterClient.DeleteDataTransportRecord, &message.DeleteImportExportRecordResp{}, string(body), controller.DefaultTimeout)
}
