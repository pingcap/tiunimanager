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
	"time"

	"github.com/pingcap/tiunimanager/common/structs"
)

const (
	DEFAULTPAGESIZE = 100
)

type BasicRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CreateSQLFileRes struct {
	ID string `json:"id"`
}

type UpdateSQLFileRes struct {
}

type ShowSQLFileRes struct {
	SqlEditorFile
}
type DeleteSQLFileRes struct {
}
type ListSQLFileRes struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	List     []SqlEditorFile `json:"list"`
}

type StatementParam struct {
	ClusterID string `json:"clusterId" swaggerignore:"true"`
	SessionId string `json:"sessionid"`
	Sql       string `json:"sql"`
}

type ShowClusterMetaReq struct {
	IsBrief      bool   `json:"isbrief" form:"isBrief"`
	ClusterID    string `json:"clusterID"`
	ShowSystemDB bool   `json:"showSystemDB" form:"showSystemDB"`
}

type ShowTableMetaReq struct {
	ClusterID string `json:"clusterID"`
	DbName    string `json:"dbName"`
	TableName string `json:"tableName"`
}

type CreateSessionReq struct {
	ClusterID string `json:"clusterID"`
	Database  string `json:"database"`
}

type CloseSessionReq struct {
	ClusterID string `json:"clusterID"`
	SessionID string `json:"sessionID"`
}

type DataRes struct {
	Columns     []*Columns `json:"columns"`
	Rows        [][]string `json:"rows"`
	StartTime   string     `json:"start_time"`
	EndTime     string     `json:"end_time"`
	ExecuteTime string     `json:"execute_time"`
	RowCount    int64      `json:"row_count"`
	Limit       int        `json:"limit"`
	Query       string     `json:"query"`
}

type DataResOther struct {
	Rows        string `json:"rows"`
	ExecuteTime string `json:"execute_time"`
}

type Columns struct {
	Col      string `json:"col"`
	DataType string `json:"data_type"`
	Nullable bool   `json:"nullable"`
}

type SqlEditorMessage struct {
	ID           uint64    `json:"id"`
	ClusterId    uint64    `json:"cluster_id"`
	Database     string    `json:"database"`
	Sql          string    `json:"sql"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message"`
	Duration     string    `json:"duration"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type MetaRes struct {
	ClusterId string     `json:"clusterid"`
	Database  string     `json:"database"`
	Table     string     `json:"table"`
	Columns   []*Columns `json:"columns"`
}

type SqlEditorFile struct {
	ID        string    `json:"id"`
	ClusterID string    `json:"cluster_id"`
	Database  string    `json:"database"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	IsDeleted int       `json:"is_deleted"`
	CreatedBy string    `json:"created_by"`
	UpdatedBy string    `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SQLFileReq struct {
	Database  string `json:"database"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	ClusterID string `json:"cluster_id,omitempty"`
}
type SQLFileUpdateReq struct {
	SqlEditorFileID string `json:"sqlEditorFileID,omitempty"`
	Database        string `json:"database"`
	Name            string `json:"name"`
	Content         string `json:"content"`
	ClusterID       string `json:"cluster_id,omitempty"`
}

type SQLFileDeleteReq struct {
	SqlEditorFileID string `json:"sqlEditorFileID,omitempty"`
	ClusterID       string `json:"cluster_id,omitempty"`
}

type ShowSQLFileReq struct {
	SqlEditorFileID string `json:"sqlEditorFileID,omitempty"`
	ClusterID       string `json:"cluster_id,omitempty"`
}

type ListSQLFileReq struct {
	ClusterID string `json:"cluster_id,omitempty"`
	structs.PageRequest
}

type DBMeta struct {
	Name   string        `json:"name"`
	Tables []*TablesMeta `json:"tables"`
}

type TablesMeta struct {
	Name    string     `json:"name"`
	Columns []*Columns `json:"columns"`
}

type StatementsRes struct {
	Data *DataRes `json:"data"`
}

type CreateSessionRes struct {
	SessionID string `json:"sessionID"`
}
type CloseSessionRes struct {
}
