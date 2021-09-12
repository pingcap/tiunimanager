package resource

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type HostStatus int32

const (
	HOST_WHATEVER HostStatus = iota - 1
	HOST_ONLINE
	HOST_OFFLINE
	HOST_INUSED
	HOST_EXHAUST
	HOST_DELETED
)

// Status is vailid for Query
func (s HostStatus) IsValid() bool {
	return (s >= HOST_WHATEVER && s <= HOST_DELETED)
}

func (s HostStatus) IsInused() bool {
	return s == HOST_INUSED || s == HOST_EXHAUST
}

func (s HostStatus) IsAvailable() bool {
	return (s == HOST_ONLINE || s == HOST_INUSED)
}

type Host struct {
	ID           string         `json:"hostId" gorm:"PrimaryKey"`
	IP           string         `json:"ip" gorm:"not null"`
	UserName     string         `json:"userName,omitempty" gorm:"size:32"`
	Passwd       string         `json:"passwd,omitempty" gorm:"size:32"`
	HostName     string         `json:"hostName" gorm:"size:255"`
	Status       int32          `json:"status" gorm:"index"` // Host Status, 0 for Online, 1 for offline
	OS           string         `json:"os" gorm:"size:32"`
	Kernel       string         `json:"kernel" gorm:"size:32"`
	CpuCores     int32          `json:"cpuCores"`
	Memory       int32          `json:"memory"`             // Host memory size, Unit:GB
	Spec         string         `json:"spec"`               // Host Spec, init while importing
	Nic          string         `json:"nic" gorm:"size:32"` // Host network type: 1GE or 10GE
	DC           string         `json:"dc" gorm:"size:32"`
	AZ           string         `json:"az" gorm:"index"`
	Rack         string         `json:"rack" gorm:"size:32"`
	Purpose      string         `json:"purpose" gorm:"index"`     // What Purpose is the host used for? [compute/storage/general]
	Performance  string         `json:"performance" gorm:"index"` // Performance type of this host [High/Medium/Low]
	Disks        []Disk         `json:"disks"`
	UsedComputes []UsedCompute  `json:"-"`
	UsedPorts    []UsedPort     `json:"-"`
	CreatedAt    int64          `json:"createTime" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (h Host) IsExhaust() bool {
	diskExaust := true
	for _, disk := range h.Disks {
		if disk.Status == int32(DISK_AVAILABLE) {
			diskExaust = false
			break
		}
	}
	return diskExaust || h.CpuCores == 0 || h.Memory == 0
}

func (h *Host) SetDiskStatus(diskId string, s DiskStatus) {
	for i := range h.Disks {
		if h.Disks[i].ID == diskId {
			h.Disks[i].Status = int32(s)
			break
		}
	}
}

func (h *Host) BeforeCreate(tx *gorm.DB) (err error) {
	err = tx.Where("IP = ? and HOST_NAME = ?", h.IP, h.HostName).First(&Host{}).Error
	if err == nil {
		return status.Errorf(codes.AlreadyExists, "host %s(%s) is existed", h.HostName, h.IP)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.ID = uuid.New().String()
		return nil
	} else {
		return err
	}
}

func (h *Host) BeforeDelete(tx *gorm.DB) (err error) {
	err = tx.Where("ID = ?", h.ID).First(h).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "host %s is not found", h.ID)
		}
	} else {
		if HostStatus(h.Status).IsInused() {
			return status.Errorf(codes.PermissionDenied, "host %s is still in used", h.ID)
		}
	}

	return err
}

func (h *Host) AfterDelete(tx *gorm.DB) (err error) {
	err = tx.Where("host_id = ?", h.ID).Delete(&Disk{}).Error
	if err != nil {
		return
	}
	h.Status = int32(HOST_DELETED)
	err = tx.Model(&h).Update("Status", h.Status).Error
	return
}

func (h *Host) AfterFind(tx *gorm.DB) (err error) {
	err = tx.Find(&(h.Disks), "HOST_ID = ?", h.ID).Error
	return
}
