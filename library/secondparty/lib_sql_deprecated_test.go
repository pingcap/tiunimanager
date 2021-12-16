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
 * @File: lib_sql_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/6
*******************************************************************************/

package secondparty

import (
	"context"
	"strings"
	"testing"

	"github.com/pingcap-inc/tiem/library/spec"
)

var secondMicro3 *SecondMicro

func init() {
	secondMicro3 = &SecondMicro{}
}

func TestSecondMicro_EditClusterConfig_Deprecated_v1(t *testing.T) {
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
	err := secondMicro3.EditClusterConfig(context.TODO(), req, 0)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_Deprecated_v2(t *testing.T) {
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
	err := secondMicro3.EditClusterConfig(context.TODO(), req, 0)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_Deprecated_v3(t *testing.T) {
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
	err := secondMicro3.EditClusterConfig(context.TODO(), req, 0)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_Deprecated_v4(t *testing.T) {
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
	err := secondMicro3.EditClusterConfig(context.TODO(), req, 0)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_Deprecated_v5(t *testing.T) {
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
	err := secondMicro3.EditClusterConfig(context.TODO(), req, 0)
	if err == nil {
		t.Error("err nil")
	}
}

func TestSecondMicro_EditClusterConfig_Deprecated_v6(t *testing.T) {
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
	err := secondMicro3.EditClusterConfig(context.TODO(), req, 0)
	if err == nil || !strings.Contains(err.Error(), "not support") {
		t.Errorf("err nil or err(%s) not contain not support", err.Error())
	}
}
