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
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/structs"
	"github.com/pingcap-inc/tiem/file-server/common"
	"github.com/pingcap-inc/tiem/file-server/controller"
	"github.com/pingcap-inc/tiem/file-server/service"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/client/cluster/clusterpb"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"net/http"
	"path/filepath"
)

func UploadImportFile(c *gin.Context) {
	ctx := framework.NewBackgroundMicroCtx(framework.NewMicroCtxFromGinCtx(c), true)
	clusterId := c.Request.FormValue("clusterId")
	if clusterId == "" {
		framework.LogWithContext(ctx).Errorf("invalid input empty clusterId")
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, "invalid input empty clusterId"))
		return
	}

	uploadPath, err := service.DirMgr.GetImportPath(ctx, clusterId)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get import path failed, %s", err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	err = service.FileMgr.UploadFile(ctx, c.Request, uploadPath)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("upload file failed, %s", err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	err = service.FileMgr.UnzipDir(ctx, filepath.Join(uploadPath, constants.DefaultZipName), filepath.Join(uploadPath, "data"))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("unzip zipfile failed, %s", err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, controller.Success(uploadPath))
}

func DownloadExportFile(c *gin.Context) {
	ctx := framework.NewBackgroundMicroCtx(framework.NewMicroCtxFromGinCtx(c), true)
	recordId := c.Param("recordId")

	request := &message.QueryDataImportExportRecordsReq{
		RecordID: recordId,
		PageRequest: structs.PageRequest{
			Page:     1,
			PageSize: 10,
		},
	}
	if err := c.ShouldBindQuery(&request); err != nil {
		framework.LogWithContext(ctx).Errorf("parse parameter error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body, err := json.Marshal(request)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("marshal request error: %s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rpcResp, err := client.ClusterClient.QueryDataTransport(ctx, &clusterpb.RpcRequest{Request: string(body)}, controller.DefaultTimeout)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query data transport record by recordId %s failed, %s", recordId, err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("find record from metadb failed, %s", err.Error())))
		return
	}
	var resp message.QueryDataImportExportRecordsResp
	err = json.Unmarshal([]byte(rpcResp.Response), &resp)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("unmarshal query data transport response failed, %s", err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("json unmarshal response failed, %s", err.Error())))
		return
	}
	if len(resp.Records) == 0 {
		framework.LogWithContext(ctx).Errorf("query data transport response empty")
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("can not found record %s", recordId)))
		return
	}
	record := resp.Records[0]
	framework.LogWithContext(ctx).Info(record)
	if record.RecordID == "" {
		framework.LogWithContext(ctx).Errorf("query data transport response recordId empty")
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("data transport recordId %s not exist", recordId)))
		return
	}
	if record.StorageType != string(constants.StorageTypeNFS) {
		framework.LogWithContext(ctx).Errorf("query data transport record stroage type is not %s", string(constants.StorageTypeNFS))
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("storage type %s can not download", record.StorageType)))
		return
	}
	if record.TransportType != string(constants.TransportTypeExport) {
		framework.LogWithContext(ctx).Errorf("query data transport record transport type is not %s", string(constants.TransportTypeExport))
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("transport type %s can not download", record.TransportType)))
		return
	}

	downloadPath := record.FilePath
	zipName := record.ZipName
	filePath := filepath.Join(filepath.Dir(downloadPath), zipName)

	totalSize, err := service.DirMgr.DirSizeB(downloadPath)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get total file size of path %s failed, %s", downloadPath, err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	if totalSize > common.MaxFileSize {
		framework.LogWithContext(ctx).Errorf("total file size %d byte reaches max download size %d byte", totalSize, common.MaxFileSize)
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, fmt.Sprintf("total file size %d byte reaches max download size %d byte", totalSize, common.MaxFileSize)))
		return
	}

	err = service.FileMgr.ZipDir(ctx, downloadPath, filePath)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("zip file %s failed, %s", filePath, err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	err = service.FileMgr.DownloadFile(ctx, c, filePath)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("download zip file %s failed, %s", filePath, err.Error())
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
}
