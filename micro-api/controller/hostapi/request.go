package hostapi

import (
	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type HostQuery struct {
	controller.PageRequest
	Purpose string `json:"purpose" form:"purpose"`
	Status  int    `json:"status" form:"status"`
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
