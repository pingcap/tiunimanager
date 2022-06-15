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

/*******************************************************************************
 * @File: parametergroup_mapping.go
 * @Description: parameter group mapping table orm
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 14:46
*******************************************************************************/

package parametergroup

import "time"

// ParameterGroupMapping
// @Description: parameter_group_mapping table orm
type ParameterGroupMapping struct {
	ParameterGroupID string    `gorm:"primaryKey;comment:'parameter group id, relation parameter_group table'"`
	ParameterID      string    `gorm:"primaryKey;comment:'group id, relation parameter table'"`
	DefaultValue     string    `gorm:"not null;comment:'parameter default value'"`
	Note             string    `gorm:"comment:'parameter remark information'"`
	CreatedAt        time.Time `gorm:"<-:create"`
	UpdatedAt        time.Time
}

// ParamDetail
// @Description: parameter group detail object
type ParamDetail struct {
	Parameter

	DefaultValue string `json:"defaultValue"`
	Note         string `json:"note"`
}
