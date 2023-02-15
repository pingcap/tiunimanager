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
	"sync"

	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/library/framework"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/models"
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
					constants.SystemServiceReady:  pushToServiceReady,
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
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "system event is empty")
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
		return errors.NewErrorf(errors.TIUNIMANAGER_SYSTEM_STATE_CONFLICT, "undefined action for event %s in state %s", event, systemInfo.State)
	} else {
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "system event %s is unrecognized", event)
	}
}

var SupportedProducts = []structs.ProductWithVersions{
	{
		ProductID:   string(constants.EMProductIDTiDB),
		ProductName: string(constants.EMProductIDTiDB),
		Versions: []structs.SpecificVersionProduct{
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.0.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.0.1"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.0.2"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.0.3"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.0.4"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.0.5"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.0.6"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.1.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.1.1"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.1.2"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.1.3"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.1.4"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.2.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.2.1"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.2.2"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.2.3"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.3.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.3.1"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v5.4.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v6.0.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v6.1.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchArm64), Version: "v6.1.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v6.2.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchArm64), Version: "v6.2.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v6.3.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchArm64), Version: "v6.3.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v6.4.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchArm64), Version: "v6.4.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchX8664), Version: "v6.5.0"},
			{ProductID: string(constants.EMProductIDTiDB), Arch: string(constants.ArchArm64), Version: "v6.5.0"},
		},
	},
}

var SupportedVendors = []structs.VendorInfo{
	{ID: string(constants.Local), Name: "local datacenter"},
}

func (p *SystemManager) GetSystemInfo(ctx context.Context, req message.GetSystemInfoReq) (resp *message.GetSystemInfoResp, err error) {
	systemInfo, err := models.GetSystemReaderWriter().GetSystemInfo(ctx)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("get system info failed, err = %s", err.Error())
		return nil, err
	} else {
		resp = &message.GetSystemInfoResp{
			Info: structs.SystemInfo{
				SystemName:                   systemInfo.SystemName,
				SystemLogo:                   systemInfo.SystemLogo,
				CurrentVersionID:             systemInfo.CurrentVersionID,
				LastVersionID:                systemInfo.LastVersionID,
				State:                        string(systemInfo.State),
				VendorZonesInitialized:       systemInfo.VendorZonesInitialized,
				VendorSpecsInitialized:       systemInfo.VendorSpecsInitialized,
				ProductComponentsInitialized: systemInfo.ProductComponentsInitialized,
				ProductVersionsInitialized:   systemInfo.ProductVersionsInitialized,
				SupportedProducts:            SupportedProducts,
				SupportedVendors:             SupportedVendors,
			},
		}
	}

	if req.WithVersionDetail && len(systemInfo.CurrentVersionID) > 0 {
		// current version
		if got, versionError := models.GetSystemReaderWriter().GetVersion(ctx, systemInfo.CurrentVersionID); versionError == nil {
			resp.CurrentVersion = structs.SystemVersionInfo{
				VersionID:   got.ID,
				Desc:        got.Desc,
				ReleaseNote: got.ReleaseNote,
			}
		} else {
			framework.LogWithContext(ctx).Errorf("get system current version failed, err = %s", versionError.Error())
			return nil, versionError
		}
	}

	if req.WithVersionDetail && len(systemInfo.LastVersionID) > 0 {
		// last version
		if got, versionError := models.GetSystemReaderWriter().GetVersion(ctx, systemInfo.LastVersionID); versionError == nil {
			resp.LastVersion = structs.SystemVersionInfo{
				VersionID:   got.ID,
				Desc:        got.Desc,
				ReleaseNote: got.ReleaseNote,
			}
		} else {
			framework.LogWithContext(ctx).Errorf("get system current version failed, err = %s", versionError.Error())
			return nil, versionError
		}
	}

	return resp, nil
}
