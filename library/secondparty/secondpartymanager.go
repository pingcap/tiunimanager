package secondparty

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	dbPb "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

type MicroSrv interface {
	MicroInit(mgrLogFilePath string)
	MicroSrvTiupDeploy(instanceName string, version string, configStrYaml string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupList(timeoutS int, flags []string) (resp *CmdListResp, err error)
	MicroSrvTiupStart(instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupDestroy(instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupDumpling(timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupLightning(timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupClusterDisplay(clusterName string, timeoutS int, flags []string) (resp *CmdClusterDisplayResp, err error)
	MicroSrvTiupGetTaskStatus(taskID uint64) (stat dbPb.TiupTaskStatus, errStr string, err error)
	MicroSrvTiupGetTaskStatusByBizID(bizID uint64) (stat dbPb.TiupTaskStatus, statErrStr string, err error)
	BackUp(cluster ClusterFacade, storage BrStorage, bizId uint64) (taskID uint64, err error)
	ShowBackUpInfo(cluster ClusterFacade) CmdShowBackUpInfoResp
	Restore(cluster ClusterFacade, storage BrStorage, bizId uint64) (taskID uint64, err error)
	ShowRestoreInfo(cluster ClusterFacade) CmdShowRestoreInfoResp
	taskStatusMapSyncer()
	startGetAllValidTaskStatusTask() CmdGetAllTaskStatusResp
	getAllValidTaskStatus() (ret []TaskStatusMember)
	startNewTiupDeployTask(taskID uint64, req *CmdDeployReq)
	startNewTiupStartTask(taskID uint64, req *CmdStartReq)
	startNewTiupListTask(req *CmdListReq) (resp CmdListResp, err error)
	startNewTiupDestroyTask(taskID uint64, req *CmdDestroyReq)
	startNewTiupDumplingTask(taskID uint64, req *CmdDumplingReq)
	startNewTiupLightningTask(taskID uint64, req *CmdLightningReq)
	startNewTiupClusterDisplayTask(req *CmdClusterDisplayReq) (resp CmdClusterDisplayResp, err error)
	startNewTiupTask(taskID uint64, tiupPath string, tiupArgs []string, TimeoutS int) (exitCh chan struct{})
	startNewBrBackUpTaskThruSQL(taskID uint64, req *CmdBackUpReq)
	startNewBrShowBackUpInfoThruSQL(req *CmdShowBackUpInfoReq) CmdShowBackUpInfoResp
	startNewBrRestoreTaskThruSQL(taskID uint64, req *CmdRestoreReq)
	startNewBrShowRestoreInfoThruSQL(req *CmdShowRestoreInfoReq) CmdShowRestoreInfoResp
	startNewBrTaskThruSQL(taskID uint64, dbConnParam *DbConnParam, brSQLCmd string) (exitCh chan struct{})
}

type SecondMicro struct {
	TiupBinPath 			string
	taskStatusCh 			chan TaskStatusMember
	taskStatusMap 			map[uint64]TaskStatusMapValue
	syncedTaskStatusMap     map[uint64]TaskStatusMapValue
	taskStatusMapMutex 		sync.Mutex
}

var logger *logrus.Entry

func (secondMicro *SecondMicro) MicroInit(mgrLogFilePath string) {
	configPath := ""
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	logger = framework.LogForkFile(configPath + common.LogFileLibTiup)

	secondMicro.syncedTaskStatusMap = make(map[uint64]TaskStatusMapValue)
	secondMicro.taskStatusCh = make(chan TaskStatusMember, 1024)
	secondMicro.taskStatusMap = make(map[uint64]TaskStatusMapValue)
	go secondMicro.taskStatusMapSyncer()
}

func (secondMicro *SecondMicro) MicroSrvTiupGetTaskStatus(taskID uint64) (stat dbPb.TiupTaskStatus, errStr string, err error) {
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

func (secondMicro *SecondMicro) MicroSrvTiupGetTaskStatusByBizID(bizID uint64) (stat dbPb.TiupTaskStatus, statErrStr string, err error) {
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
