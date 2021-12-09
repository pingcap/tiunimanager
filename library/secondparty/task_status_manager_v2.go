/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: task_status_manager_v2.go
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/8
*******************************************************************************/

package secondparty

import (
	"context"
	"time"

	"github.com/pingcap-inc/tiem/library/client"
	dbPb "github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/framework"
)

// sync(put or update) all TaskStatus to database(synced in tiUPMicro.syncedTaskStatusMap) from taskStatusMap(valid ones)
func (manager *SecondPartyManager) taskStatusMapSyncer() {
	for {
		time.Sleep(time.Second)
		resp := manager.startGetAllValidTaskStatusTask()
		var needDbUpdate []TaskStatusMember
		manager.taskStatusMapMutex.Lock()
		for _, v := range resp.Stats {
			oldv := manager.syncedTaskStatusMap[v.TaskID]
			if oldv.validFlag {
				if oldv.stat.Status == v.Status {
					assert(oldv.stat == v)
				} else {
					assert(oldv.stat.Status == TaskStatusProcessing)
					manager.syncedTaskStatusMap[v.TaskID] = TaskStatusMapValue{
						validFlag: true,
						stat:      v,
						readct:    0,
					}
					assert(v.Status == TaskStatusFinished || v.Status == TaskStatusError)
					needDbUpdate = append(needDbUpdate, v)
				}
			} else {
				manager.syncedTaskStatusMap[v.TaskID] = TaskStatusMapValue{
					validFlag: true,
					stat:      v,
					readct:    0,
				}
				needDbUpdate = append(needDbUpdate, v)
			}
		}
		manager.taskStatusMapMutex.Unlock()
		logInFunc := framework.Log().WithField("taskStatusMapSyncer", "DbClient.UpdateTiupTask")
		for _, v := range needDbUpdate {
			rsp, err := client.DBClient.UpdateTiupOperatorRecord(context.Background(), &dbPb.UpdateTiupOperatorRecordRequest{
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

func (manager *SecondPartyManager) startGetAllValidTaskStatusTask() CmdGetAllTaskStatusResp {
	manager.syncTaskStatusMap()
	return CmdGetAllTaskStatusResp{
		Stats: manager.getAllValidTaskStatus(),
	}
}

// sync(put or update) all TaskStatus to memory(tiUPMicro.taskStatusMap) from taskStatusCh which is sent by async task in sub go routine
func (manager *SecondPartyManager) syncTaskStatusMap() {
	for {
		var consumedFlag bool
		var statm TaskStatusMember
		var ok bool
		select {
		case statm, ok = <-manager.taskStatusCh:
			assert(ok)
			consumedFlag = true
		default:
		}
		if consumedFlag {
			v := manager.taskStatusMap[statm.TaskID]
			if v.validFlag {
				assert(v.stat.Status == TaskStatusProcessing)
				assert(statm.Status == TaskStatusFinished || statm.Status == TaskStatusError)
			} else {
				assert(statm.Status == TaskStatusProcessing)
			}
			manager.taskStatusMap[statm.TaskID] = TaskStatusMapValue{
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
func (manager *SecondPartyManager) getAllValidTaskStatus() (ret []TaskStatusMember) {
	var needDeleteTaskList []uint64
	for k, v := range manager.taskStatusMap {
		if v.readct > 0 && (v.stat.Status == TaskStatusFinished || v.stat.Status == TaskStatusError) {
			needDeleteTaskList = append(needDeleteTaskList, k)
		}
	}
	for _, k := range needDeleteTaskList {
		delete(manager.taskStatusMap, k)
	}
	for k, v := range manager.taskStatusMap {
		assert(k == v.stat.TaskID)
		ret = append(ret, v.stat)
	}
	for k, v := range manager.taskStatusMap {
		newv := v
		newv.readct++
		manager.taskStatusMap[k] = newv
	}
	return
}
