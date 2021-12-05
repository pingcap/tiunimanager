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

	"github.com/pingcap-inc/tiem/common/constants"
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
	string(constants.EMProductNameTiDB):          {Name: string(constants.EMProductNameTiDB), Category: int8(constants.Cluster), Trait: 0x0000000000000001},
	string(constants.EMProductNameDataMigration): {Name: string(constants.EMProductNameDataMigration), Category: int8(constants.Cluster), Trait: 0x0000000000000002},
	string(constants.PurposeCompute):             {Name: string(constants.PurposeCompute), Category: int8(constants.Component), Trait: 0x0000000000000004},
	string(constants.PurposeStorage):             {Name: string(constants.PurposeStorage), Category: int8(constants.Component), Trait: 0x0000000000000008},
	string(constants.PurposeSchedule):            {Name: string(constants.PurposeSchedule), Category: int8(constants.Component), Trait: 0x0000000000000010},
	string(constants.NVMeSSD):                    {Name: string(constants.NVMeSSD), Category: int8(constants.DiskPerf), Trait: 0x0000000000000020},
	string(constants.SSD):                        {Name: string(constants.SSD), Category: int8(constants.DiskPerf), Trait: 0x0000000000000040},
	string(constants.SATA):                       {Name: string(constants.SATA), Category: int8(constants.DiskPerf), Trait: 0x0000000000000080},
}

func GetTraitByName(name string) (trait int64, err error) {
	return DefaultLabelTypes.getTraitByName(name)
}

func GetLabelNamesByTraits(traits int64) (labelNames []string) {
	return DefaultLabelTypes.getLabelNamesByTraits(traits)
}
