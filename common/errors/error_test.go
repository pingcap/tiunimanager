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

package errors

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestError(t *testing.T) {
	err := Error(TIUNIMANAGER_MARSHAL_ERROR)
	assert.Error(t, err)
	assert.Equal(t, TIUNIMANAGER_MARSHAL_ERROR, err.code)
	assert.Equal(t, TIUNIMANAGER_MARSHAL_ERROR.Explain(), err.GetMsg())
	assert.Contains(t, err.Error(), TIUNIMANAGER_MARSHAL_ERROR.Explain())
	assert.Contains(t, err.Error(), "[10004]")
	fmt.Print(err.Error())
}

func TestNewError(t *testing.T) {
	err := NewError(TIUNIMANAGER_MARSHAL_ERROR, "mymessage")
	assert.Error(t, err)
	assert.Equal(t, TIUNIMANAGER_MARSHAL_ERROR, err.code)
	assert.Equal(t, "mymessage", err.GetMsg())
	assert.Contains(t, err.Error(), TIUNIMANAGER_MARSHAL_ERROR.Explain())
	assert.Contains(t, err.Error(), "[10004]")
	assert.Contains(t, err.Error(), "mymessage")
	fmt.Print(err.Error())
}

func TestNewEMErrorf(t *testing.T) {
	err := NewErrorf(TIUNIMANAGER_MARSHAL_ERROR, "mymessage %d", 3333)
	assert.Error(t, err)
	assert.Equal(t, TIUNIMANAGER_MARSHAL_ERROR, err.code)
	assert.Equal(t, "mymessage 3333", err.GetMsg())
	assert.Contains(t, err.Error(), TIUNIMANAGER_MARSHAL_ERROR.Explain())
	assert.Contains(t, err.Error(), "[10004]")
	assert.Contains(t, err.Error(), "mymessage")
	assert.Contains(t, err.Error(), "3333")

	fmt.Print(err.Error())
}

func TestWrapError(t *testing.T) {
	err := WrapError(TIUNIMANAGER_MARSHAL_ERROR, "mymessage", Error(TIUNIMANAGER_UNRECOGNIZED_ERROR))
	assert.Error(t, err)
	assert.Equal(t, TIUNIMANAGER_MARSHAL_ERROR, err.code)
	assert.Equal(t, "mymessage", err.GetMsg())
	assert.Contains(t, err.Error(), TIUNIMANAGER_MARSHAL_ERROR.Explain())
	assert.Contains(t, err.Error(), "[10004]")
	assert.Contains(t, err.Error(), "mymessage")
	assert.Contains(t, err.Error(), "[10000]")
	assert.Contains(t, err.Error(), TIUNIMANAGER_UNRECOGNIZED_ERROR.Explain())
	fmt.Print(err.Error())
}