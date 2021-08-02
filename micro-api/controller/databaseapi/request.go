package databaseapi

type DataExport struct {
	TenantId      string     `json:"tenantId"`
	ClusterId     string     `json:"clusterId"`
	UserName      string     `json:"userName"`
	Password      string     `json:"password"`
	FileType      string     `json:"fileType"`
	Filter 		  string 	 `json:"filter"`
}

type DataImport struct {
	TenantId      string     `json:"tenantId"`
	ClusterId     string     `json:"clusterId"`
	UserName      string     `json:"userName"`
	Password      string     `json:"password"`
	DataDir       string     `json:"dataDir"`
}

type DataTransportQuery struct {
	//controller.PageRequest
	TenantId      string     `json:"tenantId"`
	ClusterId     string     `json:"clusterId"`
	FlowId        string     `json:"flowId"`//通过flowId找到tiup的taskId，返回任务状态
}


