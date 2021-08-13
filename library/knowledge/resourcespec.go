package knowledge

import (
	"errors"
	"fmt"
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
	// todo
	return 4
}

func ParseMemory(specCode string) int {
	// todo
	return 8
}

func GenSpecCode(cpuCores int32, mem int32) string {
	return fmt.Sprintf("%dC%dG", cpuCores, mem)
}