
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package libbr

import (
	"fmt"
	"testing"
)

var dbConnParam DbConnParam
var storage BrStorage
var cmdBackUpReq CmdBackUpReq
var cmdShowBackUpInfoReq CmdShowBackUpInfoReq
var cmdRestoreReq CmdRestoreReq
var cmdShowRestoreInfoReq CmdShowRestoreInfoReq

func init() {
	fmt.Println("init before test")
	dbConnParam = DbConnParam{
		Username: "root",
		Ip:       "127.0.0.1",
		Port:     "57395",
	}
	storage = BrStorage{
		StorageType: StorageTypeLocal,
		Root:        "/tmp/backup",
	}

	cmdBackUpReq = CmdBackUpReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
		StorageAddress:  fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root),
	}
	cmdShowBackUpInfoReq = CmdShowBackUpInfoReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
	}
	cmdRestoreReq = CmdRestoreReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
		StorageAddress:  fmt.Sprintf("%s://%s", string(storage.StorageType), storage.Root),
	}
	cmdShowRestoreInfoReq = CmdShowRestoreInfoReq{
		TaskID:          0,
		DbConnParameter: dbConnParam,
	}
	//log = logger.GetRootLogger("config_key_test_log")
	BrMgrInit()
	// MicroInit("../../../bin/micro-cluster/brmgr/brmgr", "/tmp/log/br/")
}

func TestBrMgrInit(t *testing.T) {
	BrMgrInit()
	mgrTaskStatusChCap := cap(glMgrTaskStatusCh)
	if mgrTaskStatusChCap != 1024 {
		t.Errorf("glMgrTaskStatusCh cap was incorrect, got: %d, want: %d.", mgrTaskStatusChCap, 1024)
	}
}

//// todo: need to start a TiDB cluster and generate some data before running this test
//func TestBackUpAndInfo(t *testing.T) {
//	mgrHandleCmdBackUpReq(string(jsonMustMarshal(cmdBackUpReq)))
//	loopUntilBackupDone()
//}
//
//// todo: need to start a TiDB cluster and generate some data before running this test
//func TestRestoreAndInfo(t *testing.T) {
//	mgrHandleCmdRestoreReq(string(jsonMustMarshal(cmdRestoreReq)))
//	loopUntilRestoreDone()
//}