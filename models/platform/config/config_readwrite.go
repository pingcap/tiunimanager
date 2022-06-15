/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
	"github.com/pingcap/tiunimanager/common/errors"
	dbCommon "github.com/pingcap/tiunimanager/models/common"
	"gorm.io/gorm"
)

type ConfigReadWrite struct {
	dbCommon.GormDB
}

func NewConfigReadWrite(db *gorm.DB) *ConfigReadWrite {
	m := &ConfigReadWrite{
		dbCommon.WrapDB(db),
	}
	return m
}

func (m *ConfigReadWrite) CreateConfig(ctx context.Context, cfg *SystemConfig) (*SystemConfig, error) {
	return cfg, m.DB(ctx).Create(cfg).Error
}

func (m *ConfigReadWrite) GetConfig(ctx context.Context, configKey string) (config *SystemConfig, err error) {
	if "" == configKey {
		return nil, errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "config key required")
	}
	config = &SystemConfig{}
	err = m.DB(ctx).First(config, "config_key = ?", configKey).Error
	if err != nil {
		return nil, err
	}
	return config, err
}

func (m *ConfigReadWrite) UpdateConfig(ctx context.Context, config *SystemConfig) (err error) {
	if "" == config.ConfigKey {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, "config key required")
	}
	cfg := &SystemConfig{}
	err = m.DB(ctx).First(cfg, "config_key = ?", config.ConfigKey).Error
	if err != nil {
		return errors.NewErrorf(errors.TIUNIMANAGER_PARAMETER_INVALID, "config key %s not exist", config.ConfigKey)
	}
	return m.DB(ctx).Model(cfg).Update("config_value", config.ConfigValue).Error
}
