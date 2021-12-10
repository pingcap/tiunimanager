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

/*******************************************************************************
 * @File: lib_br_v2
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/8
*******************************************************************************/

package secondparty

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pingcap-inc/tiem/library/framework"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
)

func (manager *SecondPartyManager) BackUp(ctx context.Context, cluster ClusterFacade, storage BrStorage, bizID string) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvbackup, clusterfacade: %v, storage: %v, bizid: %s", cluster, storage, bizID)
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupOperatorRecord(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var backupReq CmdBackUpReq
		backupReq.TaskID = rsp.Id
		backupReq.DbConnParameter = cluster.DbConnParameter
		backupReq.DbName = cluster.DbName
		backupReq.TableName = cluster.TableName
		backupReq.StorageAddress = fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root)
		backupReq.RateLimitM = cluster.RateLimitM
		backupReq.Concurrency = cluster.Concurrency
		backupReq.CheckSum = cluster.CheckSum
		manager.startNewBrBackUpTaskThruSQL(ctx, backupReq.TaskID, &backupReq)
		return rsp.Id, nil
	}
}

func (manager *SecondPartyManager) startNewBrBackUpTaskThruSQL(ctx context.Context, taskID uint64, req *CmdBackUpReq) {
	go func() {
		var args []string
		args = append(args, "BACKUP")
		if len(req.TableName) != 0 {
			args = append(args, "TABLE", fmt.Sprintf("`%s`.`%s`", req.DbName, req.TableName))
		} else {
			args = append(args, "DATABASE")
			if len(req.DbName) != 0 {
				args = append(args, fmt.Sprintf("`%s`", req.DbName))
			} else {
				args = append(args, "*")
			}
		}
		args = append(args, "TO", fmt.Sprintf("'%s'", req.StorageAddress))
		if len(req.RateLimitM) != 0 {
			args = append(args, "RATE_LIMIT", "=", req.RateLimitM, "MB/SECOND")
		}
		if len(req.Concurrency) != 0 {
			args = append(args, "CONCURRENCY", "=", req.Concurrency)
		}
		if len(req.CheckSum) != 0 {
			args = append(args, "CHECKSUM", "=", req.CheckSum)
		}
		args = append(args, req.Flags...)
		<-manager.startNewBrTaskThruSQL(ctx, taskID, &req.DbConnParameter, strings.Join(args, " "))
	}()
}

func (manager *SecondPartyManager) ShowBackUpInfo(ctx context.Context, cluster ClusterFacade) CmdShowBackUpInfoResp {
	framework.LogWithContext(ctx).Infof("microsrvshowbackupinfo, clusterfacade: %v", cluster)
	var showBackUpInfoReq CmdShowBackUpInfoReq
	showBackUpInfoReq.DbConnParameter = cluster.DbConnParameter
	showBackUpInfoResp := manager.startNewBrShowBackUpInfoThruSQL(ctx, &showBackUpInfoReq)
	return showBackUpInfoResp
}

func (manager *SecondPartyManager) startNewBrShowBackUpInfoThruSQL(ctx context.Context, req *CmdShowBackUpInfoReq) (resp CmdShowBackUpInfoResp) {
	brSQLCmd := "SHOW BACKUPS"
	dbConnParam := req.DbConnParameter
	framework.LogWithContext(ctx).Info("task start processing:", fmt.Sprintf("brSQLCmd:%s", brSQLCmd))
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.IP, dbConnParam.Port))
	if err != nil {
		resp.ErrorStr = err.Error()
		return
	}
	defer db.Close()

	resp = execShowBackUpInfoThruSQL(ctx, db, brSQLCmd)
	return
}

func (manager *SecondPartyManager) Restore(ctx context.Context, cluster ClusterFacade, storage BrStorage, bizID string) (taskID uint64, err error) {
	framework.LogWithContext(ctx).WithField("bizid", bizID).Infof("microsrvrestore, clusterfacade: %v, storage: %v, bizid: %s", cluster, storage, bizID)
	var req dbPb.CreateTiupOperatorRecordRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupOperatorRecord(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%v", rsp, err)
		return 0, err
	} else {
		var restoreReq CmdRestoreReq
		restoreReq.TaskID = rsp.Id
		restoreReq.DbConnParameter = cluster.DbConnParameter
		restoreReq.DbName = cluster.DbName
		restoreReq.TableName = cluster.TableName
		restoreReq.StorageAddress = fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root)
		restoreReq.RateLimitM = cluster.RateLimitM
		restoreReq.Concurrency = cluster.Concurrency
		restoreReq.CheckSum = cluster.CheckSum
		manager.startNewBrRestoreTaskThruSQL(ctx, restoreReq.TaskID, &restoreReq)
		return rsp.Id, nil
	}
}

func (manager *SecondPartyManager) startNewBrRestoreTaskThruSQL(ctx context.Context, taskID uint64, req *CmdRestoreReq) {
	go func() {
		var args []string
		args = append(args, "RESTORE")
		if len(req.TableName) != 0 {
			args = append(args, "TABLE", fmt.Sprintf("`%s`.`%s`", req.DbName, req.TableName))
		} else {
			args = append(args, "DATABASE")
			if len(req.DbName) != 0 {
				args = append(args, fmt.Sprintf("`%s`", req.DbName))
			} else {
				args = append(args, "*")
			}
		}
		args = append(args, "FROM", fmt.Sprintf("'%s'", req.StorageAddress))
		if len(req.RateLimitM) != 0 {
			args = append(args, "RATE_LIMIT", "=", req.RateLimitM, "MB/SECOND")
		}
		if len(req.Concurrency) != 0 {
			args = append(args, "CONCURRENCY", "=", req.Concurrency)
		}
		if len(req.CheckSum) != 0 {
			args = append(args, "CHECKSUM", "=", req.CheckSum)
		}
		args = append(args, req.Flags...)
		<-manager.startNewBrTaskThruSQL(ctx, taskID, &req.DbConnParameter, strings.Join(args, " "))
	}()
}

func (manager *SecondPartyManager) ShowRestoreInfo(ctx context.Context, cluster ClusterFacade) CmdShowRestoreInfoResp {
	framework.LogWithContext(ctx).Infof("microsrvshowrestoreinfo, clusterfacade: %v", cluster)
	var showRestoreInfoReq CmdShowRestoreInfoReq
	showRestoreInfoReq.DbConnParameter = cluster.DbConnParameter
	showRestoreInfoResp := manager.startNewBrShowRestoreInfoThruSQL(ctx, &showRestoreInfoReq)
	return showRestoreInfoResp
}

func (manager *SecondPartyManager) startNewBrShowRestoreInfoThruSQL(ctx context.Context, req *CmdShowRestoreInfoReq) (resp CmdShowRestoreInfoResp) {
	brSQLCmd := "SHOW RESTORES"
	dbConnParam := req.DbConnParameter
	framework.LogWithContext(ctx).Infof("task start processing: brSQLCmd:%s", brSQLCmd)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.IP, dbConnParam.Port))
	if err != nil {
		resp.ErrorStr = err.Error()
		return
	}
	defer db.Close()

	resp = execShowRestoreInfoThruSQL(ctx, db, brSQLCmd)
	return
}

func (manager *SecondPartyManager) startNewBrTaskThruSQL(ctx context.Context, taskID uint64, dbConnParam *DbConnParam, brSQLCmd string) (exitCh chan struct{}) {
	exitCh = make(chan struct{})
	logInFunc := framework.LogWithContext(ctx).WithField("task", taskID)
	logInFunc.Infof("task start processing: brSQLCmd:%s", brSQLCmd)
	manager.taskStatusCh <- TaskStatusMember{
		TaskID:   taskID,
		Status:   TaskStatusProcessing,
		ErrorStr: "",
	}
	go func() {
		defer close(exitCh)

		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.IP, dbConnParam.Port))
		if err != nil {
			manager.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintln(err),
			}
			return
		}
		defer db.Close()
		t0 := time.Now()
		resp := CmdBrResp{}
		err = db.QueryRow(brSQLCmd).Scan(&resp.Destination, &resp.Size, &resp.BackupTS, &resp.Queue_time, &resp.Execution_Time)
		if err != nil {
			logInFunc.Error("query sql cmd err", err)
			manager.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintln(err),
			}
			return
		}
		successFp := func() {
			logInFunc.Info("task finished, time cost", time.Since(t0))
			manager.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusFinished,
				ErrorStr: string(jsonMustMarshal(&resp)),
			}
		}
		logInFunc.Info("sql cmd return successfully")
		successFp()
	}()
	return exitCh
}
