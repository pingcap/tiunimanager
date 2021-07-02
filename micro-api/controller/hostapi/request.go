package hostapi

import (
	"github.com/pingcap/ticp/micro-api/controller"
)

type HostQuery struct {
	controller.PageRequest
}

type Disk struct {
	Name     string `json:"name"`     // [sda/sdb/nvmep0...]
	Capacity int32  `json:"capacity"` // Disk size, Unit: GB
	Path     string `json:"path"`     // Disk mount path: [/data1]
	Status   int32  `json:"status"`   // Disk Status, 0 for available, 1 for inused
}
type HostInfo struct {
	HostName string `json:"hostName"`
	Dc       string `json:"dc"`
	Az       string `json:"az"`
	Rack     string `json:"rack"`
	Ip       string `json:"ip"`
	Status   int32  `json:"status"` // Host Status, 0 for Online, 1 for offline
	Os       string `json:"os"`
	Kernel   string `json:"kernel"`
	CpuCores int32  `json:"cpuCores"`
	Memory   int32  `json:"memory"`  // Host memory size, Unit:GB
	Nic      string `json:"nic"`     // Host network type: 1GE or 10GE
	Purpose  string `json:"purpose"` // What Purpose is the host used for? [compute/storage or both]
	Disks    []Disk `json:"disks"`
}

type ListHostCondition struct {
	Status  int32  `json:"status"`  // Host Status, 0 for Online, 1 for offline
	Purpose string `json:"purpose"` // What Purpose is the host used for? [compute/storage or both]
}
