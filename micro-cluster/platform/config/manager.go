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

package config

import (
	"context"
	"github.com/pingcap/tiunimanager/common/structs"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/platform/config"
)

type SystemConfigManager struct{}

func NewSystemConfigManager() *SystemConfigManager {
	return &SystemConfigManager{}
}

func (mgr *SystemConfigManager) GetSystemConfig(ctx context.Context, request message.GetSystemConfigReq) (resp message.GetSystemConfigResp, err error) {
	configRW := models.GetConfigReaderWriter()
	config, err := configRW.GetConfig(ctx, request.ConfigKey)
	if err != nil {
		return resp, err
	}
	resp.SystemConfig = structs.SystemConfig{
		ConfigKey:   config.ConfigKey,
		ConfigValue: config.ConfigValue,
	}
	return resp, nil
}

func (mgr *SystemConfigManager) UpdateSystemConfig(ctx context.Context, request message.UpdateSystemConfigReq) (resp message.UpdateSystemConfigResp, err error) {
	configRW := models.GetConfigReaderWriter()
	err = configRW.UpdateConfig(ctx, &config.SystemConfig{ConfigKey: request.ConfigKey, ConfigValue: request.ConfigValue})
	if err != nil {
		return resp, err
	}
	return resp, nil
}
