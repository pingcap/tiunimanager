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
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: main_test
 * @Description:
 * @Author: shenhaibo@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/18
*******************************************************************************/

package deployment

import (
	"fmt"
	"os"
	"testing"

	"github.com/pingcap/tiunimanager/util/uuidutil"
)

const (
	TestWorkFlowID         = "testworkflowid"
	TestClusterID          = "testclusterid"
	TestOperation          = "TIUP_HOME=/home/tidb/.tiup tiup cluster start testclusterid --wait-timeout 360 --yes"
	TestVersion            = "v4.0.12"
	TestDstVersion         = "v5.2.2"
	TestTiDBTopo           = ""
	TestTiDBScaleOutTopo   = ""
	TestTiDBCheckTopo      = ""
	TestTiDBEditConfigTopo = ""
	TestTiDBPushTopo       = ""
	TestNodeID             = ""

	TestResult   = "testresult"
	TestErrorStr = "testerrorstr"
)

var testTiUPHome string

func TestMain(m *testing.M) {
	testTiUPHome = "testdata/" + uuidutil.ShortId()
	os.MkdirAll(fmt.Sprintf("%s/storage", testTiUPHome), 0755)
	code := m.Run()
	os.RemoveAll("testdata/")
	os.RemoveAll("logs/")
	os.Exit(code)
}
