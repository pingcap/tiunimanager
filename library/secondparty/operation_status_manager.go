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
 * @File: operation_status_manager.go
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/8
*******************************************************************************/

package secondparty

import (
	"context"
	"time"

	"github.com/pingcap-inc/tiem/models"

	"github.com/pingcap-inc/tiem/models/workflow/secondparty"

	"github.com/pingcap-inc/tiem/library/framework"
)

type OperationStatusMember struct {
	OperationID string
	Status      secondparty.OperationStatus
	Result      string
	ErrorStr    string
}

type OperationStatusMapValue struct {
	validFlag bool // flag to help check if the value in the operationStatusMap
	stat      OperationStatusMember
	readct    uint64 // the count that the value has been read
}

// sync(put or update) all OperationStatus to database(synced in tiUPMicro.syncedOperationStatusMap) from operationStatusMap(valid ones)
func (manager *SecondPartyManager) operationStatusMapSyncer() {
	for {
		time.Sleep(time.Second)
		resp := manager.startGetAllValidOperationStatus()
		var needDbUpdate []OperationStatusMember
		manager.operationStatusMapMutex.Lock()
		for _, v := range resp.Stats {
			oldv := manager.syncedOperationStatusMap[v.OperationID]
			if oldv.validFlag {
				if oldv.stat.Status == v.Status {
					assert(oldv.stat == v)
				} else {
					assert(oldv.stat.Status == secondparty.OperationStatusProcessing)
					manager.syncedOperationStatusMap[v.OperationID] = OperationStatusMapValue{
						validFlag: true,
						stat:      v,
						readct:    0,
					}
					assert(v.Status == secondparty.OperationStatusFinished || v.Status == secondparty.OperationStatusError)
					needDbUpdate = append(needDbUpdate, v)
				}
			} else {
				manager.syncedOperationStatusMap[v.OperationID] = OperationStatusMapValue{
					validFlag: true,
					stat:      v,
					readct:    0,
				}
				needDbUpdate = append(needDbUpdate, v)
			}
		}
		manager.operationStatusMapMutex.Unlock()
		logInFunc := framework.Log().WithField("operationStatusMapSyncer", "secondPartyOperationReaderWriter.Update")
		for _, v := range needDbUpdate {
			err := models.GetSecondPartyOperationReaderWriter().Update(context.Background(), &secondparty.SecondPartyOperation{
				ID:       v.OperationID,
				Status:   v.Status,
				Result:   v.Result,
				ErrorStr: v.ErrorStr,
			})
			if err != nil {
				logInFunc.Error("err:", err, "v:", v)
			} else {
				logInFunc.Debug("update success:", v)
			}
		}
	}
}

func (manager *SecondPartyManager) startGetAllValidOperationStatus() CmdGetAllOperationStatusResp {
	manager.syncOperationStatusMap()
	return CmdGetAllOperationStatusResp{
		Stats: manager.getAllValidOperationStatus(),
	}
}

// sync(put or update) all OperationStatus to memory(tiUPMicro.operationStatusMap) from operationStatusCh which is sent by async task in sub go routine
func (manager *SecondPartyManager) syncOperationStatusMap() {
	for {
		var consumedFlag bool
		var statm OperationStatusMember
		var ok bool
		select {
		case statm, ok = <-manager.operationStatusCh:
			assert(ok)
			consumedFlag = true
		default:
		}
		if consumedFlag {
			v := manager.operationStatusMap[statm.OperationID]
			if v.validFlag {
				assert(v.stat.Status == secondparty.OperationStatusProcessing)
				assert(statm.Status == secondparty.OperationStatusFinished || statm.Status == secondparty.OperationStatusError)
			} else {
				assert(statm.Status == secondparty.OperationStatusProcessing)
			}
			manager.operationStatusMap[statm.OperationID] = OperationStatusMapValue{
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
1. delete all OperationStatusFinished and OperationStatusError operations which have been read before from the memory
2. put all the rest OperationStatusMember into the result list
3. increment the read count of rest OperationStatusMember
*/
func (manager *SecondPartyManager) getAllValidOperationStatus() (ret []OperationStatusMember) {
	var needDeleteOperationList []string
	for k, v := range manager.operationStatusMap {
		if v.readct > 0 && (v.stat.Status == secondparty.OperationStatusFinished || v.stat.Status == secondparty.OperationStatusError) {
			needDeleteOperationList = append(needDeleteOperationList, k)
		}
	}
	for _, k := range needDeleteOperationList {
		delete(manager.operationStatusMap, k)
	}
	for k, v := range manager.operationStatusMap {
		assert(k == v.stat.OperationID)
		ret = append(ret, v.stat)
	}
	for k, v := range manager.operationStatusMap {
		newv := v
		newv.readct++
		manager.operationStatusMap[k] = newv
	}
	return
}
