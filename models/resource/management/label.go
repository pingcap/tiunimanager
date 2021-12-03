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

package management

import (
	"time"

	common2 "github.com/pingcap-inc/tiem/common"
	"github.com/pingcap-inc/tiem/common/resource"
	"github.com/pingcap-inc/tiem/library/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"gorm.io/gorm"
)

type Label struct {
	Name      string `gorm:"PrimaryKey"`
	Category  int8
	Trait     int64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type Labels map[string]Label

func (labels Labels) getTraitByName(name string) (trait int64, err error) {
	if label, ok := (labels)[name]; ok {
		return label.Trait, nil
	} else {
		return 0, framework.NewTiEMErrorf(common.TIEM_RESOURCE_TRAIT_NOT_FOUND, "label type %v not found in system default labels", name)
	}
}

func (labels Labels) getLabelNamesByTraits(traits int64) (labelNames []string) {
	for _, v := range labels {
		if traits&v.Trait == v.Trait {
			labelNames = append(labelNames, v.Name)
		}
	}
	return
}

var DefaultLabelTypes = Labels{
	string(common2.EMProductNameTiDB):          {Name: string(common2.EMProductNameTiDB), Category: int8(resource.Cluster), Trait: 0x0000000000000001},
	string(common2.EMProductNameDataMigration): {Name: string(common2.EMProductNameDataMigration), Category: int8(resource.Cluster), Trait: 0x0000000000000002},
	string(resource.PurposeCompute):            {Name: string(resource.PurposeCompute), Category: int8(resource.Component), Trait: 0x0000000000000004},
	string(resource.PurposeStorage):            {Name: string(resource.PurposeStorage), Category: int8(resource.Component), Trait: 0x0000000000000008},
	string(resource.PurposeSchedule):           {Name: string(resource.PurposeSchedule), Category: int8(resource.Component), Trait: 0x0000000000000010},
	string(resource.NVMeSSD):                   {Name: string(resource.NVMeSSD), Category: int8(resource.DiskPerf), Trait: 0x0000000000000020},
	string(resource.SSD):                       {Name: string(resource.SSD), Category: int8(resource.DiskPerf), Trait: 0x0000000000000040},
	string(resource.SATA):                      {Name: string(resource.SATA), Category: int8(resource.DiskPerf), Trait: 0x0000000000000080},
}

func GetTraitByName(name string) (trait int64, err error) {
	return DefaultLabelTypes.getTraitByName(name)
}

func GetLabelNamesByTraits(traits int64) (labelNames []string) {
	return DefaultLabelTypes.getLabelNamesByTraits(traits)
}
