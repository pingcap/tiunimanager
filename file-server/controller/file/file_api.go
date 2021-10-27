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
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pingcap-inc/tiem/file-server/service"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"net/http"
	"os"
	"path/filepath"
)

func UploadImportFile(c *gin.Context) {
	var req UploadRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		_ = c.Error(err)
		return
	}
	if req.ClusterId == "" {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, "invalid input empty clusterId"))
	}
	uploadPath := service.DirMgr.GetImportPath(req.ClusterId)
	err := service.FileMgr.UploadFile(c.Request, uploadPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
	}
	err = service.FileMgr.UnzipDir(filepath.Join(uploadPath, service.DefaultDataFile) + "/data.zip", uploadPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
	}

	c.JSON(http.StatusOK, controller.Success(nil))
}

func DownloadExportFile(c *gin.Context) {
	clusterId := c.Param("clusterId")
	downloadPath := service.DirMgr.GetExportPath(clusterId)
	err := service.FileMgr.ZipDir(downloadPath + "/data", filepath.Join(downloadPath, service.DefaultDataFile))
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	filePath := filepath.Join(downloadPath, service.DefaultDataFile)
	_, err = os.Stat(filePath)
	if err != nil && !os.IsExist(err) {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+"data.zip")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	c.File(filePath)
}

