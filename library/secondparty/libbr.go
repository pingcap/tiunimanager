package secondparty

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
)

type DbConnParam struct {
	Username string
	Password string
	Ip       string
	Port     string
}

type StorageType string

const (
	StorageTypeLocal StorageType = "local"
	StorageTypeS3    StorageType = "s3"
)

type ClusterFacade struct {
	TaskID          uint64 // do not pass this value for br command
	DbConnParameter DbConnParam
	DbName          string
	TableName       string
	ClusterId       string // todo: need to know the usage
	ClusterName     string // todo: need to know the usage
	//PdAddress 			string
	RateLimitM      string
	Concurrency     string
	CheckSum        string
}

type BrStorage struct {
	StorageType StorageType
	Root        string // "/tmp/backup"
}

func (secondMicro *SecondMicro) BackUp(cluster ClusterFacade, storage BrStorage, bizId uint64) (taskID uint64, err error) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Backup
	req.BizID = bizId
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
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
		secondMicro.startNewBrBackUpTaskThruSQL(backupReq.TaskID, &backupReq)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewBrBackUpTaskThruSQL(taskID uint64, req *CmdBackUpReq) {
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
		<-secondMicro.startNewBrTaskThruSQL(taskID, &req.DbConnParameter, strings.Join(args, " "))
	}()
}

func (secondMicro *SecondMicro) ShowBackUpInfo(cluster ClusterFacade) CmdShowBackUpInfoResp {
	var showBackUpInfoReq CmdShowBackUpInfoReq
	showBackUpInfoReq.DbConnParameter = cluster.DbConnParameter
	showBackUpInfoResp := secondMicro.startNewBrShowBackUpInfoThruSQL(&showBackUpInfoReq)
	return showBackUpInfoResp
}

func (secondMicro *SecondMicro) startNewBrShowBackUpInfoThruSQL(req *CmdShowBackUpInfoReq) (resp CmdShowBackUpInfoResp) {
	brSQLCmd := "SHOW BACKUPS"
	dbConnParam := req.DbConnParameter
	logger.Info("task start processing:", fmt.Sprintf("brSQLCmd:%s", brSQLCmd))
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.Ip, dbConnParam.Port))
	if err != nil {
		resp.ErrorStr = err.Error()
		return
	}
	defer db.Close()

	resp = execShowBackUpInfoThruSQL(db, brSQLCmd)
	return
}

func execShowBackUpInfoThruSQL(db *sql.DB, showBackupSQLCmd string) (resp CmdShowBackUpInfoResp) {
	t0 := time.Now()
	err := db.QueryRow(showBackupSQLCmd).Scan(&resp.Destination, &resp.State, &resp.Progress, &resp.Queue_time, &resp.Execution_Time, &resp.Finish_Time, &resp.Connection)
	successFp := func() {
		logger.Info("showbackupinfo task finished, time cost", time.Now().Sub(t0))
	}
	if err != nil {
		logger.Errorf("query sql cmd err: %v", err)
		if err.Error() != "sql: no rows in result set" {
			logger.Debugf("(%s) != (sql: no rows in result set", err.Error())
			resp.ErrorStr = err.Error()
			return
		}
		logger.Debugf("(%s) == (sql: no rows in result set)", err.Error())
		logger.Infof("task has finished without checking db while no rows is result for sql cmd")
		resp.Progress = 100
		return
	}
	logger.Info("sql cmd return successfully")
	successFp()
	return
}

func (secondMicro *SecondMicro) Restore(cluster ClusterFacade, storage BrStorage, bizId uint64) (taskID uint64, err error) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Restore
	req.BizID = bizId
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
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
		secondMicro.startNewBrRestoreTaskThruSQL(restoreReq.TaskID, &restoreReq)
		return rsp.Id, nil
	}
}

func (secondMicro *SecondMicro) startNewBrRestoreTaskThruSQL(taskID uint64, req *CmdRestoreReq) {
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
		<-secondMicro.startNewBrTaskThruSQL(taskID, &req.DbConnParameter, strings.Join(args, " "))
	}()
}

func (secondMicro *SecondMicro) ShowRestoreInfo(cluster ClusterFacade) CmdShowRestoreInfoResp {
	var showRestoreInfoReq CmdShowRestoreInfoReq
	showRestoreInfoReq.DbConnParameter = cluster.DbConnParameter
	showRestoreInfoResp := secondMicro.startNewBrShowRestoreInfoThruSQL(&showRestoreInfoReq)
	return showRestoreInfoResp
}

func (secondMicro *SecondMicro) startNewBrShowRestoreInfoThruSQL(req *CmdShowRestoreInfoReq) (resp CmdShowRestoreInfoResp) {
	brSQLCmd := "SHOW RESTORES"
	dbConnParam := req.DbConnParameter
	logger.Infof("task start processing: brSQLCmd:%s", brSQLCmd)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.Ip, dbConnParam.Port))
	if err != nil {
		resp.ErrorStr = err.Error()
		return
	}
	defer db.Close()

	resp = execShowRestoreInfoThruSQL(db, brSQLCmd)
	return
}

func execShowRestoreInfoThruSQL(db *sql.DB, showRestoreSQLCmd string) (resp CmdShowRestoreInfoResp) {
	t0 := time.Now()
	err := db.QueryRow(showRestoreSQLCmd).Scan(&resp.Destination, &resp.State, &resp.Progress, &resp.Queue_time, &resp.Execution_Time, &resp.Finish_Time, &resp.Connection)
	successFp := func() {
		logger.Info("showretoreinfo task finished, time cost", time.Now().Sub(t0))
	}
	if err != nil {
		logger.Errorf("query sql cmd err: %v", err)
		if err.Error() != "sql: no rows in result set" {
			logger.Debugf("(%s) != (sql: no rows in result set", err.Error())
			resp.ErrorStr = err.Error()
			return resp
		}
		logger.Debugf("(%s) == (sql: no rows in result set)", err.Error())
		logger.Infof("task has finished without checking db while no rows is result for sql cmd")
		resp.Progress = 100
		return resp
	}
	logger.Info("sql cmd return successfully")
	successFp()
	return resp
}

func (secondMicro *SecondMicro) startNewBrTaskThruSQL(taskID uint64, dbConnParam *DbConnParam, brSQLCmd string) (exitCh chan struct{}) {
	exitCh = make(chan struct{})
	logger.Infof("task start processing: brSQLCmd:%s", brSQLCmd)
	secondMicro.taskStatusCh <- TaskStatusMember{
		TaskID:   taskID,
		Status:   TaskStatusProcessing,
		ErrorStr: "",
	}
	go func() {
		defer close(exitCh)

		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.Ip, dbConnParam.Port))
		if err != nil {
			secondMicro.taskStatusCh <- TaskStatusMember{
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
			logger.Error("query sql cmd err", err)
			secondMicro.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintln(err),
			}
			return
		}
		successFp := func() {
			logger.Info("task finished, time cost", time.Now().Sub(t0))
			secondMicro.taskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusFinished,
				ErrorStr: string(jsonMustMarshal(&resp)),
			}
		}
		logger.Info("sql cmd return successfully")
		successFp()
		return
	}()
	return exitCh
}
