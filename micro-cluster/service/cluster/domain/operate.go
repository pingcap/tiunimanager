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

package domain

import (
	"time"
)

const SystemOperator = "System"

type Operator struct {
	Id             string
	Name           string
	TenantId       string
	ManualOperator bool
}

// GetOperatorFromName
// @Description: get operator from name
// @Parameter name
// @return *Operator
func GetOperatorFromName(name string) *Operator {
	if name == SystemOperator {
		return BuildSystemOperator()
	}

	// todo get from repository
	return &Operator{
		Name:           name,
		ManualOperator: true,
	}
}

// BuildSystemOperator
// @Description: system operator
// @return *Operator
func BuildSystemOperator() *Operator {
	return &Operator{
		Id:             SystemOperator,
		Name:           SystemOperator,
		ManualOperator: false,
	}
}

type OperateRecord struct {
	Id          uint
	TraceId     string
	Operator    Operator
	OperateTime time.Time
	OperateType int
	FlowWorkId  string
}