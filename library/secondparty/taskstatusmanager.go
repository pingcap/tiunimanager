package secondparty

import (
	"context"
	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"time"
)

type TaskStatus int

const (
	TaskStatusInit  	 	= TaskStatus(dbPb.TiupTaskStatus_Init)
	TaskStatusProcessing  	= TaskStatus(dbPb.TiupTaskStatus_Processing)
	TaskStatusFinished   	= TaskStatus(dbPb.TiupTaskStatus_Finished)
	TaskStatusError			= TaskStatus(dbPb.TiupTaskStatus_Error)
)

type TaskStatusMember struct {
	TaskID   uint64
	Status   TaskStatus
	ErrorStr string
}

type TaskStatusMapValue struct {
	validFlag bool // flag to help check if the value in the taskStatusMap
	stat   TaskStatusMember
	readct uint64 // the count that the value has been read
}

// sync(put or update) all TaskStatus to database(synced in tiUPMicro.syncedTaskStatusMap) from taskStatusMap(valid ones)
func (secondMicro *SecondMicro) taskStatusMapSyncer() {
	for {
		time.Sleep(time.Second)
		resp := secondMicro.startGetAllValidTaskStatusTask()
		var needDbUpdate []TaskStatusMember
		secondMicro.taskStatusMapMutex.Lock()
		for _, v := range resp.Stats {
			oldv := secondMicro.syncedTaskStatusMap[v.TaskID]
			if oldv.validFlag {
				if oldv.stat.Status == v.Status {
					assert(oldv.stat == v)
				} else {
					assert(oldv.stat.Status == TaskStatusProcessing)
					secondMicro.syncedTaskStatusMap[v.TaskID] = TaskStatusMapValue{
						validFlag: true,
						stat:      v,
						readct:    0,
					}
					assert(v.Status == TaskStatusFinished || v.Status == TaskStatusError)
					needDbUpdate = append(needDbUpdate, v)
				}
			} else {
				secondMicro.syncedTaskStatusMap[v.TaskID] = TaskStatusMapValue{
					validFlag: true,
					stat:      v,
					readct:    0,
				}
				needDbUpdate = append(needDbUpdate, v)
			}
		}
		secondMicro.taskStatusMapMutex.Unlock()
		logInFunc := logger.WithField("taskStatusMapSyncer", "DbClient.UpdateTiupTask")
		for _, v := range needDbUpdate {
			rsp, err := client.DBClient.UpdateTiupTask(context.Background(), &dbPb.UpdateTiupTaskRequest{
				Id:     v.TaskID,
				Status: dbPb.TiupTaskStatus(v.Status),
				ErrStr: v.ErrorStr,
			})
			if rsp == nil || err != nil || rsp.ErrCode != 0 {
				logInFunc.Error("rsp:", rsp, "err:", err, "v:", v)
			} else {
				logInFunc.Debug("update success:", v)
			}
		}
	}
}

func (secondMicro *SecondMicro) startGetAllValidTaskStatusTask() CmdGetAllTaskStatusResp {
	secondMicro.syncTaskStatusMap()
	return CmdGetAllTaskStatusResp{
		Stats: secondMicro.getAllValidTaskStatus(),
	}
}

// sync(put or update) all TaskStatus to memory(tiUPMicro.taskStatusMap) from taskStatusCh which is sent by async task in sub go routine
func (secondMicro *SecondMicro) syncTaskStatusMap() {
	for {
		var consumedFlag bool
		var statm TaskStatusMember
		var ok bool
		select {
		case statm, ok = <-secondMicro.taskStatusCh:
			assert(ok == true)
			consumedFlag = true
		default:
		}
		if consumedFlag {
			v := secondMicro.taskStatusMap[statm.TaskID]
			if v.validFlag {
				assert(v.stat.Status == TaskStatusProcessing)
				assert(statm.Status == TaskStatusFinished || statm.Status == TaskStatusError)
			} else {
				assert(statm.Status == TaskStatusProcessing)
			}
			secondMicro.taskStatusMap[statm.TaskID] = TaskStatusMapValue{
				validFlag: true,
				readct:    0,
				stat:      statm,
			}
		} else {
			break
		}
	}
}

/**
	1. delete all TaskStatusFinished and TaskStatusError tasks which have been read before from the memory
	2. put all the rest TaskStatusMember into the result list
	3. increment the read count of rest TaskStatusMember
 */
func (secondMicro *SecondMicro) getAllValidTaskStatus() (ret []TaskStatusMember) {
	var needDeleteTaskList []uint64
	for k, v := range secondMicro.taskStatusMap {
		if v.readct > 0 && (v.stat.Status == TaskStatusFinished || v.stat.Status == TaskStatusError) {
			needDeleteTaskList = append(needDeleteTaskList, k)
		}
	}
	for _, k := range needDeleteTaskList {
		delete(secondMicro.taskStatusMap, k)
	}
	for k, v := range secondMicro.taskStatusMap {
		assert(k == v.stat.TaskID)
		ret = append(ret, v.stat)
	}
	for k, v := range secondMicro.taskStatusMap {
		newv := v
		newv.readct++
		secondMicro.taskStatusMap[k] = newv
	}
	return
}