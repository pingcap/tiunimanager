package models

import (
	"context"
	"testing"
	"time"
)

func TestDAOClusterManager_CreateTransportRecord(t *testing.T) {
	record := &TransportRecord{
		Record: Record{
			ID:       11,
			TenantId: "tenant-cc",
		},
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "path1",
		TenantId:      "tenant-cc",
		Status:        "Running",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}
	id, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	if err != nil {
		t.Errorf("TestDAOClusterManager_CreateTransportRecord failed, %s", err.Error())
		return
	}
	t.Logf("TestDAOClusterManager_CreateTransportRecord success, id: %s", id)
}

func TestDAOClusterManager_UpdateTransportRecord(t *testing.T) {
	record := &TransportRecord{
		Record: Record{
			ID:       22,
			TenantId: "tenant-cc",
		},
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "path1",
		TenantId:      "tenant-cc",
		Status:        "Running",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}
	id, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	if err != nil {
		t.Errorf("TestDAOClusterManager_UpdateTransportRecord create record failed, %s", err.Error())
		return
	}

	err = Dao.ClusterManager().UpdateTransportRecord(context.TODO(), id, record.ClusterId, "Finish", time.Now())
	if err != nil {
		t.Errorf("TestDAOClusterManager_UpdateTransportRecord update record failed, %s", err.Error())
		return
	}
	t.Logf("TestDAOClusterManager_UpdateTransportRecord success")
}

func TestDAOClusterManager_FindTransportRecordById(t *testing.T) {
	record := &TransportRecord{
		Record: Record{
			ID:       33,
			TenantId: "tenant-cc",
		},
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "path1",
		TenantId:      "tenant-cc",
		Status:        "Running",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}
	id, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	if err != nil {
		t.Errorf("TestDAOClusterManager_FindTransportRecordById create record failed, %s", err.Error())
		return
	}

	findRecord, err := Dao.ClusterManager().FindTransportRecordById(context.TODO(), id)
	if err != nil {
		t.Errorf("TestDAOClusterManager_FindTransportRecordById find record failed, %s", err.Error())
		return
	}
	t.Logf("TestDAOClusterManager_FindTransportRecordById success, record: %v", findRecord)
}

func TestDAOClusterManager_ListTransportRecord(t *testing.T) {
	record := &TransportRecord{
		Record: Record{
			ID:       44,
			TenantId: "tenant-cc",
		},
		ClusterId:     "tc-123",
		TransportType: "import",
		FilePath:      "path1",
		TenantId:      "tenant-cc",
		Status:        "Running",
		StartTime:     time.Now(),
		EndTime:       time.Now(),
	}
	_, err := Dao.ClusterManager().CreateTransportRecord(context.TODO(), record)
	if err != nil {
		t.Errorf("TestDAOClusterManager_ListTransportRecord create record failed, %s", err.Error())
		return
	}

	list, total, err := Dao.ClusterManager().ListTransportRecord(context.TODO(), record.ClusterId, "", 0, 10)
	if err != nil {
		t.Errorf("TestDAOClusterManager_ListTransportRecord create record failed, %s", err.Error())
		return
	}
	t.Logf("TestDAOClusterManager_ListTransportRecord success, total: %d, list: %v", total, list)
}
