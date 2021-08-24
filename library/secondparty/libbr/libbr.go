package libbr

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap-inc/tiem/library/thirdparty/logger"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
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

type StorageType string

const (
	StorageTypeLocal 	StorageType = "local"
	StorageTypeS3 		StorageType = "s3"
)

type CmdBackUpPreCheckReq struct {
}

type CmdBackUpPreCheckResp struct {
}

type CmdBackUpReq struct {
	TaskID 				uint64
	DbName				string
	TableName 			string
	FilterDbTableName 	string // used in br command, pending for use in SQL command
	StorageAddress 		string
	DbConnParameter		DbConnParam // only for SQL command, not used in br command
	RateLimitM      	string
	Concurrency         string // only for SQL command, not used in br command
	CheckSum            string // only for SQL command, not used in br command
	LogFile				string // used in br command, pending for use in SQL command
	TimeoutS 			int // used in br command, pending for use in SQL command
	Flags 				[]string // used in br command, pending for use in SQL command
}

type CmdBackUpResp struct {
}

type CmdShowBackUpInfoReq struct {
	DbConnParameter		DbConnParam
}

type CmdShowBackUpInfoResp struct {
	Destination		string
	State			string
	Progress		float32
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
	DbName				string
	TableName 			string
	FilterDbTableName 	string // used in br command, pending for use in SQL command
	StorageAddress 		string
	DbConnParameter		DbConnParam // only for SQL command, not used in br command
	RateLimitM      	string
	Concurrency         string // only for SQL command, not used in br command
	CheckSum            string // only for SQL command, not used in br command
	LogFile				string // used in br command, pending for use in SQL command
	TimeoutS 			int // used in br command, pending for use in SQL command
	Flags 				[]string // used in br command, pending for use in SQL command
}

type CmdRestoreResp struct {
}

type CmdShowRestoreInfoReq struct {
	DbConnParameter		DbConnParam
}

type CmdShowRestoreInfoResp struct {
	Destination		string
	State			string
	Progress		float32
	Queue_time		string
	Execution_Time	string
	Finish_Time		string
	Connection		string
}

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

//var glMgrTaskStatusCh chan TaskStatusMember
//var glMgrTaskStatusMap map[uint64]TaskStatusMapValue

var log *logger.LogRecord

var glMgrTaskStatusCh chan TaskStatusMember

func BrMgrInit() {
	glMgrTaskStatusCh = make(chan TaskStatusMember, 1024)
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
	fmt.Printf("panic: %s, with stack trace: %v\n", s, string(debug.Stack()))
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

var glBrMgrPath string

type ClusterFacade struct {
	DbConnParameter		DbConnParam
	DbName				string
	TableName 			string
	ClusterId 			string // todo: need to know the usage
	ClusterName 		string // todo: need to know the usage
	//PdAddress 			string
}

type BrStorage struct {
	StorageType StorageType
	Root string // "/tmp/backup"
}

type ProgressRate struct {
	Rate 	float32 // 0.99 means 99%
	Checked bool
	Error 	error
}

func MicroInit(brMgrPath, mgrLogFilePath string) {
	glBrMgrPath = brMgrPath
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
		var cmdMember CmdChanMember
		var ok bool
		select {
		case cmdMember, ok = <-cch:
			assert(ok)
		}
		bs := jsonMustMarshal(cmdMember.req)
		bs = append(bs, '\n')
		ct, err := inWriter.Write(bs)
		assert(ct == len(bs) && err == nil)
		output, err := outBufReader.ReadString('\n')
		//fmt.Println(string(output), len(output), err)
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
func BackUpPreCheck(cluster ClusterFacade, storage BrStorage, bizId uint) error {
	return nil
}

func backUp(backUpReq CmdBackUpReq) CmdBackUpResp {
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
	var resp CmdBackUpResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

// todo: back up data in s3
func BackUp(cluster ClusterFacade, storage BrStorage, bizId uint) error {
	var backupReq CmdBackUpReq
	backupReq.DbConnParameter = cluster.DbConnParameter
	backupReq.DbName = cluster.DbName
	backupReq.TableName = cluster.TableName
	backupReq.StorageAddress = fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root)
	backUp(backupReq)
	return nil
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

func ShowBackUpInfo(cluster ClusterFacade, bizId uint) ProgressRate {
	var progressRate ProgressRate
	var showBackUpInfoReq CmdShowBackUpInfoReq
	showBackUpInfoReq.DbConnParameter = cluster.DbConnParameter
	showBackUpInfoResp := showBackUpInfo(showBackUpInfoReq)
	progressRate.Rate = showBackUpInfoResp.Progress/100
	return progressRate
}

// TODO: restore precheck command not found in BR, may need to check later
// restore precheck list:
// 1. if the data was backed up locally, make sure the sst files have been copied to every TiKV nodes
func RestorePreCheck(cluster ClusterFacade, storage BrStorage, bizId uint) error {
	return nil
}

func restore(restoreReq CmdRestoreReq) CmdRestoreResp {
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
	var resp CmdRestoreResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

func Restore(cluster ClusterFacade, storage BrStorage, bizId uint) error {
	var restoreReq CmdRestoreReq
	restoreReq.DbConnParameter = cluster.DbConnParameter
	restoreReq.DbName = cluster.DbName
	restoreReq.TableName = cluster.TableName
	restoreReq.StorageAddress = fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root)
	restore(restoreReq)
	return nil
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

func ShowRestoreInfo(cluster ClusterFacade, bizId uint) ProgressRate {
	var progressRate ProgressRate
	var showRestoreInfoReq CmdShowRestoreInfoReq
	showRestoreInfoReq.DbConnParameter = cluster.DbConnParameter
	showRestoreInfoResp := showRestoreInfo(showRestoreInfoReq)
	progressRate.Rate = showRestoreInfoResp.Progress/100
	return progressRate
}
