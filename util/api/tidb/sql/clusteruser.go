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
 * @File: clusteruser.go
 * @Description:
 * @Author: xieyujie@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/21 14:31
*******************************************************************************/

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models/cluster/management"
)

func ExecCommandThruSQL(ctx context.Context, db *sql.DB, sqlCommand string) error {
	logInFunc := framework.LogWithContext(ctx)
	_, err := db.Exec(sqlCommand)
	if err != nil {
		logInFunc.Errorf("execute sql command %s error: %s", sqlCommand, err.Error())
		return err
	}
	return nil
}

func CreateDBUser(ctx context.Context, connec DbConnParam, user *management.DBUser, workFlowNodeID string) error {
	logInFunc := framework.LogWithContext(ctx).WithField("bizid", workFlowNodeID)
	logInFunc.Infof("createDBUser, user: %v, bizId: %s", user, workFlowNodeID)

	// connect database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", connec.Username, connec.Password, connec.IP, connec.Port))
	if err != nil {
		logInFunc.Error("conn tidb error", err)
		return err
	}
	defer db.Close()

	// execute sql command of creating user
	createSqlCommand := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'", user.Name, "%", user.Password.Val)
	err = ExecCommandThruSQL(ctx, db, createSqlCommand)
	if err != nil {
		return err
	}

	//	execute sql command of granting privileges to user
	for _, permission := range constants.DBUserPermission[constants.DBUserRoleType(user.RoleType)] {
		grantSqlCommand := fmt.Sprintf("GRANT %s ON %s.%s TO '%s'@'%s' IDENTIFIED BY '%s'",
			permission, "*", "*", user.Name, "%", user.Password.Val)
		err = ExecCommandThruSQL(ctx, db, grantSqlCommand)
		if err != nil {
			return err
		}
	}

	// save
	flushCommand := "FLUSH PRIVILEGES"
	err = ExecCommandThruSQL(ctx, db, flushCommand)
	if err != nil {
		return err
	}
	return nil
}

func UpdateDBUserPassword(ctx context.Context, connec DbConnParam, name string, password string, workFlowNodeID string) error {
	logInFunc := framework.LogWithContext(ctx).WithField("bizid", workFlowNodeID)
	logInFunc.Infof("UpdateDBUserPassword, name: %v, bizId: %s", name, workFlowNodeID)

	// connect database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", connec.Username, connec.Password, connec.IP, connec.Port))
	if err != nil {
		logInFunc.Error("conn tidb error", err)
		return err
	}
	defer db.Close()

	//execute sql command
	sqlCommand := fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'", name, "%", password)
	err = ExecCommandThruSQL(ctx, db, sqlCommand)

	if err != nil {
		return err
	}

	return nil
}

func DeleteDBUser(ctx context.Context, connec DbConnParam, name string, workFlowNodeID string) error {
	logInFunc := framework.LogWithContext(ctx).WithField("bizid", workFlowNodeID)
	logInFunc.Infof("DeleteDBUser, name: %v, bizId: %s", name, workFlowNodeID)

	// connect database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", connec.Username, connec.Password, connec.IP, connec.Port))
	if err != nil {
		logInFunc.Error("conn tidb error", err)
		return err
	}
	defer db.Close()

	//execute sql command
	sqlCommand := fmt.Sprintf("DROP USER '%s'@'%s'", name, "%")
	err = ExecCommandThruSQL(ctx, db, sqlCommand)

	if err != nil {
		return err
	}
	return nil
}
