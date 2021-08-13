package libbr

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap/tiem/library/thirdparty/logger"
	dbPb "github.com/pingcap/tiem/micro-metadb/proto"
	"io"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"
)

type CmdTypeStr string

const (
	CmdBackUpPreCheckReqTypeStr            	CmdTypeStr = "CmdBackUpPreCheckReq"
	CmdBackUpPreCheckRespTypeStr           	CmdTypeStr = "CmdBackUpPreCheckResp"
	CmdBackUpReqTypeStr             		CmdTypeStr = "CmdBackUpReq"
	CmdBackUpRespTypeStr            		CmdTypeStr = "CmdBackUpResp"
	CmdShowBackUpInfoReqTypeStr             CmdTypeStr = "CmdShowBackUpInfoReq"
	CmdShowBackUpInfoRespTypeStr            CmdTypeStr = "CmdShowBackUpInfoResp"
	CmdRestorePreCheckReqTypeStr           	CmdTypeStr = "CmdRestorePreCheckReq"
	CmdRestorePreCheckRespTypeStr          	CmdTypeStr = "CmdRestorePreCheckResp"
	CmdRestoreReqTypeStr  					CmdTypeStr = "CmdRestoreReq"
	CmdRestoreRespTypeStr 					CmdTypeStr = "CmdRestoreResp"
	CmdShowRestoreInfoReqTypeStr  			CmdTypeStr = "CmdShowRestoreInfoReq"
	CmdShowRestoreInfoRespTypeStr 			CmdTypeStr = "CmdShowRestoreInfoResp"
)

type BrScopeStr string

const (
	FullBrScope		BrScopeStr = "full"
	DbBrScope		BrScopeStr = "db"
	TableBrScope	BrScopeStr = "table"
)

type CmdReqOrResp struct {
	TypeStr CmdTypeStr
	Content string
}

type DbConnParam struct {
	Username	string
	Password	string
	Ip			string
	Port		string
}

type CmdBackUpPreCheckReq struct {
}

type CmdBackUpPreCheckResp struct {
}

type CmdBackUpReq struct {
	TaskID 				uint64
	PdAddress 			string // only used in br command, not used in SQL command
	DbName				string
	TableName 			string
	FilterDbTableName 	string // used in br command, pending for use in SQL command
	StorageAddress 		string
	BrScope      		BrScopeStr // only for br command, not used in SQL command
	DbConnParameter		DbConnParam // only for SQL command, not used in SQL command
	RateLimitM      	string
	Concurrency         string // only for SQL command, not used in SQL command
	CheckSum            string // only for SQL command, not used in SQL command
	LogFile				string // used in br command, pending for use in SQL command
	TimeoutS 			int // used in br command, pending for use in SQL command
	BrPath 				string // only for br command, not used in SQL command
	Flags 				[]string // used in br command, pending for use in SQL command
}

type CmdBackUpResp struct {
}

type CmdShowBackUpInfoReq struct {
	DbConnParameter		DbConnParam // only for SQL command, not used in SQL command
}

type CmdShowBackUpInfoResp struct {
	Destination		string
	State			string
	Progress		string
	Queue_time		string
	Execution_Time	string
	Finish_Time		*string
	Connection		string
}

type CmdRestorePreCheckReq struct {
}

type CmdRestorePreCheckResp struct {
}

type CmdRestoreReq struct {
	TaskID 				uint64
	PdAddress 			string // only used in br command, not used in SQL command
	DbName				string
	TableName 			string
	FilterDbTableName 	string // used in br command, pending for use in SQL command
	StorageAddress 		string
	BrScope      		BrScopeStr // only for br command, not used in SQL command
	DbConnParameter		DbConnParam // only for SQL command, not used in SQL command
	RateLimitM      	string
	Concurrency         string // only for SQL command, not used in SQL command
	CheckSum            string // only for SQL command, not used in SQL command
	LogFile				string // used in br command, pending for use in SQL command
	TimeoutS 			int // used in br command, pending for use in SQL command
	BrPath 				string // only for br command, not used in SQL command
	Flags 				[]string // used in br command, pending for use in SQL command
}

type CmdRestoreResp struct {
}

type CmdShowRestoreInfoReq struct {
	DbConnParameter		DbConnParam // only for SQL command, not used in SQL command
}

type CmdShowRestoreInfoResp struct {
	Destination		string
	State			string
	Progress		string
	Queue_time		string
	Execution_Time	string
	Finish_Time		string
	Connection		string
}

type StorageType string

type TaskStatus int

const (
	TaskStatusInit       TaskStatus = TaskStatus(dbPb.TiupTaskStatus_Init)
	TaskStatusProcessing TaskStatus = TaskStatus(dbPb.TiupTaskStatus_Processing)
	TaskStatusFinished   TaskStatus = TaskStatus(dbPb.TiupTaskStatus_Finished)
	TaskStatusError      TaskStatus = TaskStatus(dbPb.TiupTaskStatus_Error)
)

type TaskStatusMember struct {
	TaskID   uint64
	Status   TaskStatus
	ErrorStr string
}

type TaskStatusMapValue struct {
	validFlag bool
	stat      TaskStatusMember
	readct    uint64
}

//var glMgrTaskStatusCh chan TaskStatusMember
//var glMgrTaskStatusMap map[uint64]TaskStatusMapValue

var log *logger.LogRecord

const (
	StorageTypeLocal 	StorageType = "local"
	StorageTypeS3 		StorageType = "s3"
)

type ClusterFacade struct {
	ClusterId 	string
	ClusterName string
	PdAddress 	string
	BrScope		BrScopeStr
}

type BRPort interface {
	BackUpPreCheck(cluster ClusterFacade, storage BrStorage, bizId uint) error

	BackUp(cluster ClusterFacade, storage BrStorage, bizId uint) error

	ShowBackUpInfo(bizId uint) ProgressRate

	RestorePreCheck(cluster ClusterFacade, storage BrStorage, bizId uint) error

	Restore(cluster ClusterFacade, storage BrStorage, bizId uint) error

	ShowRestoreInfo(bizId uint) ProgressRate
}

type ProgressRate struct {
	Rate 	float32 // 0.99 means 99%
	Checked bool
	Error 	error
}

type BrStorage struct {
	StorageType StorageType
	Root string // "/tmp/backup"
}

var glMgrTaskStatusCh chan TaskStatusMember
var glMgrTaskStatusMap map[uint64]TaskStatusMapValue

func BrMgrInit() {
	glMgrTaskStatusCh = make(chan TaskStatusMember, 1024)
	glMgrTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	// TODO: comprehend this part
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	if len(configPath) == 0 {
		configPath = "./brmgr.log"
	}
}

func BrMgrRoutine() {
	inReader := bufio.NewReader(os.Stdin)
	outWriter := os.Stdout
	errw := os.Stderr

	for {
		input, err := inReader.ReadString('\n')
		if err != nil {
			myPanic(err)
		}
		errw.Write([]byte(input))
		if input[len(input)-1] == '\n' {
			cmdStr := input[:len(input)-1]
			var cmd CmdReqOrResp
			err := json.Unmarshal([]byte(cmdStr), &cmd)
			if err != nil {
				myPanic(fmt.Sprintln("cmdStr unmarshal failed err:", err, "cmdStr:", cmdStr))
			}
			//log.Info("rcv req", cmd)
			fmt.Println("rcv req: ", cmd)
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
			default:
				myPanic(fmt.Sprintln("unknown cmdStr.TypeStr:", cmd.TypeStr))
			}
			//log.Info("snd rsp", cmdResp)
			fmt.Println("snd rsp: ", cmdResp)
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
		//log.Fatal("unexpected panic with stack trace:", string(debug.Stack()))
		fmt.Println("unexpected panic with stack trace:", string(debug.Stack()))
		panic("unexpected")
	}
}

func myPanic(v interface{}) {
	s := fmt.Sprint(v)
	//log.Fatalf("panic: %s, with stack trace:", s, string(debug.Stack()))
	fmt.Println("panic: %s, with stack trace:", s, string(debug.Stack()))
	panic("unexpected")
}

func jsonMustMarshal(v interface{}) []byte {
	bs, err := json.Marshal(v)
	assert(err == nil)
	return bs
}

func mgrHandleCmdBackUpReq(jsonStr string) CmdBackUpResp {
	// TODO: ret is empty for now, may need to fill it
	ret := CmdBackUpResp{}
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
	// TODO: ret is empty for now, may need to fill it
	var ret CmdShowBackUpInfoResp
	var req CmdShowBackUpInfoReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		myPanic(fmt.Sprintln("json.unmarshal cmdshowbackupinforeq failed err:", err))
	}
	ret = mgrStartNewBrShowBackUpInfoThruSQL(&req)
	return ret
}

func mgrHandleCmdRestoreReq(jsonStr string) CmdRestoreResp {
	// TODO: ret is empty for now, may need to fill it
	ret := CmdRestoreResp{}
	var req CmdRestoreReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		myPanic(fmt.Sprintln("json.unmarshal cmdrestorereq failed err:", err))
	}
	mgrStartNewBrRestoreTaskThruSQL(req.TaskID, &req)
	return ret
}

func mgrHandleCmdShowRestoreInfoReq(jsonStr string) CmdShowRestoreInfoResp {
	// TODO: ret is empty for now, may need to fill it
	var ret CmdShowRestoreInfoResp
	var req CmdShowRestoreInfoReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		myPanic(fmt.Sprintln("json.unmarshal cmdshowrestoreinforeq failed err:", err))
	}
	ret = mgrStartNewBrShowRestoreInfoThruSQL(&req)
	return ret
}

//func mgrStartNewBrBackUpTask(taskID uint64, req *CmdBackUpReq) {
//	go func() {
//		var args []string
//		args = append(args, "backup", string(req.BrScope))
//		args = append(args, "--pd", req.PdAddress)
//		switch req.BrScope {
//		case FullBrScope:
//			if len(req.FilterDbTableName) != 0 {
//				args = append(args, "--filter", req.FilterDbTableName)
//			}
//		case DbBrScope:
//			args = append(args, "--db", req.DbName)
//		case TableBrScope:
//			args = append(args, "--db", req.DbName)
//			args = append(args, "--table", req.TableName)
//		}
//		args = append(args, "--storage", req.StorageAddress)
//		args = append(args, "--ratelimit", req.RateLimitM)
//		args = append(args, "--log-file", req.LogFile)
//		args = append(args, req.Flags...)
//		<-mgrStartNewBrTask(taskID, req.BrPath, args, req.TimeoutS)
//	}()
//}

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
	fmt.Println("task start processing:", fmt.Sprintf("brSQLCmd:%s", brSQLCmd))
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.Ip, dbConnParam.Port))
	if err != nil {
		fmt.Println("db connection err", err)
		//log.Error("db connection err", err)
		return resp
	}
	defer db.Close()
	t0 := time.Now()
	err = db.QueryRow(brSQLCmd).Scan(&resp.Destination, &resp.State, &resp.Progress, &resp.Queue_time, &resp.Execution_Time, &resp.Finish_Time, &resp.Connection)
	if err != nil {
		fmt.Println("query sql cmd err", err)
		//log.Error("query sql cmd err", err)
		return resp
	}
	successFp := func() {
		fmt.Println("task finished, time cost", time.Now().Sub(t0))
		//log.Info("task finished, time cost", time.Now().Sub(t0))
	}
	fmt.Println("sql cmd return successfully", resp.Progress)
	//log.Info("sql cmd return successfully")
	successFp()
	return resp
}

//func mgrStartNewBrRestoreTask(taskID uint64, req *CmdRestoreReq) {
//	go func() {
//		var args []string
//		args = append(args, "restore", string(req.BrScope))
//		args = append(args, "--pd", req.PdAddress)
//		switch req.BrScope {
//		case FullBrScope:
//			if len(req.FilterDbTableName) != 0 {
//				args = append(args, "--filter", req.FilterDbTableName)
//			}
//		case DbBrScope:
//			args = append(args, "--db", req.DbName)
//		case TableBrScope:
//			args = append(args, "--db", req.DbName)
//			args = append(args, "--table", req.TableName)
//		}
//		args = append(args, "--storage", req.StorageAddress)
//		args = append(args, "--ratelimit", req.RateLimitM)
//		args = append(args, "--log-file", req.LogFile)
//		args = append(args, req.Flags...)
//		<-mgrStartNewBrTask(taskID, req.BrPath, args, req.TimeoutS)
//	}()
//}

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
	fmt.Println("task start processing:", fmt.Sprintf("brSQLCmd:%s", brSQLCmd))
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.Ip, dbConnParam.Port))
	if err != nil {
		fmt.Println("db connection err", err)
		//log.Error("db connection err", err)
		return resp
	}
	defer db.Close()
	t0 := time.Now()
	err = db.QueryRow(brSQLCmd).Scan(&resp.Destination, &resp.State, &resp.Progress, &resp.Queue_time, &resp.Execution_Time, &resp.Finish_Time, &resp.Connection)
	if err != nil {
		fmt.Println("query sql cmd err", err)
		//log.Error("query sql cmd err", err)
		return resp
	}
	successFp := func() {
		fmt.Println("task finished, time cost", time.Now().Sub(t0))
		//log.Info("task finished, time cost", time.Now().Sub(t0))
	}
	fmt.Println("sql cmd return successfully", resp.Progress)
	//log.Info("sql cmd return successfully")
	successFp()
	return resp
}

//func mgrStartNewBrTask(taskID uint64, brPath string, brArgs []string, TimeoutS int) (exitCh chan struct{}) {
//	exitCh = make(chan struct{})
//	log := log.Record("task", taskID)
//	log.Info("task start processing:", fmt.Sprintf("brPath:%s brArgs:%v timeouts:%d", brPath, brArgs, TimeoutS))
//	glMgrTaskStatusCh <- TaskStatusMember{
//		TaskID:   taskID,
//		Status:   TaskStatusProcessing,
//		ErrorStr: "",
//	}
//	go func() {
//		defer close(exitCh)
//		var cmd *exec.Cmd
//		var cancelFp context.CancelFunc
//		if TimeoutS != 0 {
//			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(TimeoutS)*time.Second)
//			cancelFp = cancel
//			cmd = exec.CommandContext(ctx, brPath, brArgs...)
//		} else {
//			cmd = exec.Command(brPath, brArgs...)
//			cancelFp = func() {}
//		}
//		defer cancelFp()
//		cmd.SysProcAttr = genSysProcAttr()
//		t0 := time.Now()
//		if err := cmd.Start(); err != nil {
//			log.Error("cmd start err", err)
//			glMgrTaskStatusCh <- TaskStatusMember{
//				TaskID:   taskID,
//				Status:   TaskStatusError,
//				ErrorStr: fmt.Sprintln(err),
//			}
//			return
//		}
//		log.Info("cmd started")
//		successFp := func() {
//			log.Info("task finished, time cost", time.Now().Sub(t0))
//			glMgrTaskStatusCh <- TaskStatusMember{
//				TaskID:   taskID,
//				Status:   TaskStatusFinished,
//				ErrorStr: "",
//			}
//		}
//		log.Info("cmd wait")
//		err := cmd.Wait()
//		if err != nil {
//			log.Error("cmd wait return with err", err)
//			if exiterr, ok := err.(*exec.ExitError); ok {
//				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
//					if status.ExitStatus() == 0 {
//						successFp()
//						return
//					}
//				}
//			}
//			log.Error("task err:", err, "time cost", time.Now().Sub(t0))
//			glMgrTaskStatusCh <- TaskStatusMember{
//				TaskID:   taskID,
//				Status:   TaskStatusError,
//				ErrorStr: fmt.Sprintln(err),
//			}
//			return
//		} else {
//			log.Info("cmd wait return successfully")
//			successFp()
//			return
//		}
//	}()
//	return exitCh
//}

func mgrStartNewBrTaskThruSQL(taskID uint64, dbConnParam *DbConnParam, brSQLCmd string) (exitCh chan struct{}) {
	exitCh = make(chan struct{})
	//log := log.Record("task", taskID)
	fmt.Println("task start processing:", fmt.Sprintf("brSQLCmd:%s", brSQLCmd))
	//log.Info("task start processing:", fmt.Sprintf("brSQLCmd:%s", brSQLCmd))
	glMgrTaskStatusCh <- TaskStatusMember{
		TaskID:   taskID,
		Status:   TaskStatusProcessing,
		ErrorStr: "",
	}
	go func() {
		defer close(exitCh)

		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", dbConnParam.Username, dbConnParam.Password, dbConnParam.Ip, dbConnParam.Port))
		if err != nil {
			fmt.Println("db connection err", err)
			//log.Error("db connection err", err)
			glMgrTaskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintln(err),
			}
			return
		}
		defer db.Close()
		t0 := time.Now()
		_, err = db.Query(brSQLCmd)
		if err != nil {
			fmt.Println("query sql cmd err", err)
			//log.Error("query sql cmd err", err)
			glMgrTaskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintln(err),
			}
			return
		}
		successFp := func() {
			fmt.Println("task finished, time cost", time.Now().Sub(t0))
			//log.Info("task finished, time cost", time.Now().Sub(t0))
			glMgrTaskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusFinished,
				ErrorStr: "",
			}
		}
		fmt.Println("sql cmd return successfully")
		//log.Info("sql cmd return successfully")
		successFp()
		return
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

var glBrMgrPath string
var glBrBinPath string

func MicroInit(brMgrPath, brBinPath, mgrLogFilePath string) {
	glBrMgrPath = brMgrPath
	glBrBinPath = brBinPath
	glMicroTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	glMicroCmdChan = microStartBrMgr(mgrLogFilePath)
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
	_, err = cmd.StderrPipe()
	if err != nil {
		myPanic("unexpected")
	}
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

func microCmdChanRoutine(cch chan CmdChanMember, outReader io.Reader, inWriter io.Writer) {
	//outBufReader := bufio.NewReader(outReader)
	//for {
	//	var cmdMember CmdChanMember
	//}
}

// TODO: backup precheck command not found in BR, may need to check later
func BackUpPreCheck(cluster ClusterFacade, storage BrStorage, bizId uint) error {
	return nil
}

// 1. back up data in local
// 2. back up data in s3
func BackUp(cluster ClusterFacade, storage BrStorage, bizId uint) error {
	// back up data in local
	//var req CmdBackUpReq = CmdBackUpReq{
	//	BrScope: cluster.BrScope,
	//	PdAddress: cluster.PdAddress,
	//}
	//// TODO: define taskID
	//mgrStartNewBrBackUpTask(0, &req)
	return nil
}

//func (brMgr *BrMgr) mgrHandleCmdBackUpReq(jsonStr string) CmdBackUpResp {
//	// TODO: ret is empty for now, fill it
//	ret := CmdBackUpResp{}
//	var req CmdBackUpReq
//	err := json.Unmarshal([]byte(jsonStr), &req)
//	if err != nil {
//		myPanic(fmt.Sprintln("json.unmarshal cmdbackupreq failed err:", err))
//	}
//	brMgr.mgrStartNewBrBackUpTask(req.TaskID, &req)
//	return ret
//}

// TODO: show back up info command not found in BR, may need to check later
func ShowBackUpInfo(bizId uint) ProgressRate {
	return ProgressRate{}
}

// TODO: restore precheck command not found in BR, may need to check later
// restore precheck list:
	// 1. if the data was backed up locally, make sure the sst files have been copied to every TiKV nodes
func RestorePreCheck(cluster ClusterFacade, storage BrStorage, bizId uint) error {
	return nil
}

func Restore(cluster ClusterFacade, storage BrStorage, bizId uint) error {
	return nil
}

// TODO: show restore info command not found in BR, may need to check later
func ShowRestoreInfo(bizId uint) ProgressRate {
	return ProgressRate{}
}