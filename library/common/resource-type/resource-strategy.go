package resource

type Applicant struct {
	HolderId  string
	RequestId string
}

type Location struct {
	Region string
	Zone   string
	Rack   string
	Host   string
}

type Excluded struct {
	Hosts []string
}

type Filter struct {
	Purpose  string
	DiskType string
}

type ComputeRequirement struct {
	CpuCores int32
	Memory   int32
}

type DiskRequirement struct {
	NeedDisk bool
	Capacity int32  // Reserved, not used by now
	DiskType string // Reserved, not used by now
}

type PortRequirement struct {
	Start   int32
	End     int32
	PortCnt int32
}
type Requirement struct {
	DiskReq    DiskRequirement
	ComputeReq ComputeRequirement
	PortReq    []PortRequirement
}

type AllocStrategy int32

const (
	RandomRack         AllocStrategy = iota // Require 'Region' and 'Zone', return diff Host
	DiffRackBestEffort                      // Require 'Region' and 'Zone', try best effort to alloc host in diff rack
	UserSpecifyRack                         // Require 'Region' 'Zone' and 'Rack', return diff hosts in Rack
	UserSpecifyHost                         // Return Resource in the Host Specified
)

type AllocRequirement struct {
	Location     Location
	HostExcluded Excluded
	HostFilter   Filter
	Require      Requirement
	Strategy     AllocStrategy
	Count        int32
}

type AllocReq struct {
	Applicant Applicant
	Requires  []AllocRequirement
}

type DiskResource struct {
	DiskId   string
	DiskName string
	Path     string
	Type     string
	Capacity int32
}

type PortResource struct {
	Start int32
	End   int32
	Ports []int32
}

type HostResource struct {
	Reqseq     int32
	Location   Location
	HostId     string
	HostName   string
	HostIp     string
	UserName   string
	Passwd     string
	ComputeRes ComputeRequirement
	DiskRes    DiskResource
	PortRes    []PortResource
}

type AllocRsp struct {
	Results []HostResource
}
