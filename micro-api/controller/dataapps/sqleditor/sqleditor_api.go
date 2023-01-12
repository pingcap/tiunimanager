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
)

// CreateSession
// @Summary create session
// @Schemes
// @Description create session result
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param request_body body sqleditor.SessionParam true "request_body"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.SessionRes
// @Router /clusters/{cluster_id}/session [post]
func CreateSession(c *gin.Context) {

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
// @Router /clusters/{cluster_id}/session/{session_id} [delete]
func CloseSession(c *gin.Context) {

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
// @Router /clusters/{cluster_id}/statements [post]
func Statements(c *gin.Context) {

}

// Show All DB Meta godoc
// @Summary show all db meta
// @Schemes
// @Description show all db meta
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param isbrief query string true "value: true or false"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.ClusterMetaRes
// @Router /clusters/{cluster_id}/meta [get]
func ShowClusterMeta(c *gin.Context) {

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
// @Success 200 {object} sqleditor.ShowTableMetaRes
// @Router /clusters/{cluster_id}/dbs/{db_name}/{table_name}/meta [get]
func ShowTableMeta(c *gin.Context) {

}

// Create SQL File godoc
// @Summary create sql file
// @Description create sql file
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param request_body body sqleditor.SqlEditorFile true "request_body"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.CreateSQLFileRes
// @Router /clusters/{cluster_id}/sqlfiles [post]
func CreateSQLFile(c *gin.Context) {

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
// @Router /clusters/{cluster_id}/sqlfiles/{sql_file_id} [get]
func ShowSQLFile(c *gin.Context) {

}

// Update SQL File godoc
// @Summary update sql file
// @Description update sql file
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Param sql_file_id path string true "sql file ID"
// @Param request_body body sqleditor.SqlEditorFile true "request_body"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.UpdateSQLFileRes
// @Router /clusters/{cluster_id}/sqlfiles/{sql_file_id} [put]
func UpdateSQLFile(c *gin.Context) {

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
// @Router /clusters/{cluster_id}/sqlfiles/{sql_file_id} [delete]
func DeleteSQLFile(c *gin.Context) {

}

// List SQL File godoc
// @Summary list sql file
// @Description list sql file
// @Tags sqleditor
// @Param cluster_id path string true "cluster ID"
// @Accept json
// @Produce json
// @Success 200 {object} sqleditor.ListSQLFileRes
// @Router /clusters/{cluster_id}/sqlfiles [get]
func ListSQLFile(c *gin.Context) {

}
