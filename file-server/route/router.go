/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package route

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/file-server/controller"
	files "github.com/pingcap/tiunimanager/file-server/controller/file"
	"github.com/pingcap/tiunimanager/file-server/interceptor"
	swaggerFiles "github.com/swaggo/files" // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Route(g *gin.Engine) {
	//check system
	check := g.Group("/system")
	{
		check.GET("/check", controller.Hello)
	}

	// support swagger
	swagger := g.Group("/swagger")
	{
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// api
	apiV1 := g.Group("/fs/api/v1")
	{
		apiV1.Use(interceptor.GinOpenTracing())
		apiV1.Use(interceptor.GinTraceIDHandler())
		apiV1.Use(interceptor.AccessLog(), gin.Recovery())

		file := apiV1.Group("/file")
		{
			file.Use(interceptor.VerifyIdentity)
			file.Use(interceptor.AuditLog())

			file.POST("/import/upload", files.UploadImportFile)
			file.GET("/export/download/:recordId", files.DownloadExportFile)
		}
	}

}
