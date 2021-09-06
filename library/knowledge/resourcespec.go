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
	MemorySize                         = "MemorySize"
	DiskSize                           = "DiskSize"
	DiskType                           = "DiskType"
)

type ResourceSpecItem struct {
	Attribute ResourceSpecAttribute
	Value     interface{}
}

func ParseCpu(specCode string) int {
	cpu, err := strconv.Atoi(strings.Split(specCode, "C")[0])
	if err != nil {
		framework.LogWithCaller().Errorf("ParseCpu error, specCode = %s", specCode)
	}
	return cpu
}

func ParseMemory(specCode string) int {
	memory, err := strconv.Atoi(strings.Split(strings.Split(specCode, "C")[1], "G")[0])
	if err != nil {
		framework.LogWithCaller().Errorf("ParseMemory error, specCode = %s", specCode)
	}
	return memory
}

func GenSpecCode(cpuCores int32, mem int32) string {
	return fmt.Sprintf("%dC%dG", cpuCores, mem)
}