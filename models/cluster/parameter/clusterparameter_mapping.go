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
 * @File: clusterparameter_mapping.go
 * @Description: cluster parameter mapping table orm
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/10 14:50
*******************************************************************************/

package parameter

import (
	"time"

	"github.com/pingcap/tiunimanager/models/parametergroup"
)

// ClusterParameterMapping
// @Description: cluster_parameter_mapping table orm
type ClusterParameterMapping struct {
	ClusterID   string    `gorm:"primaryKey;comment:'cluster id. relation cluster table'"`
	ParameterID string    `gorm:"primaryKey;comment:'parameter id. relation parameter table'"`
	RealValue   string    `gorm:"not null;comment:'cluster parameter real value. value format: {\"clusterValue\":1, \"instanceValue\":[{\"id\":1,\"value\":2}]}'"`
	CreatedAt   time.Time `gorm:"<-:create"`
	UpdatedAt   time.Time
}

// ClusterParamDetail
// @Description: cluster parameter detail object
type ClusterParamDetail struct {
	parametergroup.Parameter

	DefaultValue string `json:"defaultValue"`
	RealValue    string `json:"realValue"`
	Note         string `json:"note"`
}
