package libbr

import (
	"bufio"
	"fmt"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var tmpFile *os.File

var backUpReq CmdBackUpReq
var showBackUpInfoReq CmdShowBackUpInfoReq
var restoreReq CmdRestoreReq
var showRestoreInfoReq CmdShowRestoreInfoReq

var cmdBackupReqWrap CmdReqOrResp
var cmdShowBackUpInfoReqWrap CmdReqOrResp
var cmdRestoreReqWrap CmdReqOrResp
var cmdShowRestoreInfoReqWrap CmdReqOrResp
var cmdGetAllTaskStatusReqWrap CmdReqOrResp
var cmdDefaultReqWrap CmdReqOrResp

func init() {
	logger = framework.LogForkFile(common.LogFileBrMgr)
	dbConnParam := DbConnParam{
		Username: "root",
		Ip:       "127.0.0.1",
		Port:     "4000",
	}
	storage := BrStorage{
		StorageType: StorageTypeLocal,
		Root:        "/tmp/backup",
	}
	//clusterFacade := ClusterFacade{
	//	DbConnParameter: dbConnParam,
	//}

	backUpReq = CmdBackUpReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
		StorageAddress:  fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root),
	}
	showBackUpInfoReq = CmdShowBackUpInfoReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
	}
	restoreReq = CmdRestoreReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
		StorageAddress:  fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root),
	}
	showRestoreInfoReq = CmdShowRestoreInfoReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
	}

	cmdBackupReqWrap = CmdReqOrResp{
		TypeStr: CmdBackUpReqTypeStr,
		Content: string(jsonMustMarshal(&backUpReq)),
	}
	cmdShowBackUpInfoReqWrap = CmdReqOrResp{
		TypeStr: CmdShowBackUpInfoReqTypeStr,
		Content: string(jsonMustMarshal(&showBackUpInfoReq)),
	}
	cmdRestoreReqWrap = CmdReqOrResp{
		TypeStr: CmdRestoreReqTypeStr,
		Content: string(jsonMustMarshal(&restoreReq)),
	}
	cmdShowRestoreInfoReqWrap = CmdReqOrResp{
		TypeStr: CmdShowRestoreInfoReqTypeStr,
		Content: string(jsonMustMarshal(&showRestoreInfoReq)),
	}

	cmdGetAllTaskStatusReqWrap = CmdReqOrResp{
		TypeStr: CmdGetAllTaskStatusReqTypeStr,
		Content: string(jsonMustMarshal(&CmdGetAllTaskStatusReq{})),
	}
	cmdDefaultTypeStr := CmdTypeStr("CmdDefaultReq")
	cmdDefaultReqWrap = CmdReqOrResp{
		TypeStr: cmdDefaultTypeStr,
		Content: "",
	}
}

func Test_assert(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	assert(false)
}

func TestBrCmd_BrMgrInit(t *testing.T) {
	BrMgrInit()
	mgrTaskStatusChCap := cap(glMgrTaskStatusCh)
	mgrTaskStatusMapLen := len(glMgrTaskStatusMap)
	if mgrTaskStatusChCap != 1024 {
		t.Errorf("glMgrTaskStatusCh cap was incorrect, got: %d, want: %d.", mgrTaskStatusChCap, 1024)
	}
	if mgrTaskStatusMapLen != 0 {
		t.Errorf("mgrTaskStatusMap len was incorrect, got: %d, want: %d.", mgrTaskStatusMapLen, 0)
	}
}

func TestBrCmd_BrMgrRoutine(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	BrMgrRoutine()
}

func Test_recvReqAndSndResp_v1(t *testing.T) {
	content := []byte("wrong cmd")
	tmpFile, err := ioutil.TempFile("./", "tmpFile")
	if err != nil {
		t.Errorf("fail create file tmpFile1")
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		t.Error(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Error(err)
	}
	os.Stdin = tmpFile
	inReader := bufio.NewReader(os.Stdin)
	defer tmpFile.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()

	recvReqAndSndResp(inReader, os.Stdout)
}

func Test_recvReqAndSndResp_v2(t *testing.T) {
	content := []byte("wrong cmd\n")
	tmpFile, err := ioutil.TempFile("./", "tmpFile")
	if err != nil {
		t.Errorf("fail create file tmpFile1")
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		t.Error(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Error(err)
	}
	os.Stdin = tmpFile
	inReader := bufio.NewReader(os.Stdin)
	defer tmpFile.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()

	recvReqAndSndResp(inReader, os.Stdout)
}

func Test_recvReqAndSndResp_getAllTaskStatus(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdGetAllTaskStatusReqWrap))))
	tmpFile, err := ioutil.TempFile("./", "tmpFile")
	if err != nil {
		t.Errorf("fail create file tmpFile1")
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		t.Error(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Error(err)
	}
	os.Stdin = tmpFile
	inReader := bufio.NewReader(os.Stdin)
	defer tmpFile.Close()

	recvReqAndSndResp(inReader, os.Stdout)
}

func Test_recvReqAndSndResp_backup(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdBackupReqWrap))))
	tmpFile, err := ioutil.TempFile("./", "tmpFile")
	if err != nil {
		t.Errorf("fail create file tmpFile")
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		t.Error(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Error(err)
	}
	os.Stdin = tmpFile
	inReader := bufio.NewReader(os.Stdin)
	defer tmpFile.Close()

	recvReqAndSndResp(inReader, os.Stdout)
}

func Test_recvReqAndSndResp_showBackupInfo(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdShowBackUpInfoReqWrap))))
	tmpFile, err := ioutil.TempFile("./", "tmpFile")
	if err != nil {
		t.Errorf("fail create file tmpFile")
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		t.Error(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Error(err)
	}
	os.Stdin = tmpFile
	inReader := bufio.NewReader(os.Stdin)
	defer tmpFile.Close()

	recvReqAndSndResp(inReader, os.Stdout)
}

func Test_recvReqAndSndResp_restore(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdRestoreReqWrap))))
	tmpFile, err := ioutil.TempFile("./", "tmpFile")
	if err != nil {
		t.Errorf("fail create file tmpFile")
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		t.Error(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Error(err)
	}
	os.Stdin = tmpFile
	inReader := bufio.NewReader(os.Stdin)
	defer tmpFile.Close()

	recvReqAndSndResp(inReader, os.Stdout)
}

func Test_recvReqAndSndResp_showRestoreInfo(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdShowRestoreInfoReqWrap))))
	tmpFile, err := ioutil.TempFile("./", "tmpFile")
	if err != nil {
		t.Errorf("fail create file tmpFile")
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		t.Error(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Error(err)
	}
	os.Stdin = tmpFile
	inReader := bufio.NewReader(os.Stdin)
	defer tmpFile.Close()

	recvReqAndSndResp(inReader, os.Stdout)
}

func Test_recvReqAndSndResp_default(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdDefaultReqWrap))))
	tmpFile, err := ioutil.TempFile("./", "tmpFile")
	if err != nil {
		t.Errorf("fail create file tmpFile")
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(content); err != nil {
		t.Error(err)
	}
	if _, err := tmpFile.Seek(0, 0); err != nil {
		t.Error(err)
	}
	os.Stdin = tmpFile
	inReader := bufio.NewReader(os.Stdin)
	defer tmpFile.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()

	recvReqAndSndResp(inReader, os.Stdout)
}

func Test_mgrStartNewBrBackUpTaskThruSQL(t *testing.T) {
	backUpReq.TableName = ""
	backUpReq.DbName = "test_db"
	backUpReq.RateLimitM = "100"
	backUpReq.Concurrency = "8"
	backUpReq.CheckSum = "1"
	mgrStartNewBrBackUpTaskThruSQL(0, &backUpReq)
	time.Sleep(time.Millisecond * 50)

	backUpReq.TableName = "test_tbl"
	mgrStartNewBrBackUpTaskThruSQL(0, &backUpReq)
}

func Test_mgrStartNewBrRestoreTaskThruSQL(t *testing.T) {
	restoreReq.TableName = ""
	restoreReq.DbName = "test_db"
	restoreReq.RateLimitM = "100"
	restoreReq.Concurrency = "8"
	restoreReq.CheckSum = "1"
	mgrStartNewBrRestoreTaskThruSQL(0, &restoreReq)

	restoreReq.TableName = "test_tbl"
	mgrStartNewBrRestoreTaskThruSQL(0, &restoreReq)
}