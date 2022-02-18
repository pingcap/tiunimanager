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
 * @File: statemachine_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/18
*******************************************************************************/

package system

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_StateMachine(t *testing.T) {
	// process started
	err := GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemProcessStarted)
	assert.NoError(t, err)
	systemInfo, err := models.GetSystemReaderWriter().GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, constants.SystemRunning, systemInfo.State)

	// wrong event
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemServe)
	assert.Error(t, err)
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemDataInitialized)
	assert.Error(t, err)
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemProcessUpgrade)
	assert.Error(t, err)

	// start again
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemProcessStarted)
	assert.NoError(t, err)

	systemInfo, err = models.GetSystemReaderWriter().GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, constants.SystemRunning, systemInfo.State)

	// stop
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemStop)
	assert.NoError(t, err)
	systemInfo, err = models.GetSystemReaderWriter().GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, constants.SystemUnserviceable, systemInfo.State)

	// wrong event
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemDataInitialized)
	assert.Error(t, err)

	// serve
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemServe)
	assert.NoError(t, err)
	systemInfo, err = models.GetSystemReaderWriter().GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, constants.SystemRunning, systemInfo.State)

	// stop
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemStop)
	assert.NoError(t, err)
	// upgrade
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemProcessUpgrade)
	assert.NoError(t, err)

	// wrong event
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemServe)
	assert.Error(t, err)
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemStop)
	assert.Error(t, err)
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemDataInitialized)
	assert.Error(t, err)

	//  process started
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemProcessStarted)
	assert.NoError(t, err)

	// failure
	err = GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemFailureDetected)
	assert.NoError(t, err)
	systemInfo, err = models.GetSystemReaderWriter().GetSystemInfo(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, constants.SystemFailure, systemInfo.State)

	// serve
	GetSystemManager().AcceptSystemEvent(context.TODO(), constants.SystemServe)
}

func Test_pushToServiceReady(t *testing.T) {

}

func Test_pushToDataReady(t *testing.T) {

}

func Test_pushToRunning(t *testing.T) {

}

func Test_pushToUnFailure(t *testing.T) {

}

func Test_pushToUnserviceable(t *testing.T) {

}

func Test_pushToUpgrading(t *testing.T) {

}
