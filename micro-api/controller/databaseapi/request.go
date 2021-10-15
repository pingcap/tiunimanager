package databaseapi

import "github.com/pingcap-inc/tiem/micro-api/controller"

type DataExportReq struct {
	ClusterId       string `json:"clusterId"`
	UserName        string `json:"userName"`
	Password        string `json:"password"`
	FileType        string `json:"fileType"`
	Filter          string `json:"filter"`
	Sql             string `json:"sql"`
	FilePath        string `json:"filePath"`
	StorageType     string `json:"storageType"`
	EndpointUrl     string `json:"endpointUrl"`
	BucketUrl       string `json:"bucketUrl"`
	BucketRegion    string `json:"bucketRegion"`
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type DataImportReq struct {
	ClusterId       string `json:"clusterId"`
	UserName        string `json:"userName"`
	Password        string `json:"password"`
	FilePath        string `json:"filePath"`
	StorageType     string `json:"storageType"`
	EndpointUrl     string `json:"endpointUrl"`
	BucketUrl       string `json:"bucketUrl"`
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type DataTransportQueryReq struct {
	controller.PageRequest
	RecordId string `json:"recordId" form:"recordId"`
}
