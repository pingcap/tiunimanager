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

import "time"

type BasicRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type CreateSQLFileRes struct {
	BasicRes
	Data *uint64 `json:"data"`
}

type UpdateSQLFileRes struct {
	BasicRes
	Data string `json:"data"`
}

type ShowSQLFileRes struct {
	BasicRes
	Data SqlEditorFile `json:"data"`
}
type DeleteSQLFileRes struct {
	BasicRes
	Data string `json:"data"`
}
type ListSQLFileRes struct {
	BasicRes
	Data []SqlEditorFile `json:"data"`
}

type StatementParam struct {
	SessionId string `json:"sessionid"`
	Sql       string `json:"sql"`
}

type ShowClusterMetaParam struct {
	IsBrief int `json:"isbrief"`
}

type SessionParam struct {
	Database string `json:"database"`
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
	ClusterId uint64     `json:"clusterid"`
	Database  string     `json:"database"`
	Table     string     `json:"table"`
	Columns   []*Columns `json:"columns"`
}

type SqlEditorFile struct {
	ID        uint64    `json:"id"`
	ClusterId uint64    `json:"cluster_id"`
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
	Database string `json:"database"`
	Name     string `json:"name"`
	Content  string `json:"content"`
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
	BasicRes
	Data *DataRes `json:"data"`
}

type ClusterMetaRes struct {
	BasicRes
	Data []*DBMeta `json:"data"`
}

type ShowDBsRes struct {
	BasicRes
	Data []string `json:"data"`
}

type ShowTablesRes struct {
	BasicRes
	Data []string `json:"data"`
}

type ShowTableMetaRes struct {
	BasicRes
	Data *MetaRes `json:"data"`
}

type SessionRes struct {
	BasicRes
	Data string `json:"data"`
}
