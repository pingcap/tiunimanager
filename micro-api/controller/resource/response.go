package hostapi

type DemoHostInfo struct {
	HostId   string `json:"hostId"`
	HostName string `json:"hostName"`
	HostIp   string `json:"hostIp"`
}

type ImportHostRsp struct {
	HostId string `json:"hostId"`
}

type ImportHostsRsp struct {
	HostIds []string `json:"hostIds"`
}
type ListHostRsp struct {
	Hosts []HostInfo `json:"hosts"`
}

type HostDetailsRsp struct {
	Host HostInfo `json:"host"`
}

type ZoneHostStockRsp struct {
	AvailableStocks map[string][]ZoneHostStock
}
type AllocateRsp struct {
	HostName string   `json:"hostName"`
	Ip       string   `json:"ip"`
	UserName string   `json:"userName"`
	Passwd   string   `json:"passwd"`
	CpuCores int32    `json:"cpuCore"`
	Memory   int32    `json:"memory"`
	Disk     DiskInfo `json:"disk"`
}

type AllocHostsRsp struct {
	PdHosts   []AllocateRsp `json:"pdHosts"`
	TidbHosts []AllocateRsp `json:"tidbHosts"`
	TikvHosts []AllocateRsp `json:"tikvHosts"`
}

type DomainResource struct {
	ZoneName string `json:"zoneName"`
	ZoneCode string `json:"zoneCode"`
	Purpose  string `json:"purpose"`
	SpecName string `json:"specName"`
	SpecCode string `json:"specCode"`
	Count    int32  `json:"count"`
}

type DomainResourceRsp struct {
	Resources []DomainResource `json:"resources"`
}
