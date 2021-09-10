package hostapi

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type HostQuery struct {
	controller.PageRequest
	Purpose string `json:"purpose" form:"purpose"`
	Status  int    `json:"status" form:"status"`
}

type Disk struct {
	DiskId   string `json:"diskId"`
	Name     string `json:"name"`     // [sda/sdb/nvmep0...]
	Capacity int32  `json:"capacity"` // Disk size, Unit: GB
	Path     string `json:"path"`     // Disk mount path: [/data1]
	Type     string `json:"type"`     // Disk type: [nvme-ssd/ssd/sata]
	Status   int32  `json:"status"`   // Disk Status, 0 for available, 1 for inused
	UsedBy   string `json:"usedby"`   // Disk is used by which cluster
}
type HostInfo struct {
	HostId      string `json:"hostId"`
	HostName    string `json:"hostName"`
	Dc          string `json:"dc"`
	Az          string `json:"az"`
	Rack        string `json:"rack"`
	Ip          string `json:"ip"`
	UserName    string `json:"userName"`
	Passwd      string `json:"passwd"`
	Status      int32  `json:"status"` // Host Status, 0 for Online, 1 for offline
	Os          string `json:"os"`
	Kernel      string `json:"kernel"`
	CpuCores    int32  `json:"cpuCores"`
	Memory      int32  `json:"memory"`      // Host memory size, Unit:GB
	Spec        string `json:"spec"`        // Host Spec, init while importing
	Nic         string `json:"nic"`         // Host network type: 1GE or 10GE
	Purpose     string `json:"purpose"`     // What Purpose is the host used for? [compute/storage/general]
	Performance string `json:"performance"` // Performance type of this host [High/Medium/Low]
	CreatedAt   int64  `json:"createTime"`
	Disks       []Disk `json:"disks"`
}

type ExcelField int

const (
	HOSTNAME_FIELD ExcelField = iota
	IP_FILED
	USERNAME_FIELD
	PASSWD_FIELD
	DC_FIELD
	ZONE_FIELD
	RACK_FIELD
	OS_FIELD
	KERNEL_FIELD
	CPU_FIELD
	MEM_FIELD
	NIC_FIELD
	PURPOSE_FIELD
	PERF_FIELD
	DISKS_FIELD
)

type Allocation struct {
	FailureDomain string `json:"failureDomain"`
	CpuCores      int32  `json:"cpuCores"`
	Memory        int32  `json:"memory"`
	Count         int32  `json:"count"`
}

type AllocHostsReq struct {
	PdReq   []Allocation `json:"pdReq"`
	TidbReq []Allocation `json:"tidbReq"`
	TikvReq []Allocation `json:"tikvReq"`
}
