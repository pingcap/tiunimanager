package common

var (
	TemplateFileName = "hostInfo_template.xlsx"
	TemplateFilePath = "./etc"
)

type DiskType string

const (
	NvmeSSD DiskType = "nvme_ssd"
	SSD     DiskType = "ssd"
	Sata    DiskType = "sata"
)

type HostStatus int32

const (
	HOST_WHATEVER HostStatus = iota - 1
	HOST_ONLINE
	HOST_OFFLINE
	HOST_INUSED
	HOST_EXHAUST
	HOST_DELETED
)

func (s HostStatus) IsValid() bool {
	return (s >= HOST_WHATEVER && s <= HOST_DELETED)
}

func (s HostStatus) IsInused() bool {
	return s == HOST_INUSED || s == HOST_EXHAUST
}

func (s HostStatus) IsAvailable() bool {
	return (s == HOST_ONLINE || s == HOST_INUSED)
}

type DiskStatus int32

const (
	DISK_AVAILABLE DiskStatus = iota
	DISK_INUSED
)

func (s DiskStatus) IsInused() bool {
	return s == DISK_INUSED
}

func (s DiskStatus) IsAvailable() bool {
	return s == DISK_AVAILABLE
}
