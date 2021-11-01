
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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

package libbr

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/sirupsen/logrus"
)

type CmdTypeStr string

const (
	CmdBackUpPreCheckReqTypeStr    CmdTypeStr = "CmdBackUpPreCheckReq"
	CmdBackUpPreCheckRespTypeStr   CmdTypeStr = "CmdBackUpPreCheckResp"
	CmdBackUpReqTypeStr            CmdTypeStr = "CmdBackUpReq"
	CmdBackUpRespTypeStr           CmdTypeStr = "CmdBackUpResp"
	CmdShowBackUpInfoReqTypeStr    CmdTypeStr = "CmdShowBackUpInfoReq"
	CmdShowBackUpInfoRespTypeStr   CmdTypeStr = "CmdShowBackUpInfoResp"
	CmdRestorePreCheckReqTypeStr   CmdTypeStr = "CmdRestorePreCheckReq"
	CmdRestorePreCheckRespTypeStr  CmdTypeStr = "CmdRestorePreCheckResp"
	CmdRestoreReqTypeStr           CmdTypeStr = "CmdRestoreReq"
	CmdRestoreRespTypeStr          CmdTypeStr = "CmdRestoreResp"
	CmdShowRestoreInfoReqTypeStr   CmdTypeStr = "CmdShowRestoreInfoReq"
	CmdShowRestoreInfoRespTypeStr  CmdTypeStr = "CmdShowRestoreInfoResp"
	CmdGetAllTaskStatusReqTypeStr  CmdTypeStr = "CmdGetAllTaskStatusReq"
	CmdGetAllTaskStatusRespTypeStr CmdTypeStr = "CmdGetAllTaskStatusResp"
)

type CmdReqOrResp struct {
	TypeStr CmdTypeStr
	Content string
}

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

type CmdBackUpPreCheckReq struct {
}

type CmdBackUpPreCheckResp struct {
}

type CmdBackUpReq struct {
	TaskID            uint64
	DbName            string
	TableName         string
	FilterDbTableName string // used in br command, pending for use in SQL command
	StorageAddress    string
	DbConnParameter   DbConnParam // only for SQL command, not used in br command
	RateLimitM        string
	Concurrency       string   // only for SQL command, not used in br command
	CheckSum          string   // only for SQL command, not used in br command
	LogFile           string   // used in br command, pending for use in SQL command
	TimeoutS          int      // used in br command, pending for use in SQL command
	Flags             []string // used in br command, pending for use in SQL command
}

type CmdBrResp struct {
	Destination    string
	Size           uint64
	BackupTS       uint64
	Queue_time     string
	Execution_Time string
}

type CmdShowBackUpInfoReq struct {
	TaskID          uint64
	DbConnParameter DbConnParam
}

type CmdShowBackUpInfoResp struct {
	Destination    string
	Size           uint64
	BackupTS       uint64
	State          string
	Progress       float32
	Queue_time     string
	Execution_Time string
	Finish_Time    *string
	Connection     string
	Error          error
}

type CmdRestorePreCheckReq struct {
}

type CmdRestorePreCheckResp struct {
}

type CmdRestoreReq struct {
	TaskID            uint64
	DbName            string
	TableName         string
	FilterDbTableName string // used in br command, pending for use in SQL command
	StorageAddress    string
	DbConnParameter   DbConnParam // only for SQL command, not used in br command
	RateLimitM        string
	Concurrency       string   // only for SQL command, not used in br command
	CheckSum          string   // only for SQL command, not used in br command
	LogFile           string   // used in br command, pending for use in SQL command
	TimeoutS          int      // used in br command, pending for use in SQL command
	Flags             []string // used in br command, pending for use in SQL command
}

type CmdShowRestoreInfoReq struct {
	TaskID          uint64
	DbConnParameter DbConnParam
}

type CmdShowRestoreInfoResp struct {
	Destination    string
	Size           uint64
	BackupTS       uint64
	State          string
	Progress       float32
	Queue_time     string
	Execution_Time string
	Finish_Time    *string
	Connection     string
	Error          error
}

type TaskStatusMapValue struct {
	validFlag bool
	stat      TaskStatusMember
	readct    uint64
}

type TaskStatus int

const (
	TaskStatusInit       TaskStatus = TaskStatus(dbpb.TiupTaskStatus_Init)
	TaskStatusProcessing TaskStatus = TaskStatus(dbpb.TiupTaskStatus_Processing)
	TaskStatusFinished   TaskStatus = TaskStatus(dbpb.TiupTaskStatus_Finished)
	TaskStatusError      TaskStatus = TaskStatus(dbpb.TiupTaskStatus_Error)
)

type TaskStatusMember struct {
	TaskID   uint64
	Status   TaskStatus
	ErrorStr string
}

type CmdGetAllTaskStatusReq struct {
}

type CmdGetAllTaskStatusResp struct {
	Stats []TaskStatusMember
}

//var glMgrTaskStatusCh chan TaskStatusMember
//var glMgrTaskStatusMap map[uint64]TaskStatusMapValue

var logger *logrus.Entry

var glMgrTaskStatusCh chan TaskStatusMember
var glMgrTaskStatusMap map[uint64]TaskStatusMapValue

func BrMgrInit() {
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	logger = framework.LogForkFile(configPath + common.LogFileBrMgr)

	glMgrTaskStatusCh = make(chan TaskStatusMember, 1024)
	glMgrTaskStatusMap = make(map[uint64]TaskStatusMapValue)
}

func BrMgrRoutine() {
	inReader := bufio.NewReader(os.Stdin)
	outWriter := os.Stdout
	//errw := os.Stderr

	for {
		input, err := inReader.ReadString('\n')
		if err != nil {
			myPanic(err)
		}
		//errw.Write([]byte(input))
		if input[len(input)-1] == '\n' {
			cmdStr := input[:len(input)-1]
			var cmd CmdReqOrResp
			err := json.Unmarshal([]byte(cmdStr), &cmd)
			if err != nil {
				myPanic(fmt.Sprintln("cmdStr unmarshal failed err:", err, "cmdStr:", cmdStr))
			}
			logger.Info("rcv req", cmd)
			var cmdResp CmdReqOrResp
			switch cmd.TypeStr {
			//case CmdBackUpPreCheckReqTypeStr:
			case CmdBackUpReqTypeStr:
				resp := mgrHandleCmdBackUpReq(cmd.Content)
				cmdResp.TypeStr = CmdBackUpRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			case CmdShowBackUpInfoReqTypeStr:
				resp := mgrHandleCmdShowBackUpInfoReq(cmd.Content)
				cmdResp.TypeStr = CmdShowBackUpInfoRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			//case CmdRestorePreCheckReqTypeStr:
			case CmdRestoreReqTypeStr:
				resp := mgrHandleCmdRestoreReq(cmd.Content)
				cmdResp.TypeStr = CmdRestoreRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			case CmdShowRestoreInfoReqTypeStr:
				resp := mgrHandleCmdShowRestoreInfoReq(cmd.Content)
				cmdResp.TypeStr = CmdShowRestoreInfoRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			case CmdGetAllTaskStatusReqTypeStr:
				resp := mgrHandleCmdGetAllTaskStatusReq(cmd.Content)
				cmdResp.TypeStr = CmdGetAllTaskStatusRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			default:
				myPanic(fmt.Sprintln("unknown cmdStr.TypeStr:", cmd.TypeStr))
			}
			logger.Info("snd rsp", cmdResp)
			bs := jsonMustMarshal(&cmdResp)
			bs = append(bs, '\n')
			ct, err := outWriter.Write(bs)
			assert(ct == len(bs))
			assert(err == nil)
		} else {
			myPanic("unexpected")
		}
	}
}

func assert(b bool) {
	if b {
	} else {
		logger.Fatal("unexpected panic with stack trace:", string(debug.Stack()))
		panic("unexpected")
	}
}

func myPanic(v interface{}) {
	s := fmt.Sprint(v)
	logger.Fatalf("panic: %s, with stack trace: %s", s, string(debug.Stack()))
	panic("unexpected")
}

func jsonMustMarshal(v interface{}) []byte {
	bs, err := json.Marshal(v)
	assert(err == nil)
	return bs
}

func glMgrStatusMapSync() {
	for {
		var consumedFlag bool
		var statm TaskStatusMember
		var ok bool
		select {
		case statm, ok = <-glMgrTaskStatusCh:
			assert(ok)
			consumedFlag = true
		default:
		}
		if consumedFlag {
			v := glMgrTaskStatusMap[statm.TaskID]
			if v.validFlag {
				assert(v.stat.Status == TaskStatusProcessing)
				assert(statm.Status == TaskStatusFinished || statm.Status == TaskStatusError)
			} else {
				assert(statm.Status == TaskStatusProcessing)
			}
			glMgrTaskStatusMap[statm.TaskID] = TaskStatusMapValue{
				validFlag: true,
				readct:    0,
				stat:      statm,
			}
		} else {
			break
		}
	}
}

func glMgrStatusMapGetAll() (ret []TaskStatusMember) {
	var needDeletTaskList []uint64
	for k, v := range glMgrTaskStatusMap {
		if v.readct > 0 && (v.stat.Status == TaskStatusFinished || v.stat.Status == TaskStatusError) {
			needDeletTaskList = append(needDeletTaskList, k)
		}
	}
	for _, k := range needDeletTaskList {
		delete(glMgrTaskStatusMap, k)
	}
	for k, v := range glMgrTaskStatusMap {
		assert(k == v.stat.TaskID)
		ret = append(ret, v.stat)
	}
	for k, v := range glMgrTaskStatusMap {
		newv := v
		newv.readct++
		glMgrTaskStatusMap[k] = newv
	}
	return
}

func mgrHandleCmdBackUpReq(jsonStr string) CmdBrResp {
	ret := CmdBrResp{}
	var req CmdBackUpReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		myPanic(fmt.Sprintln("json.unmarshal cmdbackupreq failed err:", err))
	}
	//mgrStartNewBrBackUpTask(req.TaskID, &req)
	mgrStartNewBrBackUpTaskThruSQL(req.TaskID, &req)
	return ret
}

func mgrHandleCmdShowBackUpInfoReq(jsonStr string) CmdShowBackUpInfoResp {
	var ret CmdShowBackUpInfoResp
	var req CmdShowBackUpInfoReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		myPanic(fmt.Sprintln("json.unmarshal cmdshowbackupinforeq failed err:", err))
	}
	ret = mgrStartNewBrShowBackUpInfoThruSQL(&req)
	return ret
}

func mgrHandleCmdRestoreReq(jsonStr string) CmdBrResp {
	ret := CmdBrResp{}
	var req CmdRestoreReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		myPanic(fmt.Sprintln("json.unmarshal cmdrestorereq failed err:", err))
	}
	mgrStartNewBrRestoreTaskThruSQL(req.TaskID, &req)
	return ret
}

func mgrHandleCmdShowRestoreInfoReq(jsonStr string) CmdShowRestoreInfoResp {
	var ret CmdShowRestoreInfoResp
	var req CmdShowRestoreInfoReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		myPanic(fmt.Sprintln("json.unmarshal cmdshowrestoreinforeq failed err:", err))
	}
	ret = mgrStartNewBrShowRestoreInfoThruSQL(&req)
	return ret
}

func mgrHandleCmdGetAllTaskStatusReq(jsonStr string) CmdGetAllTaskStatusResp {
	var req CmdGetAllTaskStatusReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		myPanic(fmt.Sprintln("json.unmarshal cmdgetallgaskstatusreq failed err:", err))
	}
	glMgrStatusMapSync()
	return CmdGetAllTaskStatusResp{
		Stats: glMgrStatusMapGetAll(),
	}
}

func mgrStartNewBrBackUpTaskThruSQL(taskID uint64, req *CmdBackUpReq) {
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
		<-mgrStartNewBrTaskThruSQL(taskID, &req.DbConnParameter, strings.Join(args, " "))
	}()
}

func mgrStartNewBrShowBackUpInfoThruSQL(req *CmdShowBackUpInfoReq) CmdShowBackUpInfoResp {
	resp := CmdShowBackUpInfoResp{}
	brSQLCmd := "SHOW BACKUPS"
	dbConnParam := req.DbConnParameter
	logger.Info("task start processing:", fmt.Sprintf("brSQLCmd:%s", brSQLCmd))
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.Ip, dbConnParam.Port))
	if err != nil {
		logger.Error("db connection err:", err)
		resp.Error = err
		return resp
	}
	defer db.Close()
	t0 := time.Now()
	err = db.QueryRow(brSQLCmd).Scan(&resp.Destination, &resp.State, &resp.Progress, &resp.Queue_time, &resp.Execution_Time, &resp.Finish_Time, &resp.Connection)
	successFp := func() {
		logger.Info("showbackupinfo task finished, time cost", time.Since(t0))
	}
	if err != nil {
		logger.Errorf("query sql cmd err: %v", err)
		if err.Error() != "sql: no rows in result set" {
			logger.Debugf("(%s) != (sql: no rows in result set", err.Error())
			resp.Error = err
			return resp
		}
		logger.Debugf("(%s) == (sql: no rows in result set)", err.Error())
		logger.Infof("task has finished without checking db while no rows is result for sql cmd")
		resp.Progress = 100
		//stat, errStr, err := MicroSrvTiupGetTaskStatus(req.TaskID)
		//logger.Infof("stat: %v, errStr: %s, err: %v", stat, errStr, err)
		//if err != nil {
		//	logger.Error("get tiup status from db error", err)
		//	resp.Error = err
		//} else if stat != dbPb.TiupTaskStatus_Finished {
		//	logger.Errorf("task has not finished: %d, with err info: %s", stat, errStr)
		//	resp.Error = errors.New(fmt.Sprintf("task has not finished: %d, with err info: %s", stat, errStr))
		//} else {
		//	logger.Infof("task has finished: %d", stat)
		//	resp.Progress = 100
		//}
		return resp
	}
	logger.Info("sql cmd return successfully")
	successFp()
	return resp
}

func mgrStartNewBrRestoreTaskThruSQL(taskID uint64, req *CmdRestoreReq) {
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
		<-mgrStartNewBrTaskThruSQL(taskID, &req.DbConnParameter, strings.Join(args, " "))
	}()
}

func mgrStartNewBrShowRestoreInfoThruSQL(req *CmdShowRestoreInfoReq) CmdShowRestoreInfoResp {
	resp := CmdShowRestoreInfoResp{}
	brSQLCmd := "SHOW RESTORES"
	dbConnParam := req.DbConnParameter
	logger.Infof("task start processing: brSQLCmd:%s", brSQLCmd)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.Ip, dbConnParam.Port))
	if err != nil {
		logger.Error("db connection err", err)
		resp.Error = err
		return resp
	}
	defer db.Close()
	t0 := time.Now()
	err = db.QueryRow(brSQLCmd).Scan(&resp.Destination, &resp.State, &resp.Progress, &resp.Queue_time, &resp.Execution_Time, &resp.Finish_Time, &resp.Connection)
	successFp := func() {
		logger.Info("showretoreinfo task finished, time cost", time.Since(t0))
	}
	if err != nil {
		logger.Errorf("query sql cmd err: %v", err)
		if err.Error() != "sql: no rows in result set" {
			logger.Debugf("(%s) != (sql: no rows in result set", err.Error())
			resp.Error = err
			return resp
		}
		logger.Debugf("(%s) == (sql: no rows in result set)", err.Error())
		logger.Infof("task has finished without checking db while no rows is result for sql cmd")
		resp.Progress = 100
		//if stat, errStr, err := MicroSrvTiupGetTaskStatus(req.TaskID); err != nil {
		//	logger.Error("get tiup status error", err)
		//	resp.Error = err
		//} else if stat != dbPb.TiupTaskStatus_Finished {
		//	logger.Errorf("task has not finished: %d, with err info: %s", stat, errStr)
		//	resp.Error = errors.New(fmt.Sprintf("task has not finished: %d, with err info: %s", stat, errStr))
		//} else {
		//	logger.Info("sql cmd return successfully")
		//	resp.Progress = 100
		//}
		return resp
	}
	logger.Info("sql cmd return successfully")
	successFp()
	return resp
}

func mgrStartNewBrTaskThruSQL(taskID uint64, dbConnParam *DbConnParam, brSQLCmd string) (exitCh chan struct{}) {
	exitCh = make(chan struct{})
	logger.Infof("task start processing: brSQLCmd:%s", brSQLCmd)
	glMgrTaskStatusCh <- TaskStatusMember{
		TaskID:   taskID,
		Status:   TaskStatusProcessing,
		ErrorStr: "",
	}
	go func() {
		defer close(exitCh)

		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.Ip, dbConnParam.Port))
		if err != nil {
			logger.Error("db connection err", err)
			glMgrTaskStatusCh <- TaskStatusMember{
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
			glMgrTaskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintln(err),
			}
			return
		}
		successFp := func() {
			logger.Info("task finished, time cost", time.Since(t0))
			glMgrTaskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusFinished,
				ErrorStr: string(jsonMustMarshal(&resp)),
			}
		}
		logger.Info("sql cmd return successfully")
		successFp()
	}()
	return exitCh
}

// micro service part
type CmdChanMember struct {
	req    CmdReqOrResp
	respCh chan CmdReqOrResp
}

var glMicroCmdChan chan CmdChanMember
var glMicroTaskStatusMap map[uint64]TaskStatusMapValue
var glMicroTaskStatusMapMutex sync.Mutex

var glBrMgrPath string

type ClusterFacade struct {
	TaskID          uint64 // do not pass this value for br command
	DbConnParameter DbConnParam
	DbName          string
	TableName       string
	ClusterId       string // todo: need to know the usage
	ClusterName     string // todo: need to know the usage
	//PdAddress 			string
}

type BrStorage struct {
	StorageType StorageType
	Root        string // "/tmp/backup"
}

type ProgressRate struct {
	Rate    float32 // 0.99 means 99%
	Checked bool
	Error   error
}

func MicroInit(brMgrPath, mgrLogFilePath string) {
	framework.LogForkFile(common.LogFileSystem).Infof("microinit brmgrpath: %s, mgrlogfilepath: %s", brMgrPath, mgrLogFilePath)

	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	logger = framework.LogForkFile(configPath + common.LogFileLibBr)

	glBrMgrPath = brMgrPath
	glMicroTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	glMicroCmdChan = microStartBrMgr(mgrLogFilePath)
	go glMicroTaskStatusMapSyncer()
}

func glMicroTaskStatusMapSyncer() {
	for {
		time.Sleep(time.Second)
		resp := microTiupGetAllTaskStatus()
		var needDbUpdate []TaskStatusMember
		glMicroTaskStatusMapMutex.Lock()
		for _, v := range resp.Stats {
			oldv := glMicroTaskStatusMap[v.TaskID]
			if oldv.validFlag {
				if oldv.stat.Status == v.Status {
					assert(oldv.stat == v)
				} else {
					assert(oldv.stat.Status == TaskStatusProcessing)
					glMicroTaskStatusMap[v.TaskID] = TaskStatusMapValue{
						validFlag: true,
						stat:      v,
						readct:    0,
					}
					assert(v.Status == TaskStatusFinished || v.Status == TaskStatusError)
					needDbUpdate = append(needDbUpdate, v)
				}
			} else {
				glMicroTaskStatusMap[v.TaskID] = TaskStatusMapValue{
					validFlag: true,
					stat:      v,
					readct:    0,
				}
				needDbUpdate = append(needDbUpdate, v)
			}
		}
		glMicroTaskStatusMapMutex.Unlock()
		logInFunc := logger.WithField("glMicroTaskStatusMapSyncer", "DbClient.UpdateTiupTask")
		for _, v := range needDbUpdate {
			rsp, err := client.DBClient.UpdateTiupTask(context.Background(), &dbpb.UpdateTiupTaskRequest{
				Id:     v.TaskID,
				Status: dbpb.TiupTaskStatus(v.Status),
				ErrStr: v.ErrorStr,
			})
			if rsp == nil || err != nil || rsp.ErrCode != 0 {
				logInFunc.Error("rsp:", rsp, "err:", err, "v:", v)
			} else {
				logInFunc.Debug("update succes:", v)
			}
		}
	}
}

func microStartBrMgr(mgrLogFilePath string) chan CmdChanMember {
	brMgrPath := glBrMgrPath
	cmd := exec.Command(brMgrPath, mgrLogFilePath)
	cmd.SysProcAttr = genSysProcAttr()
	in, err := cmd.StdinPipe()
	if err != nil {
		myPanic("unexpected")
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		myPanic("unexpected")
	}
	//_, err = cmd.StderrPipe()
	//if err != nil {
	//	myPanic("unexpected")
	//}
	cch := make(chan CmdChanMember, 1024)
	if err := cmd.Start(); err != nil {
		myPanic(fmt.Sprint("start brmgr failed with err:", err))
	}
	go microCmdChanRoutine(cch, out, in)
	go func() {
		err = cmd.Wait()
		myPanic(fmt.Sprint("wait brmgr failed with err:", err))
	}()
	return cch
}

// Basically this method get request thru cch(which contain CmdReqOrResp and a chan of it as resp),
// and write command as string to mgr, and receive command as string from mgr
// cch: where micro service receives commands
// inWriter(of mgr): where micro service passes the command to the mgr
// outReader(of mgr): where micro service receives the output of mgr from
func microCmdChanRoutine(cch chan CmdChanMember, outReader io.Reader, inWriter io.Writer) {
	outBufReader := bufio.NewReader(outReader)
	for {
		cmdMember, ok := <-cch
		assert(ok)
		bs := jsonMustMarshal(cmdMember.req)
		bs = append(bs, '\n')
		ct, err := inWriter.Write(bs)
		assert(ct == len(bs) && err == nil)

		output, err := outBufReader.ReadString('\n')
		/*
			for len(output) == 0 {
				if err != nil {
					logger.Infof("Error while reading outReader from brmgr: %v\n", err)
				}
				output, err = outBufReader.ReadString('\n')
			}
		*/
		assert(len(output) > 1 && err == nil && output[len(output)-1] == '\n')
		var resp CmdReqOrResp
		err = json.Unmarshal([]byte(output[:len(output)-1]), &resp)
		assert(err == nil)
		select {
		case cmdMember.respCh <- resp:
		default:
		}
	}
}

// TODO: backup precheck command not found in BR, may need to check later
func BackUpPreCheck(cluster ClusterFacade, storage BrStorage, bizId uint64) error {
	return nil
}

func backUp(backUpReq CmdBackUpReq) CmdBrResp {
	assert(cap(glMicroCmdChan) > 0)
	cmdReq := CmdReqOrResp{
		TypeStr: CmdBackUpReqTypeStr,
		Content: string(jsonMustMarshal(&backUpReq)),
	}
	respCh := make(chan CmdReqOrResp, 1)
	glMicroCmdChan <- CmdChanMember{
		req:    cmdReq,
		respCh: respCh,
	}
	respCmd := <-respCh
	assert(respCmd.TypeStr == CmdBackUpRespTypeStr)
	var resp CmdBrResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

func BackUp(cluster ClusterFacade, storage BrStorage, bizId uint64) (taskID uint64, err error) {
	var req dbpb.CreateTiupTaskRequest
	req.Type = dbpb.TiupTaskType_Backup
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
		backUp(backupReq)
		return rsp.Id, nil
	}
}

func showBackUpInfo(showBackUpInfoReq CmdShowBackUpInfoReq) CmdShowBackUpInfoResp {
	assert(cap(glMicroCmdChan) > 0)
	cmdReq := CmdReqOrResp{
		TypeStr: CmdShowBackUpInfoReqTypeStr,
		Content: string(jsonMustMarshal(&showBackUpInfoReq)),
	}
	respCh := make(chan CmdReqOrResp, 1)
	glMicroCmdChan <- CmdChanMember{
		req:    cmdReq,
		respCh: respCh,
	}
	respCmd := <-respCh
	assert(respCmd.TypeStr == CmdShowBackUpInfoRespTypeStr)
	var resp CmdShowBackUpInfoResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

func ShowBackUpInfo(cluster ClusterFacade) CmdShowBackUpInfoResp {
	var showBackUpInfoReq CmdShowBackUpInfoReq
	showBackUpInfoReq.DbConnParameter = cluster.DbConnParameter
	showBackUpInfoReq.TaskID = cluster.TaskID
	showBackUpInfoResp := showBackUpInfo(showBackUpInfoReq)
	return showBackUpInfoResp
}

// TODO: restore precheck command not found in BR, may need to check later
func RestorePreCheck(cluster ClusterFacade, storage BrStorage, bizId uint64) error {
	return nil
}

func restore(restoreReq CmdRestoreReq) CmdBrResp {
	assert(cap(glMicroCmdChan) > 0)
	cmdReq := CmdReqOrResp{
		TypeStr: CmdRestoreReqTypeStr,
		Content: string(jsonMustMarshal(&restoreReq)),
	}
	respCh := make(chan CmdReqOrResp, 1)
	glMicroCmdChan <- CmdChanMember{
		req:    cmdReq,
		respCh: respCh,
	}
	respCmd := <-respCh
	assert(respCmd.TypeStr == CmdRestoreRespTypeStr)
	var resp CmdBrResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

func Restore(cluster ClusterFacade, storage BrStorage, bizId uint64) (taskID uint64, err error) {
	var req dbpb.CreateTiupTaskRequest
	req.Type = dbpb.TiupTaskType_Restore
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
		restore(restoreReq)
		return rsp.Id, nil
	}
}

func showRestoreInfo(showRestoreInfoReq CmdShowRestoreInfoReq) CmdShowRestoreInfoResp {
	assert(cap(glMicroCmdChan) > 0)
	cmdReq := CmdReqOrResp{
		TypeStr: CmdShowRestoreInfoReqTypeStr,
		Content: string(jsonMustMarshal(&showRestoreInfoReq)),
	}
	respCh := make(chan CmdReqOrResp, 1)
	glMicroCmdChan <- CmdChanMember{
		req:    cmdReq,
		respCh: respCh,
	}
	respCmd := <-respCh
	assert(respCmd.TypeStr == CmdShowRestoreInfoRespTypeStr)
	var resp CmdShowRestoreInfoResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

func ShowRestoreInfo(cluster ClusterFacade) CmdShowRestoreInfoResp {
	var showRestoreInfoReq CmdShowRestoreInfoReq
	showRestoreInfoReq.DbConnParameter = cluster.DbConnParameter
	showRestoreInfoResp := showRestoreInfo(showRestoreInfoReq)
	return showRestoreInfoResp
}

func MicroSrvTiupGetTaskStatus(taskID uint64) (stat dbpb.TiupTaskStatus, errStr string, err error) {
	var req dbpb.FindTiupTaskByIDRequest
	req.Id = taskID
	logger.Debugf("FindTiupTaskByID: %d", taskID)
	rsp, err := client.DBClient.FindTiupTaskByID(context.Background(), &req)
	logger.Debugf("FindTiupTaskByID: %d. rsp: %v, err: %v", taskID, rsp, err)
	if err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("err:%s, rsp.ErrCode:%d, rsp.ErrStr:%s", err, rsp.ErrCode, rsp.ErrStr)
		return stat, "", err
	} else {
		assert(rsp.TiupTask != nil && rsp.TiupTask.ID == taskID)
		stat = rsp.TiupTask.Status
		errStr = rsp.TiupTask.ErrorStr
		return stat, errStr, nil
	}
}

func microTiupGetAllTaskStatus() CmdGetAllTaskStatusResp {
	assert(cap(glMicroCmdChan) > 0)
	cmdReq := CmdReqOrResp{
		TypeStr: CmdGetAllTaskStatusReqTypeStr,
		Content: string(jsonMustMarshal(&CmdGetAllTaskStatusReq{})),
	}
	respCh := make(chan CmdReqOrResp, 1)
	glMicroCmdChan <- CmdChanMember{
		req:    cmdReq,
		respCh: respCh,
	}
	respCmd := <-respCh
	assert(respCmd.TypeStr == CmdGetAllTaskStatusRespTypeStr)
	var resp CmdGetAllTaskStatusResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}
