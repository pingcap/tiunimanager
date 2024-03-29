/******************************************************************************
 * Copyright (c)  2023 PingCAP, Inc.                                          *
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

package dbagent

import (
	"os"
	"testing"

	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/models"
)

func TestMain(m *testing.M) {
	var testFilePath string
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			models.MockDB()
			testFilePath = d.GetDataDir()
			os.MkdirAll(testFilePath, 0755)
			models.MockDB()
			return models.Open(d)
		},
	)
	code := m.Run()
	os.RemoveAll(testFilePath)

	os.Exit(code)
}
