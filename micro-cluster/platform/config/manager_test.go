/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
	"github.com/golang/mock/gomock"
	"github.com/pingcap/tiunimanager/message"
	"github.com/pingcap/tiunimanager/models"
	"github.com/pingcap/tiunimanager/models/platform/config"
	"github.com/pingcap/tiunimanager/test/mockmodels/mockconfig"
	"github.com/stretchr/testify/assert"
	"testing"
)

func init() {
	models.MockDB()
}

func TestSystemConfigManager_GetSystemConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	configRW := mockconfig.NewMockReaderWriter(ctrl)
	configRW.EXPECT().GetConfig(gomock.Any(), gomock.Any()).Return(&config.SystemConfig{ConfigKey: "key", ConfigValue: "value"}, nil).AnyTimes()
	models.SetConfigReaderWriter(configRW)

	mgr := NewSystemConfigManager()
	_, err := mgr.GetSystemConfig(context.TODO(), message.GetSystemConfigReq{
		ConfigKey: "key",
	})
	assert.Nil(t, err)
}
