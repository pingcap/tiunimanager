package domain

type BackupRecord struct {
	Id         uint
	ClusterId  string
	Range      BackupRange
	BackupType BackupType
	OperatorId string
	Size       float32
	FilePath   string
}

type RecoverRecord struct {
	Id         uint
	ClusterId  string
	OperatorId string
	BackupRecord   BackupRecord
}

type BackupStrategy struct {
	ValidityPeriod  	int64
	CronString 			string
}
