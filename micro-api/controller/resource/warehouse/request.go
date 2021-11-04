package warehouse

type HostFilter struct {
	Arch string `json:"Arch"`
}

type GetHierarchyReq struct {
	Filter HostFilter `json:"Filter"`
	Level  int32      `json:"Level"`
	Depth  int32      `json:"Depth"`
}
