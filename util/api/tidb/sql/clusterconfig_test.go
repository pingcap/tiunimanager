/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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
 ******************************************************************************/

/*******************************************************************************
 * @File: clusterconfig_test.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/8 14:54
*******************************************************************************/

package sql

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pingcap-inc/tiem/library/spec"
)

const TestWorkFlowNodeID = "testworkflownodeid"

var dbConnParam3 DbConnParam
var req ClusterEditConfigReq

func init() {
	dbConnParam3 = DbConnParam{
		Username: "root",
		IP:       "127.0.0.1",
		Port:     "4000",
	}
}

func TestSecondMicro_EditClusterConfig_v1(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				TiDBClusterComponent: spec.TiDBClusterComponent_TiKV,
				ConfigKey:            "split.qps-threshold",
				ConfigValue:          "1000",
			},
		},
	}
	err := SqlService.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_v2(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				InstanceAddr:         "127.0.0.1:10020",
				TiDBClusterComponent: spec.TiDBClusterComponent_TiKV,
				ConfigKey:            "split.qps-threshold",
				ConfigValue:          "3000",
			},
		},
	}
	err := SqlService.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_v3(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				TiDBClusterComponent: spec.TiDBClusterComponent_TiKV,
				ConfigKey:            "log-level",
				ConfigValue:          "'warn'",
			},
		},
	}
	err := SqlService.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_v4(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				TiDBClusterComponent: spec.TiDBClusterComponent_PD,
				ConfigKey:            "log.level",
				ConfigValue:          "'info'",
			},
		},
	}
	err := SqlService.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_v5(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				TiDBClusterComponent: spec.TiDBClusterComponent_TiDB,
				ConfigKey:            "tidb_slow_log_threshold",
				ConfigValue:          "200",
			},
		},
	}
	err := SqlService.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_v6(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				TiDBClusterComponent: spec.TiDBClusterComponent_TiFlash,
				ConfigKey:            "tidb_slow_log_threshold",
				ConfigValue:          "200",
			},
		},
	}
	err := SqlService.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil || !strings.Contains(err.Error(), "not support") {
		t.Errorf("err nil or err(%s) not contain not support", err.Error())
	}
}

func Test_execEditConfigThruSQL(t *testing.T) {
	sqlCommand := "set config tikv `split.qps-threshold` = 1000"
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec(sqlCommand).
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	err = execEditConfigThruSQL(context.TODO(), db, sqlCommand)
	if err == nil || !strings.Contains(err.Error(), "some error") {
		t.Errorf("err(%s) should contain 'some error'", err.Error())
	}
}

func Test_execShowWarningsThruSQL(t *testing.T) {
	sqlCommand := "show warnings"
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery(sqlCommand).
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	err = execShowWarningsThruSQL(context.TODO(), db)
	if err != nil {
		t.Errorf("err not nil")
	}
}
