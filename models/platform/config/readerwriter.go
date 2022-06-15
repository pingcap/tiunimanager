/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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

package config

import "context"

type ReaderWriter interface {
	// CreateConfig
	// @Description: create system config
	// @Receiver m
	// @Parameter ctx
	// @Parameter cfg
	// @Return *SystemConfig
	// @Return error
	CreateConfig(ctx context.Context, cfg *SystemConfig) (*SystemConfig, error)

	// GetConfig
	// @Description: get system config by configKey
	// @Receiver m
	// @Parameter ctx
	// @Parameter configKey
	// @Return *SystemConfig
	// @Return error
	GetConfig(ctx context.Context, configKey string) (config *SystemConfig, err error)

	// UpdateConfig
	// @Description: update system config
	// @Receiver m
	// @Parameter ctx
	// @Parameter configKey
	// @Return *SystemConfig
	// @Return error
	UpdateConfig(ctx context.Context, config *SystemConfig) (err error)
}
