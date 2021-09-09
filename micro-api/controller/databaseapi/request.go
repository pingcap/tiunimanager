package databaseapi

import "github.com/pingcap-inc/tiem/micro-api/controller"

type DataExportReq struct {
	ClusterId     string     `json:"clusterId"`
	UserName      string     `json:"userName"`
	Password      string     `json:"password"`
	FileType      string     `json:"fileType"`
	Filter 		  string 	 `json:"filter"`
	FilePath 	  string 	 `json:"filePath"`
}

type DataImportReq struct {
	ClusterId     string     `json:"clusterId"`
	UserName      string     `json:"userName"`
	Password      string     `json:"password"`
	FilePath      string     `json:"filePath"`
}

type DataTransportQueryReq struct {
	controller.PageRequest
	RecordId	string		`json:"recordId" form:"recordId"`
}