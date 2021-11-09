
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

package knowledge

import (
	"errors"
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"strconv"
	"strings"
)

type ResourceSpec struct {
	SpecItems []ResourceSpecItem	`json:"specItems"`
}

func (spec *ResourceSpec) GetAttributeValue(attribute ResourceSpecAttribute) (interface{}, error) {
	for _, i := range spec.SpecItems {
		if attribute == i.Attribute {
			return i.Value, nil
		}
	}

	return nil, errors.New("")
}

type ResourceSpecAttribute string

const (
	CpuCoreCount ResourceSpecAttribute = "CpuCoreCount"
	MemorySize 	 ResourceSpecAttribute = "MemorySize"
	DiskSize     ResourceSpecAttribute = "DiskSize"
	DiskType     ResourceSpecAttribute = "DiskType"
)

type ResourceSpecItem struct {
	Attribute ResourceSpecAttribute
	Value     interface{}
}

func ParseCpu(specCode string) int {
	cpu, err := strconv.Atoi(strings.Split(specCode, "C")[0])
	if err != nil {
		framework.Log().Errorf("ParseCpu error, specCode = %s", specCode)
	}
	return cpu
}

func ParseMemory(specCode string) int {
	memory, err := strconv.Atoi(strings.Split(strings.Split(specCode, "C")[1], "G")[0])
	if err != nil {
		framework.Log().Errorf("ParseMemory error, specCode = %s", specCode)
	}
	return memory
}

func GenSpecCode(cpuCores int32, mem int32) string {
	return fmt.Sprintf("%dC%dG", cpuCores, mem)
}