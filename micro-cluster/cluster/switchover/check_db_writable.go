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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var databaseName = constants.SwitchoverExclusiveDBNameForReadWriteHealthTest

type CheckWritable struct {
	gorm.Model
}

func (p *Manager) checkClusterWritable(ctx context.Context, userName, pwd, addr string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", userName, pwd, addr, databaseName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		framework.LogWithContext(ctx).Error("checkClusterWritable gorm.Open err:", err)
		return err
	} else {
		framework.LogWithContext(ctx).Info("checkClusterWritable gorm.Open success")
	}

	// Migrate the schema
	err = db.AutoMigrate(&CheckWritable{})
	if err != nil {
		framework.LogWithContext(ctx).Error("checkClusterWritable AutoMigrate err:", err)
		return err
	} else {
		framework.LogWithContext(ctx).Info("checkClusterWritable AutoMigrate success")
	}

	ret := db.Create(&CheckWritable{})
	if ret.Error != nil {
		framework.LogWithContext(ctx).Error("checkClusterWritable Create Record err:", ret.Error)
		return err
	} else {
		framework.LogWithContext(ctx).Info("checkClusterWritable Create Record success")
	}

	return nil
}
