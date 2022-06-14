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
 * @File: initdata_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/18
*******************************************************************************/

package models

import (
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func Test_allVersionInitializers(t *testing.T) {
	// open empty
	err := Open(framework.Current.(*framework.BaseFramework))
	defer func() {
		defaultDb = nil
		os.RemoveAll(framework.Current.(*framework.BaseFramework).GetDataDir() + constants.DBDirPrefix + constants.DatabaseFileName)
	}()

	assert.NoError(t, err)
	err = IncrementVersionData("", inTestingVersion)
	assert.NoError(t, err)
	// todo add assertion for each new version here
}

func Test_initBySql(t *testing.T) {
	err := Open(framework.Current.(*framework.BaseFramework))
	file := framework.Current.(*framework.BaseFramework).GetDataDir() + constants.DBDirPrefix + "test.sql"

	defer func() {
		defaultDb = nil
		os.RemoveAll(framework.Current.(*framework.BaseFramework).GetDataDir() + constants.DBDirPrefix + constants.DatabaseFileName)
		os.RemoveAll(file)
	}()

	assert.NoError(t, err)
	ioutil.WriteFile(file, []byte("select * from system_infos;"), 0600)
	err = initBySql(nil, file, "test")
	assert.NoError(t, err)
}
