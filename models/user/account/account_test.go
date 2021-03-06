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

package account

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUser_GenSaltAndHash_CheckPassword(t *testing.T) {
	user := &User{
		ID:       "user01",
		Name:     "user",
	}

	err := user.GenSaltAndHash("12345")
	assert.NoError(t, err)

	_, err = user.CheckPassword("")
	assert.Error(t, err)

	_, err = user.CheckPassword("123456789012345678901")
	assert.Error(t, err)

	got, err := user.CheckPassword("1234")
	assert.NoError(t, err)
	assert.False(t, got)

	got, err = user.CheckPassword("12345")
	assert.NoError(t, err)
	assert.True(t, got)
}
