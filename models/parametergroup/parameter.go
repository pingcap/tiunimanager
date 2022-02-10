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
 ******************************************************************************/

/*******************************************************************************
 * @File: parameter.go
 * @Description: parameter table orm
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 14:37
*******************************************************************************/

package parametergroup

import "time"

// Parameter
// @Description: parameter table orm
type Parameter struct {
	ID             string    `gorm:"primaryKey;comment:'parameter id'" json:"id"`
	Category       string    `gorm:"uniqueIndex:idx_category_name_instancetype;not null;comment:'parameter category'" json:"category"`
	Name           string    `gorm:"uniqueIndex:idx_category_name_instancetype;not null;comment:'parameter name'" json:"name"`
	InstanceType   string    `gorm:"uniqueIndex:idx_category_name_instancetype;not null;comment:'parameter instance type. e.g.: TiDB, TiKV, PD'" json:"instanceType"`
	SystemVariable string    `gorm:"comment:'parameter mapped system variables for updating parameter values by sql and api'" json:"systemVariable"`
	Type           int       `gorm:"default:0;comment:'parameter type. optional values: [0: int, 1: string, 2: bool, 3: float, 4: array]'" json:"type"`
	Unit           string    `gorm:"comment:'unit of parameter values'" json:"unit"`
	Range          string    `gorm:"comment:'range of parameter values'" json:"range"`
	HasReboot      int       `gorm:"default:0;comment:'whether parameter updates require a restart. optional values: [0: false, 1: true]'" json:"hasReboot"`
	HasApply       int       `gorm:"default:1;comment:'whether to apply the parameter set by default. optional values: [0: false, 1: true]'" json:"hasApply"`
	UpdateSource   int       `gorm:"default:0;comment:'parameter update data source. optional values: [0: tiup, 1: sql, 2: tiup+sql, 3: api]'" json:"updateSource"`
	ReadOnly       int       `gorm:"default:0;comment:'specify if the parameter is read-only. optional values: [0: false, 1: true]'" json:"readOnly"`
	Description    string    `gorm:"comment:'parameter description information'" json:"description"`
	CreatedAt      time.Time `gorm:"<-:create" json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
