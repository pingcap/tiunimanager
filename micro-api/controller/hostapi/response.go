package hostapi

type ImportHostRsp struct {
	HostId string `json:"hostId"`
}

type ListHostRsp struct {
	Hosts []HostInfo `json:"hosts"`
}

type HostDetailsRsp struct {
	Host HostInfo `json:"host"`
}
