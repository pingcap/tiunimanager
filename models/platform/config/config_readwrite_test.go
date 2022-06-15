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
	"github.com/stretchr/testify/assert"
	"testing"
)

var rw *ConfigReadWrite

func TestConfigReadWrite_GetConfig(t *testing.T) {
	config := &SystemConfig{
		ConfigKey:   "key1",
		ConfigValue: "value",
	}
	configCreate, errCreate := rw.CreateConfig(context.TODO(), config)
	assert.NoError(t, errCreate)

	configGet, errGet := rw.GetConfig(context.TODO(), configCreate.ConfigKey)
	assert.NoError(t, errGet)
	assert.Equal(t, configCreate.ConfigValue, configGet.ConfigValue)
}

func TestConfigReadWrite_UpdateConfig(t *testing.T) {
	config := &SystemConfig{
		ConfigKey:   "key2",
		ConfigValue: "value",
	}
	configCreate, errCreate := rw.CreateConfig(context.TODO(), config)
	assert.NoError(t, errCreate)

	config.ConfigValue = "value2"
	errUpdate := rw.UpdateConfig(context.TODO(), config)
	assert.NoError(t, errUpdate)

	configGet, errGet := rw.GetConfig(context.TODO(), configCreate.ConfigKey)
	assert.NoError(t, errGet)
	assert.Equal(t, config.ConfigValue, configGet.ConfigValue)
}
