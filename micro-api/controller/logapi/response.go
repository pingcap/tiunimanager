package logapi

type SearchTiDBLogRsp struct {
	Took    int                   `json:"took" example:"10"`
	Results []SearchTiDBLogDetail `json:"results"`
}

type SearchTiDBLogDetail struct {
	Index      string                 `json:"index" example:"tiem-tidb-cluster-2021.09.23"`
	Id         string                 `json:"id" example:"zvadfwf"`
	Level      string                 `json:"level" example:"warn"`
	SourceLine string                 `json:"sourceLine" example:"main.go:210"`
	Message    string                 `json:"message"  example:"tidb log"`
	Ip         string                 `json:"ip" example:"127.0.0.1"`
	ClusterId  string                 `json:"clusterId" example:"abc"`
	Module     string                 `json:"module" example:"tidb"`
	Ext        map[string]interface{} `json:"ext"`
	Timestamp  string                 `json:"timestamp" example:"2021-09-23 14:23:10"`
}
