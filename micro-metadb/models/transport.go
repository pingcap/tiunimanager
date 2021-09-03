package models

import (
	"strconv"
	"time"
)

type TransportRecord struct {
	Record
	ClusterId     string
	TransportType string
	FilePath      string
	TenantId      string
	Status        string
	StartTime     time.Time
	EndTime       time.Time
}

func (m *DAOClusterManager) CreateTransportRecord(record *TransportRecord) (id string, err error) {
	err = m.Db().Create(record).Error
	if err != nil {
		return "", err
	}
	return strconv.Itoa(int(record.ID)), nil
}

func (m *DAOClusterManager) UpdateTransportRecord(id, clusterId, status string, endTime time.Time) (err error) {
	record := TransportRecord{
		ClusterId: clusterId,
	}
	uintId, _ := strconv.ParseInt(id, 10, 64)
	record.ID = uint(uintId)
	err = m.Db().Model(&record).Updates(map[string]interface{}{"Status": status, "EndTime": endTime}).Error
	return err
}

func (m *DAOClusterManager) FindTransportRecordById(id string) (record *TransportRecord, err error) {
	record = &TransportRecord{}
	uintId, _ := strconv.ParseInt(id, 10, 64)

	err = m.Db().Where("id = ?", uintId).First(record).Error
	if err != nil {
		return record, err
	}
	return record, nil
}

func (m *DAOClusterManager) ListTransportRecord(clusterId string, recordId string, offset int32, length int32) (records []*TransportRecord, total int64, err error) {
	records = make([]*TransportRecord, length)

	db := m.Db().Model(TransportRecord{}).Where("cluster_id = ?", clusterId)
	if recordId != "" {
		db.Where("id = ?", recordId)
	}

	err = db.Count(&total).Offset(int(offset - 1)).Limit(int(length)).Find(&records).Error

	return
}
