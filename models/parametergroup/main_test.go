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

/*******************************************************************************
 * @File: main_test.go
 * @Description: parameter group main test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/13 14:54
*******************************************************************************/

package parametergroup

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/pingcap-inc/tiem/models/cluster/management"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/util/uuidutil"

	"github.com/pingcap-inc/tiem/library/framework"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testRW *ParameterGroupReadWrite

func TestMain(m *testing.M) {
	testFilePath := "testdata/" + uuidutil.ShortId()
	os.MkdirAll(testFilePath, 0755)

	logins := framework.LogForkFile(constants.LogFileSystem)

	defer func() {
		os.RemoveAll(testFilePath)
		os.Remove(testFilePath)
	}()

	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			dbFile := testFilePath + constants.DBDirPrefix + constants.DatabaseFileName
			db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

			if err != nil || db.Error != nil {
				logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, db.Error)
			} else {
				logins.Infof("open database successful, filepath: %s", dbFile)
			}
			db.Migrator().CreateTable(Parameter{})
			db.Migrator().CreateTable(ParameterGroup{})
			db.Migrator().CreateTable(ParameterGroupMapping{})
			db.Migrator().CreateTable(management.Cluster{})

			testRW = NewParameterGroupReadWrite(db)
			return nil
		},
	)

	os.Exit(m.Run())
}

func buildParamGroup(count uint, params []*Parameter, addParams []message.ParameterInfo) (pgs []*ParameterGroup, err error) {
	pgs = make([]*ParameterGroup, count)
	for i := range pgs {
		pgs[i] = &ParameterGroup{
			Name:           "test_param_group_" + uuidutil.GenerateID(),
			ClusterSpec:    "8C16G",
			HasDefault:     1,
			DBType:         1,
			GroupType:      1,
			ClusterVersion: "5.0",
			Note:           "test param group " + strconv.Itoa(i),
		}
		pgm := make([]*ParameterGroupMapping, len(params))
		for j := range params {
			pgm[j] = &ParameterGroupMapping{
				ParameterID:  params[j].ID,
				DefaultValue: strconv.Itoa(j),
				Note:         "test param " + strconv.Itoa(j),
			}
		}
		pg, err := testRW.CreateParameterGroup(context.TODO(), pgs[i], pgm, addParams)
		if err != nil {
			return pgs, err
		}
		pgs[i].ID = pg.ID
	}
	return pgs, nil
}

func buildParams(count uint) (params []*Parameter, err error) {
	params = make([]*Parameter, count)
	for i := range params {
		params[i] = &Parameter{
			Category:     "basic",
			Name:         "test_param_" + uuidutil.GenerateID(),
			InstanceType: "TiKV",
			Type:         0,
			Unit:         "kb",
			Range:        "[\"0\", \"10\"]",
			HasReboot:    1,
			UpdateSource: 1,
			Description:  "test param name order " + strconv.Itoa(i),
		}
		parameter, err := testRW.CreateParameter(context.TODO(), params[i])
		if err != nil {
			return params, err
		}
		params[i].ID = parameter.ID
	}
	return params, nil
}
