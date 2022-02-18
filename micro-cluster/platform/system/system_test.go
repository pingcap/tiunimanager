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
 * @File: system_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/18
*******************************************************************************/

package system

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSystemManager_AcceptSystemEvent(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		err := GetSystemManager().AcceptSystemEvent(context.TODO(), "")
		assert.Error(t, err)
	})
	t.Run("unknown", func(t *testing.T) {
		err := GetSystemManager().AcceptSystemEvent(context.TODO(), "unknown")
		assert.Error(t, err)
	})
}

func TestSystemManager_GetSystemInfo(t *testing.T) {
	systemInfo, err := GetSystemManager().GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.NotEmpty(t, systemInfo.SystemName)
	assert.NotEmpty(t, systemInfo.State)
}

func TestSystemManager_GetSystemVersionInfo(t *testing.T) {
	_, err := GetSystemManager().GetSystemVersionInfo(context.TODO())
	assert.Error(t, err)

}
