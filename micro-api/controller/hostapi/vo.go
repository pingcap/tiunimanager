package hostapi

type ZoneHostStock struct {
	ZoneBaseInfo
	SpecBaseInfo
	Count int
}

type ZoneBaseInfo struct {
	ZoneCode string
	ZoneName string
}

type SpecBaseInfo struct {
	SpecCode string
	SpecName string
}

type HostInfo struct {
	ID        string     `json:"hostId"`
	IP        string     `json:"ip"`
	UserName  string     `json:"userName,omitempty"`
	Passwd    string     `json:"passwd,omitempty"`
	HostName  string     `json:"hostName"`
	Status    int32      `json:"status"` // Host Status, 0 for Online, 1 for offline
	Arch      string     `json:"arch"`   // x86 or arm64
	OS        string     `json:"os"`
	Kernel    string     `json:"kernel"`
	CpuCores  int32      `json:"cpuCores"`
	Memory    int32      `json:"memory"` // Host memory size, Unit:GB
	Spec      string     `json:"spec"`   // Host Spec, init while importing
	Nic       string     `json:"nic"`    // Host network type: 1GE or 10GE
	Region    string     `json:"region"`
	AZ        string     `json:"az"`
	Rack      string     `json:"rack"`
	Purpose   string     `json:"purpose"`  // What Purpose is the host used for? [compute/storage/general]
	DiskType  string     `json:"diskType"` // Disk type of this host [sata/ssd/nvme_ssd]
	Reserved  bool       `json:"reserved"` // Whether this host is reserved - will not be allocated
	CreatedAt int64      `json:"createTime"`
	Disks     []DiskInfo `json:"disks"`
}

type DiskInfo struct {
	ID       string `json:"diskId"`
	HostId   string `json:"hostId,omitempty"`
	Name     string `json:"name"`     // [sda/sdb/nvmep0...]
	Capacity int32  `json:"capacity"` // Disk size, Unit: GB
	Path     string `json:"path"`     // Disk mount path: [/data1]
	Type     string `json:"type"`     // Disk type: [nvme-ssd/ssd/sata]
	Status   int32  `json:"status"`   // Disk Status, 0 for available, 1 for inused
	UsedBy   string `json:"usedBy,omitempty"`
}
