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

/*******************************************************************************
 * @File: common_test.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/21 17:24
*******************************************************************************/

package parameter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var doc = []byte(`
{
	"student": {
		"name": "zhangsan",
		"age": 18,
		"weight": 52.4,
		"male": true,
		"hobby": ["swimming", "climbing", "music"],
		"test": null,
		"friends": [
			{
				"name": "lisi",
				"age": 22,
				"weight": 46.3,
				"male": false,
				"hobby": ["music", "books"]
			},
			{
				"name": "wangwu",
				"age": 19,
				"weight": 49.0,
				"male": true,
				"hobby": ["golf", "skating"]
			}
		]
	}
}`)

func Test_flattenedParameters(t *testing.T) {
	v := map[string]interface{}{}
	err := json.Unmarshal(doc, &v)
	assert.NoError(t, err)

	parameters, err := FlattenedParameters(v)
	assert.NoError(t, err)

	assert.Equal(t, parameters["student.name"], "zhangsan")
	assert.Equal(t, parameters["student.age"], "18")
	assert.Equal(t, parameters["student.weight"], "52.4")
	assert.Equal(t, parameters["student.friends[0].hobby"], "[\"music\",\"books\"]")
	assert.Equal(t, parameters["student.friends[1].age"], "19")
}
