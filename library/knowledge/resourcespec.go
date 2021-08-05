package knowledge

import "errors"

type ResourceSpec struct {
	SpecItems []ResourceSpecItem
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