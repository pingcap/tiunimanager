/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
 * @File: product.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/6
*******************************************************************************/

package product

import (
	"encoding/json"
	"github.com/pingcap/tiunimanager/common/errors"
	"gorm.io/gorm"
)

// ProductInfo info provided by Enterprise Manager
type ProductInfo struct {
	ProductID   string `gorm:"uniqueIndex;size:32"`
	ProductName string `gorm:"size:32"`
}

// ProductVersion info provided by Enterprise Manager
type ProductVersion struct {
	ProductID string `gorm:"uniqueIndex:product_arch_version;size:32"`
	Arch      string `gorm:"uniqueIndex:product_arch_version;size:32"`
	Version   string `gorm:"uniqueIndex:product_arch_version;size:32"`
	Desc      string `gorm:""`
}

//ProductComponentInfo Enterprise Manager offers two products, TiDB and TiDB Data Migration,
//each consisting of multiple components, each with multiple ports enabled and a limit on the number of instances each component can start.
type ProductComponentInfo struct {
	ProductID               string  `gorm:"uniqueIndex:product_component;"`
	ComponentID             string  `gorm:"uniqueIndex:product_component;"`
	ComponentName           string  `gorm:"size:32;"`
	PurposeType             string  `gorm:"size:32;comment:Compute/Storage/Schedule"`
	StartPort               int32   `gorm:"comment:'starting value of the port opened by the component'"`
	EndPort                 int32   `gorm:"comment:'ending value of the port opened by the component'"`
	MaxPort                 int32   `gorm:"comment:'maximum number of ports that can be opened by the component'"`
	MinInstance             int32   `gorm:"default:0;comment:'minimum number of instances started by the component'"`
	MaxInstance             int32   `gorm:"default:128;comment:'PD: 1、3、5、7，Prometheus：1，AlertManger：1，Grafana'"`
	SuggestedInstancesCount []int32 `gorm:"-"`
	SuggestedInstancesInfo  string  `gorm:""`
}

func (t *ProductComponentInfo) BeforeSave(tx *gorm.DB) (err error) {
	if t.SuggestedInstancesCount == nil {
		t.SuggestedInstancesCount = make([]int32, 0)
	}
	p, jsonErr := json.Marshal(t.SuggestedInstancesCount)
	if jsonErr == nil {
		t.SuggestedInstancesInfo = string(p)
	} else {
		return errors.NewError(errors.TIUNIMANAGER_PARAMETER_INVALID, jsonErr.Error())
	}

	return nil
}

func (t *ProductComponentInfo) AfterFind(tx *gorm.DB) (err error) {
	if len(t.SuggestedInstancesInfo) > 0 {
		t.SuggestedInstancesCount = make([]int32, 0)
		err = json.Unmarshal([]byte(t.SuggestedInstancesInfo), &t.SuggestedInstancesCount)
	}
	return err
}
