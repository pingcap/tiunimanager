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
	"github.com/pingcap-inc/tiem/file-server/service"
	"github.com/pingcap-inc/tiem/micro-api/controller"
	"net/http"
)

func UploadFile(c *gin.Context) {
	err := service.FileMgr.UploadFile(c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, controller.Fail(http.StatusBadRequest, err.Error()))
	} else {
		c.JSON(http.StatusOK, controller.Success(nil))
	}
}

func DownloadFile(c *gin.Context) {
	//todo
}

