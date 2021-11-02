package secondparty

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

type MicroSrv interface {
	MicroInit(mgrLogFilePath string)
	MicroSrvTiupDeploy(tiupComponent TiUPComponentTypeStr, instanceName string, version string, configStrYaml string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupStart(tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupList(tiupComponent TiUPComponentTypeStr, timeoutS int, flags []string) (resp *CmdListResp, err error)
	MicroSrvTiupDestroy(tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvTiupDisplay(tiupComponent TiUPComponentTypeStr, instanceName string, timeoutS int, flags []string) (resp *CmdDisplayResp, err error)
	MicroSrvDumpling(timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvLightning(timeoutS int, flags []string, bizID uint64) (taskID uint64, err error)
	MicroSrvBackUp(cluster ClusterFacade, storage BrStorage, bizId uint64) (taskID uint64, err error)
	MicroSrvShowBackUpInfo(cluster ClusterFacade) CmdShowBackUpInfoResp
	MicroSrvRestore(cluster ClusterFacade, storage BrStorage, bizId uint64) (taskID uint64, err error)
	MicroSrvShowRestoreInfo(cluster ClusterFacade) CmdShowRestoreInfoResp
	MicroSrvGetTaskStatus(taskID uint64) (stat dbPb.TiupTaskStatus, errStr string, err error)
	MicroSrvGetTaskStatusByBizID(bizID uint64) (stat dbPb.TiupTaskStatus, statErrStr string, err error)
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

func (secondMicro *SecondMicro) MicroSrvGetTaskStatus(taskID uint64) (stat dbPb.TiupTaskStatus, errStr string, err error) {
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

func (secondMicro *SecondMicro) MicroSrvGetTaskStatusByBizID(bizID uint64) (stat dbPb.TiupTaskStatus, statErrStr string, err error) {
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
