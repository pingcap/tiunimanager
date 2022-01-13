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
 ******************************************************************************/

package switchover

import (
	"context"
	"fmt"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var databaseName = constants.SwitchoverReadWriteHealthTestDBName
var tableNamePrefix = constants.SwitchoverReadWriteHealthTestTableNamePrefix

type CheckWritable struct {
	gorm.Model

	// private field which is ignored by gorm
	tableNameSuffix string `gorm:"-"`
}

func (c CheckWritable) TableName() string {
	return fmt.Sprintf("%s_%s", tableNamePrefix, c.tableNameSuffix)
}

// set cluster Read-Writeable to normal user and changeFeedTask's user
func (p *Manager) clusterRestrictedReadOnlyOp(ctx context.Context, clusterID, addr, op string, setReadOnlyFlag bool) (readOnlyFlag bool, err error) {
	funcName := "clusterRestrictedReadOnlyOp"
	framework.LogWithContext(ctx).Infof("start %s", funcName)
	defer framework.LogWithContext(ctx).Infof("exit %s", funcName)
	if op == "set" || op == "get" {
	} else {
		framework.LogWithContext(ctx).Fatalf(
			"%s clusterRestrictedReadOnlyOp clusterID:%s unknow op:%s", funcName, clusterID, op)
	}
	name, pwd, err := mgr.clusterGetCDCUserNameAndPwd(ctx, clusterID)
	if err != nil {
		framework.LogWithContext(ctx).Warnf(
			"%s clusterGetCDCUserNameAndPwd clusterID:%s err:%s", funcName, clusterID, err)
		return readOnlyFlag, err
	} else {
		framework.LogWithContext(ctx).Debugf(
			"%s clusterGetCDCUserNameAndPwd clusterID:%s success", funcName, clusterID)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8mb4&parseTime=True&loc=Local", name, pwd, addr)
	mylog := framework.LogWithContext(ctx).WithField("clusterID", clusterID)
	mylog.Info("checkClusterWritable gorm.Open with dsn:", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		mylog.Warnf("%s gorm.Open err:", funcName, err)
		return readOnlyFlag, err
	} else {
		mylog.Debugf("%s gorm.Open success", funcName)
	}
	if op == "set" {
		value := 0
		if setReadOnlyFlag {
			value = 1
		}
		err = db.Exec(fmt.Sprintf("SET GLOBAL tidb_restricted_read_only = %d ;", value)).Error
		if err != nil {
			return false, err
		} else {
			return setReadOnlyFlag, err
		}
	} else {

	}
	return readOnlyFlag, nil
}

func (p *Manager) checkClusterWritable(ctx context.Context, clusterID, userName, pwd, addr string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", userName, pwd, addr, databaseName)
	mylog := framework.LogWithContext(ctx).WithField("clusterID", clusterID)
	mylog.Info("checkClusterWritable gorm.Open with dsn:", dsn)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		mylog.Warn("checkClusterWritable gorm.Open err:", err)
		return err
	} else {
		mylog.Info("checkClusterWritable gorm.Open success")
	}
	suffixOfTableName := uuidutil.GenerateID()
	var table = &CheckWritable{tableNameSuffix: suffixOfTableName}
	// Migrate the schema
	err = db.AutoMigrate(table)
	if err != nil {
		mylog.Warnf("checkClusterWritable AutoMigrate table %s err:%s", table.TableName(), err)
		return err
	} else {
		mylog.Info("checkClusterWritable AutoMigrate success")
		defer func() {
			err := db.Migrator().DropTable(table)
			if err != nil {
				mylog.Warnf("checkClusterWritable DropTable %s err:%s", table.TableName(), err)
			} else {
				mylog.Infof("checkClusterWritable DropTable %s success", table.TableName())
			}
		}()
	}
	ret := db.Create(&CheckWritable{})
	if ret.Error != nil {
		mylog.Warn("checkClusterWritable Create Record err:", ret.Error)
		return ret.Error
	} else {
		mylog.Info("checkClusterWritable Create Record success")
	}
	return nil
}
