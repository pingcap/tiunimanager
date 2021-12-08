/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
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

package parameter

import (
	"github.com/pingcap-inc/tiem/library/knowledge"
)

type ParamRealValue struct {
	Cluster   string                    `json:"cluster" example:"1"`
	Instances []*ParamInstanceRealValue `json:"instances"`
}

type ParamInstanceRealValue struct {
	Instance string `json:"instance" example:"172.16.10.2"`
	Value    string `json:"value" example:"2"`
}

type ParamItem struct {
	Definition   knowledge.Parameter `json:"definition"`
	CurrentValue ParamInstance       `json:"currentValue"`
}

type ParamInstance struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}
