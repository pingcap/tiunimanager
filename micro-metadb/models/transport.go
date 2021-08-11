package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type TransportRecord struct {
	ID            string `gorm:"primaryKey"`
	ClusterId     string
	TransportType string
	FilePath      string
	TenantId      string
	Status        string
	StartTime     time.Time
	EndTime       time.Time
}

func (rc *TransportRecord) TableName() string {
	return "transport_record"
}

func (rc *TransportRecord) BeforeCreate(tx *gorm.DB) error {
	if rc.ID == "" {
		rc.ID = uuid.New().String()
	}
	return nil
}

func CreateTransportRecord(record *TransportRecord) (id string, err error) {
	err = MetaDB.Create(record).Error
	if err != nil {
		return "", err
	}
	return record.ID, nil
}

func UpdateTransportRecord(id, clusterId, status string, endTime time.Time) (err error) {
	record := TransportRecord{
		ID:        id,
		ClusterId: clusterId,
	}
	err = MetaDB.Model(&record).Updates(map[string]interface{}{"Status": status, "EndTime": endTime}).Error
	return err
}

func FindTransportRecordById(id string) (record *TransportRecord, err error) {
	err = MetaDB.First(record, &TransportRecord{ID: id}).Error
	if err != nil {
		return record, err
	}
	return record, nil
}

func ListTransportRecord(clusterId string, recordId string, offset int32, length int32) (records []*TransportRecord, total int64, err error) {
	records = make([]*TransportRecord, length)

	db := MetaDB.Model(TransportRecord{}).Where("cluster_id = ?", clusterId)
	if recordId != "" {
		db.Where("id = ?", recordId)
	}

	err = db.Count(&total).Offset(int(offset - 1)).Limit(int(length)).Find(&records).Error

	return
}
