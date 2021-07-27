package domain

import (
	"time"
)

type TransportType uint32
const (
	TransportTypeExport TransportType = 0
	TransportTypeImport TransportType = 1
)

type TransportRecord struct {
	Id 				uint
	ClusterId 		string
	StartTime 		time.Time
	EndTime 		time.Time
	Operator 		Operator
	TransportType   TransportType
	FilePath 		string
}

type ImportInfo struct {
	ClusterId 		string
	UserName		string
	Password 		string
	FilePath 		string
}

type ExportInfo struct {
	ClusterId 		string
	UserName		string
	Password 		string
	FileType 		string
}