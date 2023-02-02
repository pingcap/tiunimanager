/******************************************************************************
 * Copyright (c)  2023 PingCAP, Inc.                                          *
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
 ******************************************************************************/

package sqleditor

import (
	"github.com/gin-gonic/gin"
	"github.com/pingcap/tiunimanager/common/client"
	"github.com/pingcap/tiunimanager/message/dataapps/sqleditor"
	"github.com/pingcap/tiunimanager/micro-api/controller"
)

// CreateSession
// @Summary create session
// @Schemes
// @Description create session result
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param request_body body sqleditor.CreateSessionReq true "request_body"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.CreateSessionRes
// @Router /{cluster_id}/session [post]
func CreateSession(c *gin.Context) {
	req := sqleditor.CreateSessionReq{
		ClusterID: c.Param("clusterId"),
	}
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateSession, &sqleditor.CreateSessionRes{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Close Session godoc
// @Summary close session
// @Schemes
// @Description close session result
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param session_id path string true "session ID"
// @Accept json
// @Produce json
// @Success 200
// @Router /{cluster_id}/session/{session_id} [delete]
func CloseSession(c *gin.Context) {
	req := sqleditor.CloseSessionReq{
		ClusterID: c.Param("clusterId"),
		SessionID: c.Param("sessionId"),
	}
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CloseSession, &sqleditor.CloseSessionRes{},
			requestBody,
			controller.DefaultTimeout)
	}

}

// Execute SQL godoc
// @Summary execute sql
// @Schemes
// @Description execute sql result
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param request_body body sqleditor.StatementParam true "request_body"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.StatementsRes
// @Router /{cluster_id}/statements [post]
func Statements(c *gin.Context) {
	if body, ok := controller.HandleJsonRequestFromBody(c,
		&sqleditor.StatementParam{},
		func(c *gin.Context, req interface{}) error {
			req.(*sqleditor.StatementParam).ClusterID = c.Param("clusterId")
			return nil
		}); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.Statements,
			&sqleditor.StatementsRes{}, body, controller.DefaultTimeout)
	}
}

// Show All DB Meta godoc
// @Summary show all db meta
// @Schemes
// @Description show all db meta
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param isbrief query string true "value: true or false"
// @Param showSystemDB query string true "value: true or false"
// @Accept json
// @Produce json
// @Success 200 {object} []sqleditor.DBMeta
// @Router /{cluster_id}/meta [get]
func ShowClusterMeta(c *gin.Context) {
	req := sqleditor.ShowClusterMetaReq{
		ClusterID: c.Param("clusterId"),
	}
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.ShowClusterMeta, &[]sqleditor.DBMeta{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Show Table Meta godoc
// @Summary show table meta
// @Description table meta
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param db_name path string true "db name"
// @Param table_name path string true "table_name"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.MetaRes
// @Router /{cluster_id}/dbs/{db_name}/{table_name}/meta [get]
func ShowTableMeta(c *gin.Context) {
	req := sqleditor.ShowTableMetaReq{
		ClusterID: c.Param("clusterId"),
		DbName:    c.Param("dbName"),
		TableName: c.Param("tableName"),
	}
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.ShowTableMeta, &sqleditor.MetaRes{},
			requestBody,
			controller.DefaultTimeout)
	}

}

// Create SQL File godoc
// @Summary create sql file
// @Description create sql file
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param request_body body sqleditor.SQLFileReq true "request_body"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.CreateSQLFileRes
// @Router /{cluster_id}/sqlfiles [post]
func CreateSQLFile(c *gin.Context) {
	req := sqleditor.SQLFileReq{
		ClusterID: c.Param("clusterId"),
	}
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.CreateSQLFile, &sqleditor.CreateSQLFileRes{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Update SQL File godoc
// @Summary update sql file
// @Description update sql file
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param sql_file_id path string true "sql file ID"
// @Param request_body body sqleditor.SQLFileUpdateReq true "request_body"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.UpdateSQLFileRes
// @Router /{cluster_id}/sqlfiles/{sql_file_id} [put]
func UpdateSQLFile(c *gin.Context) {
	req := sqleditor.SQLFileUpdateReq{
		ClusterID:       c.Param("clusterId"),
		SqlEditorFileID: c.Param("sqlFileId"),
	}
	if requestBody, ok := controller.HandleJsonRequestFromBody(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.UpdateSQLFile, &sqleditor.UpdateSQLFileRes{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// Delete SQL File godoc
// @Summary delete sql file
// @Description delete sql file
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param sql_file_id path string true "sql file ID"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.DeleteSQLFileRes
// @Router /{cluster_id}/sqlfiles/{sql_file_id} [delete]
func DeleteSQLFile(c *gin.Context) {
	req := sqleditor.SQLFileDeleteReq{
		ClusterID:       c.Param("clusterId"),
		SqlEditorFileID: c.Param("sqlFileId"),
	}
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.DeleteSQLFile, &sqleditor.DeleteSQLFileRes{},
			requestBody,
			controller.DefaultTimeout)
	}
}

// List SQL File godoc
// @Summary list sql file
// @Description list sql file
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param queryReq query sqleditor.ListSQLFileReq false "query request"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.ListSQLFileRes
// @Router /{cluster_id}/sqlfiles [get]
func ListSQLFile(c *gin.Context) {
	req := sqleditor.ListSQLFileReq{
		ClusterID: c.Param("clusterId"),
	}
	if requestBody, ok := controller.HandleJsonRequestFromQuery(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.ListSQLFile, &sqleditor.ListSQLFileRes{},
			requestBody,
			controller.DefaultTimeout)
	}

}

// Show SQL File godoc
// @Summary show sql file
// @Description show sql file
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param sql_file_id path string true "sql file ID"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.ShowSQLFileRes
// @Router /{cluster_id}/sqlfiles/{sql_file_id} [get]
func ShowSQLFile(c *gin.Context) {
	req := sqleditor.ShowSQLFileReq{
		ClusterID:       c.Param("clusterId"),
		SqlEditorFileID: c.Param("sqlFileId"),
	}
	if requestBody, ok := controller.HandleJsonRequestWithBuiltReq(c, &req); ok {
		controller.InvokeRpcMethod(c, client.ClusterClient.ShowSQLFile, &sqleditor.ShowSQLFileRes{},
			requestBody,
			controller.DefaultTimeout)
	}
}
