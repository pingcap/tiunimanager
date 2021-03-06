/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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
 * @File: clusterconfig.go
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/15
*******************************************************************************/

package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/library/spec"
)

type ClusterEditConfigReq struct {
	DbConnParameter  DbConnParam
	ComponentConfigs []ClusterComponentConfig
}

type ClusterComponentConfig struct {
	TiDBClusterComponent spec.TiDBClusterComponent
	InstanceAddr         string
	ConfigKey            string
	ConfigValue          string
}

type ShowWarningsResp struct {
	Level   string
	Code    string
	Message string
}

var SqlService ClusterConfigService

type ClusterConfigService interface {
	EditClusterConfig(ctx context.Context, req ClusterEditConfigReq, workFlowNodeID string) error
}

type ClusterConfigServiceImpl struct{}

func init() {
	SqlService = new(ClusterConfigServiceImpl)
}

func (service *ClusterConfigServiceImpl) EditClusterConfig(ctx context.Context, req ClusterEditConfigReq, workFlowNodeID string) error {
	logInFunc := framework.LogWithContext(ctx).WithField("workflownodeid", workFlowNodeID)
	logInFunc.Infof("editclusterconfig, clustereditconfigreq: %v", req)

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", req.DbConnParameter.Username, req.DbConnParameter.Password, req.DbConnParameter.IP, req.DbConnParameter.Port))
	if err != nil {
		logInFunc.Error("conn tidb error", err)
		return err
	}
	defer db.Close()

	for _, config := range req.ComponentConfigs {
		var args []string
		if config.TiDBClusterComponent == spec.TiDBClusterComponent_TiDB {
			args = append(args, "set")
		} else if config.TiDBClusterComponent == spec.TiDBClusterComponent_TiKV || config.TiDBClusterComponent == spec.TiDBClusterComponent_PD {
			args = append(args, "set config")
			if len(config.InstanceAddr) != 0 {
				args = append(args, fmt.Sprintf("\"%s\"", config.InstanceAddr))
			} else {
				args = append(args, string(config.TiDBClusterComponent))
			}
		} else {
			return fmt.Errorf("not support %s", string(config.TiDBClusterComponent))
		}

		args = append(args, fmt.Sprintf("`%s` = %s", config.ConfigKey, config.ConfigValue))
		sqlCommand := strings.Join(args, " ")

		logInFunc.Infof("task start processing: %s, on: %s:%s", sqlCommand, req.DbConnParameter.IP, req.DbConnParameter.Port)

		err = execEditConfigThruSQL(ctx, db, sqlCommand)
		if err != nil {
			return err
		}
		err = execShowWarningsThruSQL(ctx, db)
		if err != nil {
			return err
		}
	}
	return nil
}

func execEditConfigThruSQL(ctx context.Context, db *sql.DB, sqlCommand string) error {
	logInFunc := framework.LogWithContext(ctx)
	_, err := db.Exec(sqlCommand)
	if err != nil {
		logInFunc.Errorf("set config error: %s", err.Error())
		return err
	}
	return nil
}

func execShowWarningsThruSQL(ctx context.Context, db *sql.DB) error {
	logInFunc := framework.LogWithContext(ctx)
	var showWarningsResp ShowWarningsResp
	err := db.QueryRow("show warnings").Scan(&showWarningsResp.Level, &showWarningsResp.Code, &showWarningsResp.Message)
	if err != nil {
		logInFunc.Infof("task finished as show warnings error: %s", err.Error())
		return nil
	}

	logInFunc.Errorf("set config error: %s", showWarningsResp.Message)
	return errors.New(showWarningsResp.Message)
}
