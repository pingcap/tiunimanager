package models

import (
	"time"
)

type TransportRecord struct {
	ID				string 		`gorm:"primaryKey"`
	ClusterId 		string
	TransportType   string
	FilePath 		string
	TenantId       	string
	Status 			string
	StratTime 		time.Time
	EndTime			time.Time
}

func (rc *TransportRecord) TableName() string {
	return "transport_record"
}

func CreateTransportRecord(record *TransportRecord) (id string, err error) {
	err = MetaDB.Create(record).Error
	if err != nil {
		return "", err
	}
	return record.ID,nil
}

func UpdateTransportRecord(id, clusterId, status string, endTime time.Time) (err error) {
	record := TransportRecord{
		ID: id,
		ClusterId: clusterId,
	}
	err = MetaDB.Model(&record).Updates(map[string]interface{}{"Status": status, "EndTime": endTime}).Error
	return err
}

func FindTransportRecordById(id string) (record TransportRecord, err error) {
	err = MetaDB.First(&record, &TransportRecord{ID: id}).Error
	if err != nil {
		return record, err
	}
	return record, nil
}