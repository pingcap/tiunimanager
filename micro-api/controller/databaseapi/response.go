package databaseapi

import (
	"time"
)

type DataExportResp struct {
	RecordId        string     `json:"recordId"`
}

type DataImportResp struct {
	RecordId        string     `json:"recordId"`
}

type DataTransportInfo struct {
	RecordId        string      	`json:"recordId"`
	ClusterId     	string			`json:"clusterId"`
	TransportType 	string    		`json:"transportType"`
	StartTime   	time.Time    	`json:"startTime"`
	EndTime   		time.Time    	`json:"endTime"`
	Status        	string         	`json:"status"`
	FilePath       	string       	`json:"filePath"`
}
