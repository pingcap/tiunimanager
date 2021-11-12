/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package file

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/file-server/controller"
	"github.com/pingcap-inc/tiem/file-server/service"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbService "github.com/pingcap-inc/tiem/micro-metadb/service"
	"net/http"
	"path/filepath"
	"strconv"
)

func UploadImportFile(c *gin.Context) {
	clusterId := c.Request.FormValue("clusterId")
	if clusterId == "" {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, "invalid input empty clusterId"))
		return
	}

	uploadPath, err := service.DirMgr.GetImportPath(clusterId)
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	err = service.FileMgr.UploadFile(c.Request, uploadPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	err = service.FileMgr.UnzipDir(filepath.Join(uploadPath, common.DefaultZipName), filepath.Join(uploadPath, "data"))
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, controller.Success(nil))
}

func DownloadExportFile(c *gin.Context) {
	recordId, err := strconv.Atoi(c.Param("recordId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("input record id invalid, %s", err.Error())))
		return
	}

	queryReq := &dbpb.DBFindTransportRecordByIDRequest{
		RecordId: int64(recordId),
	}

	ctx := framework.NewMicroCtxFromGinCtx(c)
	resp, err := client.DBClient.FindTrasnportRecordByID(ctx, queryReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("find record from metadb failed, %s", err.Error())))
		return
	}
	if resp.GetStatus().GetCode() != dbService.ClusterSuccessResponseStatus.GetCode() || resp.GetRecord() == nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("find record from metadb failed, %s", resp.GetStatus().GetMessage())))
		return
	}
	framework.LogWithContext(ctx).Info(resp.GetRecord())
	record := resp.GetRecord()
	if record.GetStorageType() != common.NfsStorageType {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("storage type %s can not download", record.GetStorageType())))
		return
	}
	if record.GetTransportType() != string(common.TransportTypeExport) {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("transport type %s can not download", record.GetTransportType())))
		return
	}

	downloadPath := resp.GetRecord().GetFilePath()
	zipName := resp.GetRecord().GetZipName()
	filePath := filepath.Join(filepath.Dir(downloadPath), zipName)

	err = service.FileMgr.ZipDir(downloadPath, filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	err = service.FileMgr.DownloadFile(c, filePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
}
