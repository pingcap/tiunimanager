package domain

import (
	"time"
)

type BackupRecord struct {
	Id 				uint
	ClusterId 		string
	StartTime 		time.Time
	EndTime 		time.Time
	Range 			BackupRange
	Way 			BackupWay
	Operator 		Operator
	Size 			float32
	Status 			string
	FilePath 		string
}

type BackupRange 		int
type BackupWay 			int

type BackupStrategy struct {
	ValidityPeriod  	int64
	CronString 			string
}
