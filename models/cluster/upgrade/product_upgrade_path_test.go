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
 * @File: productupgradepath_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/5
*******************************************************************************/

package upgrade

import (
	"context"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"
)

func TestGormProductUpgradePathReadWrite_Create_Fail(t *testing.T) {
	path, err := testRW.Create(context.TODO(), "", constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	if err == nil {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}

	path, err = testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, "", "v4.0.12", "v5.0.0")
	if err == nil {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}

	path, err = testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "", "v5.0.0")
	if err == nil {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}

	path, err = testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "")
	if err == nil {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}
}

func TestGormProductUpgradePathReadWrite_Create_Success(t *testing.T) {
	path, err := testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	defer testRW.Delete(context.TODO(), path.ID)
	if err != nil || path.ID == "" {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}

	// idempotent create
	path, err = testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	if err != nil || path.ID == "" {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}
}

func TestGormProductUpgradePathReadWrite_queryByPathParam_Fail(t *testing.T) {
	path, err := testRW.queryByPathParam(context.TODO(), "", constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	if err == nil {
		t.Errorf("queryByPathParam() error: %v: path: %v", err, path)
	}

	path, err = testRW.queryByPathParam(context.TODO(), constants.UpgradeTypeInPlace, "", "v4.0.12", "v5.0.0")
	if err == nil {
		t.Errorf("queryByPathParam() error: %v: path: %v", err, path)
	}

	path, err = testRW.queryByPathParam(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "", "v5.0.0")
	if err == nil {
		t.Errorf("queryByPathParam() error: %v: path: %v", err, path)
	}

	path, err = testRW.queryByPathParam(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "")
	if err == nil {
		t.Errorf("queryByPathParam() error: %v: path: %v", err, path)
	}
}

func TestGormProductUpgradePathReadWrite_queryByPathParam_Success(t *testing.T) {
	path, err := testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	defer testRW.Delete(context.TODO(), path.ID)
	if err != nil || path.ID == "" {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}

	path, err = testRW.queryByPathParam(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	if err != nil || path.ID == "" {
		t.Errorf("queryByPathParam() error: %v: path: %v", err, path)
	}
}

func TestGormProductUpgradePathReadWrite_Get_Fail1(t *testing.T) {
	path, err := testRW.Get(context.TODO(), "")
	if err == nil {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}
}

func TestGormProductUpgradePathReadWrite_Get_Fail2(t *testing.T) {
	path, err := testRW.Get(context.TODO(), "nosuchid")
	if err == nil {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}
}

func TestGormProductUpgradePathReadWrite_Get_Success(t *testing.T) {
	path, err := testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	defer testRW.Delete(context.TODO(), path.ID)
	if err != nil || path.ID == "" {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}

	path, err = testRW.Get(context.TODO(), path.ID)
	if err != nil || path.ID == "" {
		t.Errorf("Get() error: %v, path: %v", err, path)
	}
}

func TestGormProductUpgradePathReadWrite_QueryBySrcVersion_Fail1(t *testing.T) {
	paths, err := testRW.QueryBySrcVersion(context.TODO(), "")
	if err == nil {
		t.Errorf("Create() error: %v: paths: %v", err, paths)
	}
}

func TestGormProductUpgradePathReadWrite_QueryBySrcVersion_Fail2(t *testing.T) {
	paths, err := testRW.QueryBySrcVersion(context.TODO(), "nosuchversion")
	if len(paths) != 0 {
		t.Errorf("Create() error: %v: paths: %v", err, paths)
	}
}

func TestGormProductUpgradePathReadWrite_QueryBySrcVersion_Fail3(t *testing.T) {
	paths, err := testRW.QueryBySrcVersion(context.TODO(), "v4.0.12")
	if len(paths) != 0 {
		t.Errorf("Create() error: %v: paths: %v", err, paths)
	}
}

func TestGormProductUpgradePathReadWrite_QueryBySrcVersion_Success(t *testing.T) {
	path, err := testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	defer testRW.Delete(context.TODO(), path.ID)
	if err != nil || path.ID == "" {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}
	paths, err := testRW.QueryBySrcVersion(context.TODO(), "v4.0.12")
	if err != nil || paths == nil || len(paths) != 1 {
		t.Errorf("Create() error: %v: paths: %v", err, paths)
	}

	path, err = testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	if err != nil || path.ID == "" {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}
	paths, err = testRW.QueryBySrcVersion(context.TODO(), "v4.0.12")
	if err != nil || paths == nil || len(paths) != 1 {
		t.Errorf("Create() error: %v: paths: %v", err, paths)
	}

	path, err = testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "v5.3.0")
	if err != nil || path.ID == "" {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}
	paths, err = testRW.QueryBySrcVersion(context.TODO(), "v4.0.12")
	if err != nil || paths == nil || len(paths) != 2 {
		t.Errorf("Create() error: %v: paths: %v", err, paths)
	}
}

func TestGormProductUpgradePathReadWrite_Delete_Fail(t *testing.T) {
	err := testRW.Delete(context.TODO(), "")
	if err == nil {
		t.Errorf("Create() error: %v", err)
	}
}

func TestGormProductUpgradePathReadWrite_Delete_Success(t *testing.T) {
	path, err := testRW.Create(context.TODO(), constants.UpgradeTypeInPlace, constants.EMProductIDTiDB, "v4.0.12", "v5.0.0")
	if err != nil || path.ID == "" {
		t.Errorf("Create() error: %v: path: %v", err, path)
	}
	err = testRW.Delete(context.TODO(), path.ID)
	if err != nil {
		t.Errorf("Delete() error: %v", err)
	}
}
