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

package secondparty

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
	"time"

	"github.com/pingcap-inc/tiem/library/framework"
)

type FieldKey string

func myPanic(v interface{}) {
	s := fmt.Sprint(v)
	framework.Log().Errorf("panic: %s, with stack trace: %s", s, string(debug.Stack()))
	panic("unexpected" + s)
}

func newTmpFileWithContent(filePrefix string, content []byte) (fileName string, err error) {
	tmpfile, err := ioutil.TempFile("", fmt.Sprintf("%s-*.yaml", filePrefix))
	if err != nil {
		err = fmt.Errorf("fail to create temp file err: %s", err)
		return "", err
	}
	fileName = tmpfile.Name()
	var ct int
	ct, err = tmpfile.Write(content)
	if err != nil || ct != len(content) {
		tmpfile.Close()
		os.Remove(fileName)
		err = fmt.Errorf(fmt.Sprint("fail to write content to temp file ", fileName, "err:", err, "length of content:", "writed:", ct))
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		myPanic(fmt.Sprintln("fail to close temp file ", fileName))
	}
	return fileName, nil
}

func jsonMustMarshal(v interface{}) []byte {
	bs, err := json.Marshal(v)
	if err != nil {
		return []byte(fmt.Sprintf("fail marshal %v: %v", v, err))
	}
	return bs
}

func execShowBackUpInfoThruSQL(ctx context.Context, db *sql.DB, showBackupSQLCmd string) (resp CmdShowBackUpInfoResp) {
	t0 := time.Now()
	err := db.QueryRow(showBackupSQLCmd).Scan(&resp.Destination, &resp.State, &resp.Progress, &resp.QueueTime, &resp.ExecutionTime, &resp.FinishTime, &resp.Connection)
	logInFunc := framework.LogWithContext(ctx)
	successFp := func() {
		logInFunc.Info("showbackupinfo task finished, time cost", time.Since(t0))
	}
	if err != nil {
		logInFunc.Errorf("query sql cmd err: %v", err)
		if err.Error() != "sql: no rows in result set" {
			logInFunc.Debugf("(%s) != (sql: no rows in result set", err.Error())
			resp.ErrorStr = err.Error()
			return
		}
		logInFunc.Debugf("(%s) == (sql: no rows in result set)", err.Error())
		logInFunc.Infof("task has finished without checking db while no rows is result for sql cmd")
		resp.Progress = 100
		return
	}
	logInFunc.Info("sql cmd return successfully")
	successFp()
	return
}

func execShowRestoreInfoThruSQL(ctx context.Context, db *sql.DB, showRestoreSQLCmd string) (resp CmdShowRestoreInfoResp) {
	t0 := time.Now()
	err := db.QueryRow(showRestoreSQLCmd).Scan(&resp.Destination, &resp.State, &resp.Progress, &resp.QueueTime, &resp.ExecutionTime, &resp.FinishTime, &resp.Connection)
	logInFunc := framework.LogWithContext(ctx)
	successFp := func() {
		logInFunc.Info("showretoreinfo task finished, time cost", time.Since(t0))
	}
	if err != nil {
		logInFunc.Errorf("query sql cmd err: %v", err)
		if err.Error() != "sql: no rows in result set" {
			logInFunc.Debugf("(%s) != (sql: no rows in result set", err.Error())
			resp.ErrorStr = err.Error()
			return resp
		}
		logInFunc.Debugf("(%s) == (sql: no rows in result set)", err.Error())
		logInFunc.Infof("task has finished without checking db while no rows is result for sql cmd")
		resp.Progress = 100
		return resp
	}
	logInFunc.Info("sql cmd return successfully")
	successFp()
	return resp
}
