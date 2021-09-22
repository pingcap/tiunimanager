package logapi

type SearchTiDBLogRsp struct {
	Took    int                   `json:"took"`
	Results []SearchTiDBLogDetail `json:"results"`
}

type SearchTiDBLogDetail struct {
	Index      string                 `json:"index"`
	Id         string                 `json:"id"`
	Level      string                 `json:"level"`
	SourceLine string                 `json:"sourceLine"`
	Message    string                 `json:"message"`
	Ip         string                 `json:"ip"`
	ClusterId  string                 `json:"clusterId"`
	Module     string                 `json:"module"`
	Ext        map[string]interface{} `json:"ext"`
	Timestamp  string                 `json:"timestamp"`
}
