package libbr

type ClusterFacade struct {
	ClusterId 	string
	ClusterName string
	PdAddress 	string
}

type BRPort interface {
	BackUpPreCheck(cluster ClusterFacade, storage BrStorage, bizId uint) error

	BackUp(cluster ClusterFacade, storage BrStorage, bizId uint) error

	ShowBackUpInfo(bizId uint) ProgressRate

	RestorePreCheck(cluster ClusterFacade, storage BrStorage, bizId uint) error

	Restore(cluster ClusterFacade, storage BrStorage, bizId uint) error

	ShowRestoreInfo(bizId uint) ProgressRate
}

type ProgressRate struct {
	Rate 	float32 // 0.99 means 99%
	Checked bool
	Error 	error
}

type BrStorage struct {
	StorageType StorageType
	Root string // "/tmp/backup"
}

type StorageType string

const (
	StorageTypeLocal 	StorageType = "local"
	StorageTypeS3 		StorageType = "s3"
)
