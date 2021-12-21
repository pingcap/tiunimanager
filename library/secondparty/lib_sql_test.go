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

/*******************************************************************************
 * @File: lib_sql_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/15
*******************************************************************************/

package secondparty

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"

	"github.com/DATA-DOG/go-sqlmock"
)

var secondPartyManager3 *SecondPartyManager
var dbConnParam3 DbConnParam
var req ClusterEditConfigReq
var manager *SecondPartyManager

func init() {
	secondPartyManager3 = &SecondPartyManager{}
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
				TiDBClusterComponent: constants.ComponentIDTiKV,
				ConfigKey:            "split.qps-threshold",
				ConfigValue:          "1000",
			},
		},
	}
	err := secondPartyManager3.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
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
				TiDBClusterComponent: constants.ComponentIDTiKV,
				ConfigKey:            "split.qps-threshold",
				ConfigValue:          "3000",
			},
		},
	}
	err := secondPartyManager3.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_v3(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				TiDBClusterComponent: constants.ComponentIDTiKV,
				ConfigKey:            "log-level",
				ConfigValue:          "'warn'",
			},
		},
	}
	err := secondPartyManager3.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_v4(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				TiDBClusterComponent: constants.ComponentIDPD,
				ConfigKey:            "log.level",
				ConfigValue:          "'info'",
			},
		},
	}
	err := secondPartyManager3.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_v5(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				TiDBClusterComponent: constants.ComponentIDTiDB,
				ConfigKey:            "tidb_slow_log_threshold",
				ConfigValue:          "200",
			},
		},
	}
	err := secondPartyManager3.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_v6(t *testing.T) {
	req = ClusterEditConfigReq{
		DbConnParameter: dbConnParam3,
		ComponentConfigs: []ClusterComponentConfig{
			{
				TiDBClusterComponent: constants.ComponentIDTiFlash,
				ConfigKey:            "tidb_slow_log_threshold",
				ConfigValue:          "200",
			},
		},
	}
	err := secondPartyManager3.EditClusterConfig(context.TODO(), req, TestWorkFlowNodeID)
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

func TestSecondMicro_SetClusterDbPassword_v1(t *testing.T) {
	//invalid password length
	req := ClusterSetDbPswReq{
		DbConnParameter: DbConnParam{
			Username: "root",
			Password: "121345",
			IP:       "127.0.0.1",
			Port:     "4321",
		},
	}
	err := manager.SetClusterDbPassword(context.TODO(), req, "11")
	fmt.Println(err)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_SetClusterDbPassword_v2(t *testing.T) {
	//normal
	req := ClusterSetDbPswReq{
		DbConnParameter: DbConnParam{
			Username: "root",
			Password: "12345678",
			IP:       "127.0.0.1",
			Port:     "4321",
		},
	}
	err := manager.SetClusterDbPassword(context.TODO(), req, "22")
	fmt.Println(err)
	if err == nil {
		t.Error("err nil")
	}
}
