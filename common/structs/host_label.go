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
 * @File: resource.go
 * @Description:
 * @Author: duanbing@pingcap.com
 * @Version: 1.0.0
 * @Date: 2021/12/4
*******************************************************************************/

package structs

import (
	"github.com/pingcap/tiunimanager/common/constants"
	"github.com/pingcap/tiunimanager/common/errors"
)

type Label struct {
	Name     string
	Category int8
	Trait    int64
}

func GetTraitByName(name string) (trait int64, err error) {
	return DefaultLabelTypes.getTraitByName(name)
}

func GetLabelNamesByTraits(traits int64) (labelNames []string) {
	return DefaultLabelTypes.getLabelNamesByTraits(traits)
}

type Labels map[string]Label

var DefaultLabelTypes = Labels{
	string(constants.EMProductIDTiDB):              {Name: string(constants.EMProductIDTiDB), Category: int8(constants.Cluster), Trait: 0x0000000000000001},
	string(constants.EMProductIDDataMigration):     {Name: string(constants.EMProductIDDataMigration), Category: int8(constants.Cluster), Trait: 0x0000000000000002},
	string(constants.EMProductIDEnterpriseManager): {Name: string(constants.EMProductIDEnterpriseManager), Category: int8(constants.Cluster), Trait: 0x0000000000000004},
	string(constants.PurposeCompute):               {Name: string(constants.PurposeCompute), Category: int8(constants.Component), Trait: 0x0000000000000008},
	string(constants.PurposeStorage):               {Name: string(constants.PurposeStorage), Category: int8(constants.Component), Trait: 0x0000000000000010},
	string(constants.PurposeSchedule):              {Name: string(constants.PurposeSchedule), Category: int8(constants.Component), Trait: 0x0000000000000020},
	string(constants.NVMeSSD):                      {Name: string(constants.NVMeSSD), Category: int8(constants.DiskPerf), Trait: 0x0000000000000040},
	string(constants.SSD):                          {Name: string(constants.SSD), Category: int8(constants.DiskPerf), Trait: 0x0000000000000080},
	string(constants.SATA):                         {Name: string(constants.SATA), Category: int8(constants.DiskPerf), Trait: 0x0000000000000100},
}

func (labels Labels) getTraitByName(name string) (trait int64, err error) {
	if label, ok := (labels)[name]; ok {
		return label.Trait, nil
	} else {
		return 0, errors.NewErrorf(errors.TIUNIMANAGER_RESOURCE_TRAIT_NOT_FOUND, "label type %v not found in system default labels", name)
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
