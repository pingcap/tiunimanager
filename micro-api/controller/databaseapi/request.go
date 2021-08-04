package databaseapi

type DataExport struct {
	ClusterId     string     `json:"clusterId"`
	UserName      string     `json:"userName"`
	Password      string     `json:"password"`
	FileType      string     `json:"fileType"`
	Filter 		  string 	 `json:"filter"`
}

type DataImport struct {
	ClusterId     string     `json:"clusterId"`
	UserName      string     `json:"userName"`
	Password      string     `json:"password"`
	FilePath      string     `json:"filePath"`
}

type DataTransportQuery struct {
	ClusterId     string     `json:"clusterId"`
	RecordId      string     `json:"recordId"`
}


