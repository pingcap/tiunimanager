/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"strings"
)

type BackupSQLReq struct {
	NodeID          string
	DbName          string
	TableName       string
	StorageAddress  string
	DbConnParameter DbConnParam // only for SQL command, not used in br command
	RateLimitM      string
	Concurrency     string // only for SQL command, not used in br command
	CheckSum        string // only for SQL command, not used in br command
}

type BRSQLResp struct {
	Destination   string
	Size          uint64
	BackupTS      uint64
	QueueTime     string
	ExecutionTime string
}

type RestoreSQLReq struct {
	NodeID          string
	DbName          string
	TableName       string
	StorageAddress  string
	DbConnParameter DbConnParam // only for SQL command, not used in br command
	RateLimitM      string
	Concurrency     string // only for SQL command, not used in br command
	CheckSum        string // only for SQL command, not used in br command
}

func ExecBackupSQL(ctx context.Context, request BackupSQLReq, workflowNodeId string) (resp BRSQLResp, err error) {
	framework.LogWithContext(ctx).Infof("begin exec backup sql, request: %+v, bizId: %s", request, workflowNodeId)

	dbConnParam := request.DbConnParameter
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username,
		dbConnParam.Password, dbConnParam.IP, dbConnParam.Port))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("open tidb connection failed %s", err.Error())
		return
	}
	defer db.Close()

	var args []string
	args = append(args, "BACKUP")
	if len(request.TableName) != 0 {
		args = append(args, "TABLE", fmt.Sprintf("`%s`.`%s`", request.DbName, request.TableName))
	} else {
		args = append(args, "DATABASE")
		if len(request.DbName) != 0 {
			args = append(args, fmt.Sprintf("`%s`", request.DbName))
		} else {
			args = append(args, "*")
		}
	}
	args = append(args, "TO", fmt.Sprintf("'%s'", request.StorageAddress))
	if len(request.RateLimitM) != 0 {
		args = append(args, "RATE_LIMIT", "=", request.RateLimitM, "MB/SECOND")
	}
	if len(request.Concurrency) != 0 {
		args = append(args, "CONCURRENCY", "=", request.Concurrency)
	}
	if len(request.CheckSum) != 0 {
		args = append(args, "CHECKSUM", "=", request.CheckSum)
	}
	brSQLCmd := strings.Join(args, " ")
	err = db.QueryRow(brSQLCmd).Scan(&resp.Destination, &resp.Size, &resp.BackupTS, &resp.QueueTime,
		&resp.ExecutionTime)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query backup sql cmd failed %s", err.Error())
		return
	}
	framework.LogWithContext(ctx).Info("do backup sql cmd %s succeed", brSQLCmd)
	return
}

func ExecRestoreSQL(ctx context.Context, request RestoreSQLReq, workflowNodeId string) (resp BRSQLResp, err error) {
	framework.LogWithContext(ctx).Infof("begin exec restore sql, request: %+v, bizId: %s", request, workflowNodeId)

	dbConnParam := request.DbConnParameter
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username,
		dbConnParam.Password, dbConnParam.IP, dbConnParam.Port))
	if err != nil {
		framework.LogWithContext(ctx).Errorf("open tidb connection failed %s", err.Error())
		return
	}
	defer db.Close()

	var args []string
	args = append(args, "RESTORE")
	if len(request.TableName) != 0 {
		args = append(args, "TABLE", fmt.Sprintf("`%s`.`%s`", request.DbName, request.TableName))
	} else {
		args = append(args, "DATABASE")
		if len(request.DbName) != 0 {
			args = append(args, fmt.Sprintf("`%s`", request.DbName))
		} else {
			args = append(args, "*")
		}
	}
	args = append(args, "FROM", fmt.Sprintf("'%s'", request.StorageAddress))
	if len(request.RateLimitM) != 0 {
		args = append(args, "RATE_LIMIT", "=", request.RateLimitM, "MB/SECOND")
	}
	if len(request.Concurrency) != 0 {
		args = append(args, "CONCURRENCY", "=", request.Concurrency)
	}
	if len(request.CheckSum) != 0 {
		args = append(args, "CHECKSUM", "=", request.CheckSum)
	}
	brSQLCmd := strings.Join(args, " ")
	err = db.QueryRow(brSQLCmd).Scan(&resp.Destination, &resp.Size, &resp.BackupTS, &resp.QueueTime,
		&resp.ExecutionTime)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("query restore sql cmd failed %s", err.Error())
		return
	}
	framework.LogWithContext(ctx).Info("do restore sql cmd %s succeed", brSQLCmd)
	return
}
