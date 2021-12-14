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
 * @Description: cluster parameter main test
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/13 16:54
*******************************************************************************/

package parameter

import (
	"context"
	"os"
	"strconv"
	"testing"

	common2 "github.com/pingcap-inc/tiem/models/common"

	"github.com/pingcap-inc/tiem/models/cluster/management"
	"github.com/pingcap-inc/tiem/models/parametergroup"

	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/util/uuidutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var clusterParameterRW *ClusterParameterReadWrite
var parameterGroupRW *parametergroup.ParameterGroupReadWrite
var clusterRW *management.GormClusterReadWrite

func TestMain(m *testing.M) {
	testFilePath := "testdata/" + uuidutil.ShortId()
	os.MkdirAll(testFilePath, 0755)

	logins := framework.LogForkFile(common.LogFileSystem)

	defer func() {
		os.RemoveAll(testFilePath)
		os.Remove(testFilePath)
	}()

	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			dbFile := testFilePath + common.DBDirPrefix + common.DatabaseFileName
			db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

			if err != nil || db.Error != nil {
				logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, db.Error)
			} else {
				logins.Infof("open database successful, filepath: %s", dbFile)
			}
			db.Migrator().CreateTable(parametergroup.Parameter{})
			db.Migrator().CreateTable(parametergroup.ParameterGroup{})
			db.Migrator().CreateTable(parametergroup.ParameterGroupMapping{})

			db.Migrator().CreateTable(management.Cluster{})
			db.Migrator().CreateTable(ClusterParameterMapping{})

			clusterParameterRW = NewClusterParameterReadWrite(db)
			parameterGroupRW = parametergroup.NewParameterGroupReadWrite(db)
			clusterRW = management.NewGormClusterReadWrite(db)
			return nil
		},
	)

	os.Exit(m.Run())
}

func buildParamGroup(count uint, params []*parametergroup.Parameter) (pgs []*parametergroup.ParameterGroup, err error) {
	pgs = make([]*parametergroup.ParameterGroup, count)
	for i := range pgs {
		pgs[i] = &parametergroup.ParameterGroup{
			Name:           "test_param_group_" + uuidutil.GenerateID(),
			ClusterSpec:    "8C16G",
			HasDefault:     1,
			DBType:         1,
			GroupType:      1,
			ClusterVersion: "5.0",
			Note:           "test param group " + strconv.Itoa(i),
		}
		pgm := make([]*parametergroup.ParameterGroupMapping, len(params))
		for j := range params {
			pgm[j] = &parametergroup.ParameterGroupMapping{
				ParameterID:  params[j].ID,
				DefaultValue: strconv.Itoa(j),
				Note:         "test param " + strconv.Itoa(j),
			}
		}
		pg, err := parameterGroupRW.CreateParameterGroup(context.TODO(), pgs[i], pgm)
		if err != nil {
			return pgs, err
		}
		pgs[i].ID = pg.ID
	}
	return pgs, nil
}

func buildParams(count uint) (params []*parametergroup.Parameter, err error) {
	params = make([]*parametergroup.Parameter, count)
	for i := range params {
		params[i] = &parametergroup.Parameter{
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
		parameter, err := parameterGroupRW.CreateParameter(context.TODO(), params[i])
		if err != nil {
			return params, err
		}
		params[i].ID = parameter.ID
	}
	return params, nil
}

func buildCluster() (*management.Cluster, error) {
	cluster, err := clusterRW.Create(context.TODO(), &management.Cluster{
		Entity: common2.Entity{
			TenantId: "111",
		},
		Name:    "build_test_cluster_" + uuidutil.GenerateID(),
		OwnerId: "111",
	})
	if err != nil {
		return nil, err
	}
	return cluster, nil
}
