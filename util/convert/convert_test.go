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
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: convert_test.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/11/26 18:06
*******************************************************************************/

package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type src struct {
	Id      int64    `json:"id"`
	Name    string   `json:"name"`
	Hobbies []string `json:"hobbies"`
	Gender  bool     `json:"gender"`
}

type dst src

func TestConvertObj(t *testing.T) {
	s := src{
		Id:      1,
		Name:    "zhangsan",
		Hobbies: []string{"movie", "running", "swimming"},
		Gender:  true,
	}
	d := new(dst)
	err := ConvertObj(s, &d)
	assert.Nil(t, err)
	assert.Equal(t, "zhangsan", d.Name)
	assert.Equal(t, 1, int(d.Id))
	assert.Equal(t, 3, len(d.Hobbies))
	assert.True(t, d.Gender)
}
