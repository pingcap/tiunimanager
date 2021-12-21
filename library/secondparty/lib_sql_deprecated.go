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
 * @File: lib_sql
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/6
*******************************************************************************/

package secondparty

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pingcap-inc/tiem/common/constants"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pingcap-inc/tiem/library/framework"
)

// Deprecated
func (secondMicro *SecondMicro) EditClusterConfig(ctx context.Context, req ClusterEditConfigReq, bizID uint64) error {
	logInFunc := framework.LogWithContext(ctx).WithField("bizid", bizID)
	logInFunc.Infof("editclusterconfig, clustereditconfigreq: %v, bizid: %d", req, bizID)

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", req.DbConnParameter.Username, req.DbConnParameter.Password, req.DbConnParameter.IP, req.DbConnParameter.Port))
	if err != nil {
		logInFunc.Error("conn tidb error", err)
		return err
	}
	defer db.Close()

	for _, config := range req.ComponentConfigs {
		var args []string
		if config.TiDBClusterComponent == constants.ComponentIDTiDB {
			args = append(args, "set")
		} else if config.TiDBClusterComponent == constants.ComponentIDTiKV || config.TiDBClusterComponent == constants.ComponentIDPD {
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
