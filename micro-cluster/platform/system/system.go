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
 * @File: system.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/16
*******************************************************************************/

package system

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/platform/system"
	"sync"
)

type SystemManager struct {
	actionBindings map[constants.SystemEvent]map[constants.SystemState]action
}

var manager *SystemManager
var once sync.Once

func GetSystemManager() *SystemManager {
	once.Do(func() {
		if manager == nil {
			manager = &SystemManager{}
			manager.actionBindings = map[constants.SystemEvent]map[constants.SystemState]action{
				constants.SystemProcessStarted: {
					constants.SystemInitialing:    pushToServiceReady,
					constants.SystemUpgrading:     pushToServiceReady,
					constants.SystemUnserviceable: pushToServiceReady,
					constants.SystemRunning:       pushToServiceReady,
					constants.SystemDataReady:     pushToServiceReady,
					constants.SystemFailure:       pushToServiceReady,
				},
				constants.SystemDataInitialized: {
					constants.SystemServiceReady: pushToDataReady,
				},
				constants.SystemServe: {
					constants.SystemDataReady:     pushToRunning,
					constants.SystemUnserviceable: pushToRunning,
					constants.SystemFailure:       pushToRunning,
				},
				constants.SystemStop: {
					constants.SystemRunning: pushToUnserviceable,
				},
				constants.SystemProcessUpgrade: {
					constants.SystemUnserviceable: pushToUpgrading,
				},
				constants.SystemFailureDetected: {
					constants.SystemRunning: pushToFailure,
				},
			}
		}
	})
	return manager
}

func getActionBindings() map[constants.SystemEvent]map[constants.SystemState]action {
	return GetSystemManager().actionBindings
}

func (p *SystemManager) AcceptSystemEvent(ctx context.Context, event constants.SystemEvent) error {
	if len(event) == 0 {
		return errors.NewError(errors.TIEM_PARAMETER_INVALID, "system event is empty")
	}
	return acceptSystemEvent(ctx, event)
}

func acceptSystemEvent(ctx context.Context, event constants.SystemEvent) error {
	if statusMapAction, ok := getActionBindings()[event]; ok {
		systemInfo, err := models.GetSystemReaderWriter().GetSystemInfo(context.TODO())
		if err != nil {
			return err
		}
		if actionFunc, statusOK := statusMapAction[systemInfo.State]; statusOK {
			return actionFunc(ctx, event, systemInfo.State)
		}
		return errors.NewErrorf(errors.TIEM_SYSTEM_STATE_CONFLICT, "undefined action for event %s in state %s", event, systemInfo.State)
	} else {
		return errors.NewErrorf(errors.TIEM_PARAMETER_INVALID, "system event %s is unrecognized", event)
	}
}

func (p *SystemManager) GetSystemInfo(ctx context.Context) (*system.SystemInfo, error) {
	return models.GetSystemReaderWriter().GetSystemInfo(ctx)
}

func (p *SystemManager) GetSystemVersionInfo(ctx context.Context) (*system.VersionInfo, error) {
	var systemInfo *system.SystemInfo
	var versionInfo *system.VersionInfo
	return versionInfo, errors.OfNullable(nil).
		BreakIf(func() error {
			got, err := p.GetSystemInfo(ctx)
			systemInfo = got
			return err
		}).
		BreakIf(func() error {
			got, err := models.GetSystemReaderWriter().GetVersion(ctx, systemInfo.CurrentVersionID)
			if err != nil {
				versionInfo = got
			}
			return err
		}).
		If(func(err error) {
			framework.LogWithContext(ctx).Errorf("get system version info failed, err = %s", err.Error())
		}).
		Else(func() {
			framework.LogWithContext(ctx).Infof("get system version info succeed, info = %v", *versionInfo)
		}).
		Present()
}
