package backuprestore

import (
	"time"

	"github.com/pingcap-inc/tiem/micro-api/controller"
)

type BackupRecord struct {
	ID           int64                 `json:"id"`
	ClusterId    string                `json:"clusterId"`
	StartTime    time.Time             `json:"startTime"`
	EndTime      time.Time             `json:"endTime"`
	BackupType   string                `json:"backupType"`   // 全量/增量
	BackupMethod string                `json:"backupMethod"` // 物理/逻辑
	BackupMode   string                `json:"backupMode"`   // 手动/自动
	Operator     controller.Operator   `json:"operator"`
	Size         float32               `json:"size"`
	Status       controller.StatusInfo `json:"status"`
	FilePath     string                `json:"filePath"`
}
