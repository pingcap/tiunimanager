package libtiup

import (
	"bufio"
	"fmt"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"io/ioutil"
	"os"
	"testing"
)

var tmpFile *os.File

var deployReq CmdDeployReq
var startReq CmdStartReq
var listReq CmdListReq
var destroyReq CmdDestroyReq
var dumplingReq CmdDumplingReq
var lightningReq CmdLightningReq
var clusterDisplayReq CmdClusterDisplayReq

var cmdDeployReqWrap CmdReqOrResp
var cmdStartReqWrap CmdReqOrResp
var cmdListReqWrap CmdReqOrResp
var cmdDestroyReqWrap CmdReqOrResp
var cmdDumplingReqWrap CmdReqOrResp
var cmdLightningReqWrap CmdReqOrResp
var cmdClusterDisplayReqWrap CmdReqOrResp

var cmdGetAllTaskStatusReqWrap CmdReqOrResp
var cmdDefaultReqWrap CmdReqOrResp

func init() {
	logger = framework.LogForkFile(common.LogFileTiupMgr)

	deployReq = CmdDeployReq{
		InstanceName: "test-tidb",
		Version: "v1",
		TimeoutS: 0,
		Flags: []string{},
		TaskID: 1,
	}
	startReq = CmdStartReq{
		TaskID: 1,
		InstanceName: "test-tidb",
		TimeoutS: 0,
		Flags: []string{},
	}
	listReq = CmdListReq{
		TaskID: 1,
		TimeoutS: 0,
		Flags: []string{},
	}
	destroyReq = CmdDestroyReq{
		TaskID: 1,
		InstanceName: "test-tidb",
		TimeoutS: 0,
		Flags: []string{},
	}
	dumplingReq = CmdDumplingReq{
		TaskID: 1,
		TimeoutS: 0,
		Flags: []string{},
	}
	lightningReq = CmdLightningReq{
		TaskID: 1,
		TimeoutS: 0,
		Flags: []string{},
	}
	clusterDisplayReq = CmdClusterDisplayReq{
		ClusterName: "test-tidb",
		TimeoutS: 0,
		Flags: []string{},
	}

	cmdDeployReqWrap = CmdReqOrResp{
		TypeStr: CmdDeployReqTypeStr,
		Content: string(jsonMustMarshal(&deployReq)),
	}
	cmdStartReqWrap = CmdReqOrResp{
		TypeStr: CmdStartReqTypeStr,
		Content: string(jsonMustMarshal(&startReq)),
	}
	cmdListReqWrap = CmdReqOrResp{
		TypeStr: CmdListReqTypeStr,
		Content: string(jsonMustMarshal(&listReq)),
	}
	cmdDestroyReqWrap = CmdReqOrResp{
		TypeStr: CmdDestroyReqTypeStr,
		Content: string(jsonMustMarshal(&destroyReq)),
	}
	cmdDumplingReqWrap = CmdReqOrResp{
		TypeStr: CmdDumplingReqTypeStr,
		Content: string(jsonMustMarshal(&dumplingReq)),
	}
	cmdLightningReqWrap = CmdReqOrResp{
		TypeStr: CmdLightningReqTypeStr,
		Content: string(jsonMustMarshal(&lightningReq)),
	}
	cmdClusterDisplayReqWrap = CmdReqOrResp{
		TypeStr: CmdClusterDisplayReqTypeStr,
		Content: string(jsonMustMarshal(&clusterDisplayReq)),
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

func TestTiUPCmd_TiupMgrInit(t *testing.T) {
	TiupMgrInit()
	mgrTaskStatusChCap := cap(glMgrTaskStatusCh)
	mgrTaskStatusMapLen := len(glMgrTaskStatusMap)
	if mgrTaskStatusChCap != 1024 {
		t.Errorf("glMgrTaskStatusCh cap was incorrect, got: %d, want: %d.", mgrTaskStatusChCap, 1024)
	}
	if mgrTaskStatusMapLen != 0 {
		t.Errorf("mgrTaskStatusMap len was incorrect, got: %d, want: %d.", mgrTaskStatusMapLen, 0)
	}
}

func TestTiUPCmd_TiupMgrRoutine(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The test did not panic")
		}
	}()
	TiupMgrRoutine()
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

func Test_recvReqAndSndResp_deploy(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdDeployReqWrap))))
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

func Test_recvReqAndSndResp_start(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdStartReqWrap))))
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

func Test_recvReqAndSndResp_list(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdListReqWrap))))
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

func Test_recvReqAndSndResp_destroy(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdDestroyReqWrap))))
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

func Test_recvReqAndSndResp_dumpling(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdDumplingReqWrap))))
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

func Test_recvReqAndSndResp_lightning(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdLightningReqWrap))))
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

func Test_recvReqAndSndResp_clusterDisplay(t *testing.T) {
	content := []byte(fmt.Sprintf("%s\n", string(jsonMustMarshal(&cmdClusterDisplayReqWrap))))
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
