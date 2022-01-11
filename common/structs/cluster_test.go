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

package structs

import (
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_validateCreateClusterParameter(t *testing.T)  {
	parameter := CreateClusterParameter{
		Name: "ValidName",
		DBUser: "ValidUser",
		DBPassword: "ValidPassword",
		Type: "TiDB",
		Version: "v5.2.2",
		Region: "region1",
		CpuArchitecture: "X86",
	}
	t.Run("valid", func(t *testing.T) {
		err := validator.New().Struct(parameter)
		assert.NoError(t, err)
	})
	t.Run("name", func(t *testing.T) {
		parameter.Name = "len"
		err := validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.Name = "ValidString☢"
		err = validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.Name = "TooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLong"
		err = validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.Name = "ValidString_"
		err = validator.New().Struct(parameter)
		assert.NoError(t, err)

		parameter.Name = "ValidString"
	})
	t.Run("user", func(t *testing.T) {
		parameter.DBUser = ""
		err := validator.New().Struct(parameter)
		assert.NoError(t, err)

		parameter.DBUser = "TooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLong"
		err = validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.DBUser = "ValidString"
	})
	t.Run("password", func(t *testing.T) {
		parameter.DBPassword = "len"
		err := validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.DBPassword = "ValidString☢"
		err = validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.DBPassword = "TooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLong"
		err = validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.DBPassword = "ValidPassword"
	})
	t.Run("type", func(t *testing.T) {
		parameter.Type = "invalid"
		err := validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.Type = ""
		err = validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.Type = "TiDB"
	})
	t.Run("version", func(t *testing.T) {
		parameter.Version = ""
		err := validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.Version = "invalid"
		err = validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.Version = "TiDB"
	})
	t.Run("region", func(t *testing.T) {
		parameter.Region = ""
		err := validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.Region = "TooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLongTooLong"
		err = validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.Version = "region1"
	})
	t.Run("CpuArchitecture", func(t *testing.T) {
		parameter.CpuArchitecture = ""
		err := validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.CpuArchitecture = "invalid"
		err = validator.New().Struct(parameter)
		assert.Error(t, err)

		parameter.CpuArchitecture = "X86"
	})
}