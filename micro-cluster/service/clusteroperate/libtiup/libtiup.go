package libtiup

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/pingcap/ticp/addon/logger"
	"github.com/pingcap/ticp/micro-metadb/client"
	dbPb "github.com/pingcap/ticp/micro-metadb/proto"
)

// micro service --fork&exec--> tiup manager --fork&exec--> tiup process
//                                          |         ......
//                                          |--fork&exec--> tiup process

type CmdTypeStr string

const (
	CmdDeployReqTypeStr            CmdTypeStr = "CmdDeployReq"
	CmdDeployRespTypeStr           CmdTypeStr = "CmdDeployResp"
	CmdStartReqTypeStr             CmdTypeStr = "CmdStartReq"
	CmdStartRespTypeStr            CmdTypeStr = "CmdStartResp"
	CmdListReqTypeStr              CmdTypeStr = "CmdListReq"
	CmdListRespTypeStr             CmdTypeStr = "CmdListResp"
	CmdDestroyReqTypeStr           CmdTypeStr = "CmdDestroyReq"
	CmdDestroyRespTypeStr          CmdTypeStr = "CmdDestroyResp"
	CmdGetAllTaskStatusReqTypeStr  CmdTypeStr = "CmdGetAllTaskStatusReq"
	CmdGetAllTaskStatusRespTypeStr CmdTypeStr = "CmdGetAllTaskStatusResp"
)

type CmdReqOrResp struct {
	TypeStr CmdTypeStr
	Content string
}

type CmdDeployReq struct {
	TaskID        uint64
	InstanceName  string
	Version       string
	ConfigStrYaml string
	TimeoutS      int
	TiupPath      string
	Flags         []string
}

type CmdStartReq struct {
	TaskID       uint64
	InstanceName string
	TimeoutS     int
	TiupPath     string
	Flags        []string
}

type CmdListReq struct {
	TaskID   uint64
	TimeoutS int
	TiupPath string
	Flags    []string
}

type CmdDestroyReq struct {
	TaskID       uint64
	InstanceName string
	TimeoutS     int
	TiupPath     string
	Flags        []string
}

type CmdDeployResp struct {
}

type CmdStartResp struct {
}

type CmdListResp struct {
}

type CmdDestroyResp struct {
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

type CmdGetAllTaskStatusReq struct {
}

type CmdGetAllTaskStatusResp struct {
	Stats []TaskStatusMember
}

type TaskStatusMapValue struct {
	validFlag bool
	stat      TaskStatusMember
	readct    uint64
}

var glMgrTaskStatusCh chan TaskStatusMember
var glMgrTaskStatusMap map[uint64]TaskStatusMapValue

func TiupMgrInit() {
	glMgrTaskStatusCh = make(chan TaskStatusMember, 1024)
	glMgrTaskStatusMap = make(map[uint64]TaskStatusMapValue)
}

func assert(b bool) {
	if b {
	} else {
		panic("unexpected")
	}
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
			assert(ok == true)
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

func mgrHandleCmdDeployReq(jsonStr string) CmdDeployResp {
	ret := CmdDeployResp{}
	var req CmdDeployReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		panic(fmt.Sprintln("json.Unmarshal CmdDeployReq failed err:", err))
	}
	mgrStartNewTiupDeployTask(req.TaskID, &req)
	return ret
}

func mgrHandleCmdStarReq(jsonStr string) CmdStartResp {
	ret := CmdStartResp{}
	var req CmdStartReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		panic(fmt.Sprintln("json.Unmarshal CmdStartReq failed err:", err))
	}
	mgrStartNewTiupStartTask(req.TaskID, &req)
	return ret
}

func mgrHandleCmdListReq(jsonStr string) CmdListResp {
	ret := CmdListResp{}
	var req CmdListReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		panic(fmt.Sprintln("json.Unmarshal CmdListReq failed err:", err))
	}
	mgrStartNewTiupListTask(req.TaskID, &req)
	return ret
}

func mgrHandleCmdDestroyReq(jsonStr string) CmdDestroyResp {
	ret := CmdDestroyResp{}
	var req CmdDestroyReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		panic(fmt.Sprintln("json.Unmarshal CmdDestroyReq failed err:", err))
	}
	mgrStartNewTiupDestroyTask(req.TaskID, &req)
	return ret
}

func mgrHandleCmdGetAllTaskStatusReq(jsonStr string) CmdGetAllTaskStatusResp {
	var req CmdGetAllTaskStatusReq
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		panic(fmt.Sprintln("json.Unmarshal CmdGetAllTaskStatusReq failed err:", err))
	}
	glMgrStatusMapSync()
	return CmdGetAllTaskStatusResp{
		Stats: glMgrStatusMapGetAll(),
	}
}

func newTmpFileWithContent(content []byte) (fileName string, err error) {
	tmpfile, err := ioutil.TempFile("", "ticp-topology-*.yaml")
	if err != nil {
		err = fmt.Errorf("fail to create temp file err:", err)
		return "", err
	}
	fileName = tmpfile.Name()
	var ct int
	ct, err = tmpfile.Write(content)
	if err != nil || ct != len(content) {
		tmpfile.Close()
		os.Remove(fileName)
		err = fmt.Errorf("fail to write content to temp file ", fileName, "err:", err, "length of content:", "writed:", ct)
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		panic(fmt.Sprintln("fail to close temp file ", fileName))
	}
	return fileName, nil
}

func mgrStartNewTiupTask(taskID uint64, tiupPath string, tiupArgs []string, TimeoutS int) (exitCh chan struct{}) {
	exitCh = make(chan struct{})
	glMgrTaskStatusCh <- TaskStatusMember{
		TaskID:   taskID,
		Status:   TaskStatusProcessing,
		ErrorStr: "",
	}
	go func() {
		defer close(exitCh)
		var cmd *exec.Cmd
		if TimeoutS != 0 {
			ctx, _ := context.WithTimeout(context.Background(), time.Duration(TimeoutS)*time.Second)
			exec.CommandContext(ctx, tiupPath, tiupArgs...)
		} else {
			cmd = exec.Command(tiupPath, tiupArgs...)
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
		}
		if err := cmd.Start(); err != nil {
			glMgrTaskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintln(err),
			}
			return
		}
		successFp := func() {
			glMgrTaskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusFinished,
				ErrorStr: "",
			}
		}
		err := cmd.Wait()
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
					if status.ExitStatus() == 0 {
						successFp()
						return
					}
				}
			}
			glMgrTaskStatusCh <- TaskStatusMember{
				TaskID:   taskID,
				Status:   TaskStatusError,
				ErrorStr: fmt.Sprintln(err),
			}
			return
		} else {
			successFp()
			return
		}
	}()
	return exitCh
}

func mgrStartNewTiupDeployTask(taskID uint64, req *CmdDeployReq) {
	topologyTmpFilePath, err := newTmpFileWithContent([]byte(req.ConfigStrYaml))
	if err != nil {
		glMgrTaskStatusCh <- TaskStatusMember{
			TaskID:   taskID,
			Status:   TaskStatusError,
			ErrorStr: fmt.Sprintln(err),
		}
		return
	}
	go func() {
		defer os.Remove(topologyTmpFilePath)
		var args []string
		args = append(args, "cluster", "deploy", req.InstanceName, req.Version, topologyTmpFilePath)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-mgrStartNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func mgrStartNewTiupStartTask(taskID uint64, req *CmdStartReq) {
	go func() {
		var args []string
		args = append(args, "cluster", "start", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-mgrStartNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func mgrStartNewTiupListTask(taskID uint64, req *CmdListReq) {
	go func() {
		var args []string
		args = append(args, "cluster", "list")
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-mgrStartNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func mgrStartNewTiupDestroyTask(taskID uint64, req *CmdDestroyReq) {
	go func() {
		var args []string
		args = append(args, "cluster", "destroy", req.InstanceName)
		args = append(args, req.Flags...)
		args = append(args, "--yes")
		<-mgrStartNewTiupTask(taskID, req.TiupPath, args, req.TimeoutS)
	}()
}

func TiupMgrRoutine() {
	inReader := bufio.NewReader(os.Stdin)
	outWriter := os.Stdout
	errw := os.Stderr
	//errw.Write([]byte("TiupMgrRoutine enter\n"))
	for {
		//errw.Write([]byte("TiupMgrRoutine read\n"))
		input, err := inReader.ReadString('\n')
		//fmt.Println("input:", input, len(input), input[:len(input)-1], len(input[:len(input)-1]))
		if err != nil {
			panic(err)
		}
		errw.Write([]byte(input))
		if input[len(input)-1] == '\n' {
			cmdStr := input[:len(input)-1]
			var cmd CmdReqOrResp
			err := json.Unmarshal([]byte(cmdStr), &cmd)
			if err != nil {
				panic(fmt.Sprintln("cmdStr unmarshal failed err:", err, "cmdStr:", cmdStr))
			}
			var cmdResp CmdReqOrResp
			switch cmd.TypeStr {
			case CmdDeployReqTypeStr:
				resp := mgrHandleCmdDeployReq(cmd.Content)
				cmdResp.TypeStr = CmdDeployRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			case CmdStartReqTypeStr:
				resp := mgrHandleCmdStarReq(cmd.Content)
				cmdResp.TypeStr = CmdStartRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			case CmdListReqTypeStr:
				resp := mgrHandleCmdListReq(cmd.Content)
				cmdResp.TypeStr = CmdListRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			case CmdDestroyReqTypeStr:
				resp := mgrHandleCmdDestroyReq(cmd.Content)
				cmdResp.TypeStr = CmdDeployRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			case CmdGetAllTaskStatusReqTypeStr:
				resp := mgrHandleCmdGetAllTaskStatusReq(cmd.Content)
				cmdResp.TypeStr = CmdGetAllTaskStatusRespTypeStr
				cmdResp.Content = string(jsonMustMarshal(&resp))
			default:
				panic(fmt.Sprintln("unknown cmdStr.TypeStr:", cmd.TypeStr))
			}
			bs := jsonMustMarshal(&cmdResp)
			bs = append(bs, '\n')
			//errw.Write([]byte("TiupMgrRoutine write\n"))
			ct, err := outWriter.Write(bs)
			//errw.Write([]byte(fmt.Sprintf("TiupMgrRoutine write finished %s %d %d %d %v\n", bs, len(bs), ct, err == nil, err)))
			assert(ct == len(bs))
			assert(err == nil)
		} else {
			panic("unexpected")
		}
	}
}

var glMicroCmdChan chan CmdChanMember

var glMicroTaskStatusMap map[uint64]TaskStatusMapValue
var glMicroTaskStatusMapMutex sync.Mutex

var glTiUPMgrPath string
var glTiUPBinPath string

func MicroInit(tiupMgrPath, tiupBinPath string) {
	glTiUPMgrPath = tiupMgrPath
	glTiUPBinPath = tiupBinPath
	glMicroTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	glMicroCmdChan = microStartTiupMgr()
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
		log := logger.WithContext(nil).WithField("glMicroTaskStatusMapSyncer", "DbClient.UpdateTiupTask")
		for _, v := range needDbUpdate {
			rsp, err := client.DBClient.UpdateTiupTask(context.Background(), &dbPb.UpdateTiupTaskRequest{
				Id:     v.TaskID,
				Status: dbPb.TiupTaskStatus(v.Status),
				ErrStr: v.ErrorStr,
			})
			if rsp == nil || err != nil || rsp.ErrCode != 0 {
				log.Error("rsp:", rsp, "err:", err, "v:", v)
			} else {
				log.Debug("update succes:", v)
			}
		}
	}
}

func microTiupDeploy(deployReq CmdDeployReq) CmdDeployResp {
	assert(cap(glMicroCmdChan) > 0)
	cmdReq := CmdReqOrResp{
		TypeStr: CmdDeployReqTypeStr,
		Content: string(jsonMustMarshal(&deployReq)),
	}
	respCh := make(chan CmdReqOrResp, 1)
	glMicroCmdChan <- CmdChanMember{
		req:    cmdReq,
		respCh: respCh,
	}
	respCmd := <-respCh
	assert(respCmd.TypeStr == CmdDeployRespTypeStr)
	var resp CmdDeployResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

func MicroSrvTiupDeploy(instanceName string, version string, configStrYaml string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Deploy
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var deployReq CmdDeployReq
		deployReq.InstanceName = instanceName
		deployReq.Version = version
		deployReq.ConfigStrYaml = configStrYaml
		deployReq.TimeoutS = timeoutS
		deployReq.Flags = flags
		deployReq.TiupPath = glTiUPBinPath
		deployReq.TaskID = rsp.Id
		microTiupDeploy(deployReq)
		return rsp.Id, nil
	}
}

func microTiupList(req CmdListReq) CmdListResp {
	assert(cap(glMicroCmdChan) > 0)
	cmdReq := CmdReqOrResp{
		TypeStr: CmdListReqTypeStr,
		Content: string(jsonMustMarshal(&req)),
	}
	respCh := make(chan CmdReqOrResp, 1)
	glMicroCmdChan <- CmdChanMember{
		req:    cmdReq,
		respCh: respCh,
	}
	respCmd := <-respCh
	assert(respCmd.TypeStr == CmdListRespTypeStr)
	var resp CmdListResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

func MicroSrvTiupList(timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_List
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		//fmt.Println(rsp, err)
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var req CmdListReq
		req.TaskID = rsp.Id
		req.TimeoutS = timeoutS
		req.TiupPath = glTiUPBinPath
		req.Flags = flags
		microTiupList(req)
		return rsp.Id, nil
	}
}

func microTiupStart(req CmdStartReq) CmdStartResp {
	assert(cap(glMicroCmdChan) > 0)
	cmdReq := CmdReqOrResp{
		TypeStr: CmdStartReqTypeStr,
		Content: string(jsonMustMarshal(&req)),
	}
	respCh := make(chan CmdReqOrResp, 1)
	glMicroCmdChan <- CmdChanMember{
		req:    cmdReq,
		respCh: respCh,
	}
	respCmd := <-respCh
	assert(respCmd.TypeStr == CmdStartRespTypeStr)
	var resp CmdStartResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

func MicroSrvTiupStart(instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Start
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var req CmdStartReq
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = glTiUPBinPath
		req.Flags = flags
		microTiupStart(req)
		return rsp.Id, nil
	}
}

func microTiupDestroy(req CmdDestroyReq) CmdDestroyResp {
	assert(cap(glMicroCmdChan) > 0)
	cmdReq := CmdReqOrResp{
		TypeStr: CmdDestroyReqTypeStr,
		Content: string(jsonMustMarshal(&req)),
	}
	respCh := make(chan CmdReqOrResp, 1)
	glMicroCmdChan <- CmdChanMember{
		req:    cmdReq,
		respCh: respCh,
	}
	respCmd := <-respCh
	assert(respCmd.TypeStr == CmdDestroyRespTypeStr)
	var resp CmdDestroyResp
	err := json.Unmarshal([]byte(respCmd.Content), &resp)
	assert(err == nil)
	return resp
}

func MicroSrvTiupDestroy(instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error) {
	var req dbPb.CreateTiupTaskRequest
	req.Type = dbPb.TiupTaskType_Destroy
	req.BizID = bizID
	rsp, err := client.DBClient.CreateTiupTask(context.Background(), &req)
	if rsp == nil || err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("rsp:%v, err:%s", err, rsp)
		return 0, err
	} else {
		var req CmdDestroyReq
		req.TaskID = rsp.Id
		req.InstanceName = instanceName
		req.TimeoutS = timeoutS
		req.TiupPath = glTiUPBinPath
		req.Flags = flags
		microTiupDestroy(req)
		return rsp.Id, nil
	}
}

func MicroSrvTiupGetTaskStatus(taskID uint64) (stat dbPb.TiupTaskStatus, errStr string, err error) {
	var req dbPb.FindTiupTaskByIDRequest
	req.Id = taskID
	rsp, err := client.DBClient.FindTiupTaskByID(context.Background(), &req)
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

func MicroSrvTiupGetTaskStatusByBizID(bizID uint64) (stat dbPb.TiupTaskStatus, statErrStr string, err error) {
	var req dbPb.GetTiupTaskStatusByBizIDRequest
	req.BizID = bizID
	rsp, err := client.DBClient.GetTiupTaskStatusByBizID(context.Background(), &req)
	if err != nil || rsp.ErrCode != 0 {
		err = fmt.Errorf("err:%s, rsp.ErrCode:%d, rsp.ErrStr:%s", err, rsp.ErrCode, rsp.ErrStr)
		return stat, "", err
	} else {
		return rsp.Stat, rsp.StatErrStr, nil
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

type CmdChanMember struct {
	req    CmdReqOrResp
	respCh chan CmdReqOrResp
}

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

func microStartTiupMgr() chan CmdChanMember {
	tiupMgrPath := glTiUPMgrPath
	cmd := exec.Command(tiupMgrPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
	in, err := cmd.StdinPipe()
	if err != nil {
		//fmt.Println("err:", err)
		panic("unexpected")
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		//fmt.Println("err:", err)
		panic("unexpected")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		//fmt.Println("err:", err)
		panic("unexpected")
	}
	_ = stderr
	cch := make(chan CmdChanMember, 1024)
	if err := cmd.Start(); err != nil {
		//fmt.Println("err0:", err)
		assert(false)
	}
	go microCmdChanRoutine(cch, out, in)
	go func() {
		err = cmd.Wait()
		//fmt.Println("err1:", err)
		assert(false)
	}()
	return cch
}
