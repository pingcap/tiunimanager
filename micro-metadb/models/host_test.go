package models

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"

	"github.com/pingcap-inc/tiem/library/common/resource-type"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func TestCreateHost(t *testing.T) {
	h := &resource.Host{
		HostName:     "TestCreateHostRepeated",
		IP:           "111.111.111.111",
		Status:       0,
		OS:           "CentOS",
		Kernel:       "5.0.0",
		FreeCpuCores: 5,
		FreeMemory:   8,
		Nic:          "1GE",
		AZ:           "Zone1",
		Rack:         "3-1",
		Purpose:      "TestCompute",
		Disks: []resource.Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}
	id, _ := Dao.ResourceManager().CreateHost(h)
	defer Dao.ResourceManager().DeleteHost(id)
	type args struct {
		host *resource.Host
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		assert  func(result string) bool
	}{
		{"normal", args{host: &resource.Host{
			HostName:     "TestCreateHost1",
			IP:           "192.168.125.132",
			Status:       0,
			OS:           "CentOS",
			Kernel:       "5.0.0",
			FreeCpuCores: 5,
			FreeMemory:   8,
			Nic:          "1GE",
			AZ:           "Zone1",
			Rack:         "3-1",
			Purpose:      "TestCompute",
			Disks: []resource.Disk{
				{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
			},
		}}, false, func(result string) bool {
			return len(result) > 0
		}},
		{"same name", args{host: &resource.Host{
			HostName:     "TestCreateHostRepeated",
			IP:           "111.111.111.111",
			Status:       0,
			OS:           "CentOS",
			Kernel:       "5.0.0",
			FreeCpuCores: 5,
			FreeMemory:   8,
			Nic:          "1GE",
			AZ:           "Zone1",
			Rack:         "3-1",
			Purpose:      "TestCompute",
			Disks: []resource.Disk{
				{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
			},
		}}, true, func(result string) bool {
			return true
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, err := Dao.ResourceManager().CreateHost(tt.args.host)
			defer Dao.ResourceManager().DeleteHost(gotId)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.assert(gotId) {
				t.Errorf("CreateHost() assert false, gotId == %v", gotId)
			}
		})
	}
}

func TestCreateHostsInBatch(t *testing.T) {
	type args struct {
		hosts []*resource.Host
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(result []string) bool
	}{
		{"normal", args{hosts: []*resource.Host{
			{
				HostName:     "TestCreateHostsInBatch1",
				IP:           "192.168.11.111",
				Status:       0,
				OS:           "CentOS",
				Kernel:       "5.0.0",
				FreeCpuCores: 5,
				FreeMemory:   8,
				Nic:          "1GE",
				AZ:           "Zone1",
				Rack:         "3-1",
				Purpose:      "TestCompute",
				Disks:        []resource.Disk{{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1}},
			},
			{
				HostName:     "TestCreateHostsInBatch2",
				IP:           "192.111.125.111",
				Status:       0,
				OS:           "CentOS",
				Kernel:       "5.0.0",
				FreeCpuCores: 5,
				FreeMemory:   8,
				Nic:          "1GE",
				AZ:           "Zone1",
				Rack:         "3-1",
				Purpose:      "TestCompute",
				Disks:        []resource.Disk{{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1}},
			},
		}}, false, []func(result []string) bool{
			func(result []string) bool { return len(result) == 2 },
			func(result []string) bool { return len(result[0]) > 0 },
		}},
		{"same name", args{hosts: []*resource.Host{
			{
				HostName:     "TestCreateHostsInBatchRepeated",
				IP:           "444.555.666.777",
				Status:       0,
				OS:           "CentOS",
				Kernel:       "5.0.0",
				FreeCpuCores: 5,
				FreeMemory:   8,
				Nic:          "1GE",
				AZ:           "Zone1",
				Rack:         "3-1",
				Purpose:      "TestCompute",
				Disks:        []resource.Disk{{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1}},
			},
			{
				HostName:     "TestCreateHostsInBatchRepeated",
				IP:           "444.555.666.777",
				Status:       0,
				OS:           "CentOS",
				Kernel:       "5.0.0",
				FreeCpuCores: 5,
				FreeMemory:   8,
				Nic:          "1GE",
				AZ:           "Zone1",
				Rack:         "3-1",
				Purpose:      "TestCompute",
				Disks:        []resource.Disk{{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1}},
			},
		}}, true, []func(result []string) bool{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIds, err := Dao.ResourceManager().CreateHostsInBatch(tt.args.hosts)
			defer Dao.ResourceManager().DeleteHostsInBatch(gotIds)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateHostsInBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(gotIds) {
					t.Errorf("CreateHostsInBatch() assert false, index = %v, gotIds = %v", i, gotIds)
				}
			}
		})
	}
}

func TestDeleteHost(t *testing.T) {
	h := &resource.Host{
		HostName:     "TestDeleteHost1",
		IP:           "192.99.999.132",
		Status:       0,
		OS:           "CentOS",
		Kernel:       "5.0.0",
		FreeCpuCores: 5,
		FreeMemory:   8,
		Nic:          "1GE",
		AZ:           "Zone1",
		Rack:         "3-1",
		Purpose:      "TestCompute",
		Disks: []resource.Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}
	id, _ := Dao.ResourceManager().CreateHost(h)
	defer Dao.ResourceManager().DeleteHost(id)
	type args struct {
		hostId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{id}, false},
		{"deleted", args{id}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Dao.ResourceManager().DeleteHost(tt.args.hostId); (err != nil) != tt.wantErr {
				t.Errorf("DeleteHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteHostsInBatch(t *testing.T) {
	h := &resource.Host{
		HostName:     "主机1",
		IP:           "111.99.999.132",
		Status:       0,
		OS:           "CentOS",
		Kernel:       "5.0.0",
		FreeCpuCores: 5,
		FreeMemory:   8,
		Nic:          "1GE",
		AZ:           "Zone1",
		Rack:         "3-1",
		Purpose:      "TestCompute",
		Disks: []resource.Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}
	id1, _ := Dao.ResourceManager().CreateHost(h)
	defer Dao.ResourceManager().DeleteHost(id1)
	h2 := &resource.Host{
		HostName:     "主机1",
		IP:           "222.99.999.132",
		Status:       0,
		OS:           "CentOS",
		Kernel:       "5.0.0",
		FreeCpuCores: 5,
		FreeMemory:   8,
		Nic:          "1GE",
		AZ:           "Zone1",
		Rack:         "3-1",
		Purpose:      "TestCompute",
		Disks: []resource.Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}
	id2, _ := Dao.ResourceManager().CreateHost(h2)
	defer Dao.ResourceManager().DeleteHost(id2)

	type args struct {
		hostIds []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{hostIds: []string{id1, id2}}, false},
		{"normal", args{hostIds: []string{id1, id2}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Dao.ResourceManager().DeleteHostsInBatch(tt.args.hostIds); (err != nil) != tt.wantErr {
				t.Errorf("DeleteHostsInBatch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDiskStatus_IsAvailable(t *testing.T) {
	tests := []struct {
		name string
		s    resource.DiskStatus
		want bool
	}{
		{"normal", resource.DISK_AVAILABLE, true},
		{"want false", 999, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiskStatus_IsInused(t *testing.T) {
	tests := []struct {
		name string
		s    resource.DiskStatus
		want bool
	}{
		{"normal", resource.DISK_INUSED, true},
		{"want false", 999, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.IsInused(); got != tt.want {
				t.Errorf("IsInused() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDisk_BeforeCreate(t *testing.T) {
	type fields struct {
		ID        string
		HostId    string
		Name      string
		Capacity  int32
		Path      string
		Status    int32
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &resource.Disk{
				ID:        tt.fields.ID,
				HostId:    tt.fields.HostId,
				Name:      tt.fields.Name,
				Capacity:  tt.fields.Capacity,
				Path:      tt.fields.Path,
				Status:    tt.fields.Status,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
				DeletedAt: tt.fields.DeletedAt,
			}
			if err := d.BeforeCreate(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("BeforeCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFindHostById(t *testing.T) {
	h := &resource.Host{
		HostName:     "主机1",
		IP:           "876.111.111.111",
		Status:       0,
		OS:           "CentOS",
		Kernel:       "5.0.0",
		FreeCpuCores: 5,
		FreeMemory:   8,
		Nic:          "1GE",
		AZ:           "Zone1",
		Rack:         "3-1",
		Purpose:      "TestCompute",
		Disks: []resource.Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}
	id1, _ := Dao.ResourceManager().CreateHost(h)
	defer Dao.ResourceManager().DeleteHost(id1)
	type args struct {
		hostId string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(result *resource.Host) bool
	}{
		{"normal", args{hostId: id1}, false, []func(result *resource.Host) bool{
			func(result *resource.Host) bool { return id1 == result.ID },
			func(result *resource.Host) bool { return h.IP == result.IP },
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Dao.ResourceManager().FindHostById(tt.args.hostId)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindHostById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(got) {
					t.Errorf("FindHostById() assert false, index = %v, got =  %v", i, got)
				}
			}
		})
	}
}

func TestGetFailureDomain(t *testing.T) {
	type args struct {
		domain string
	}
	tests := []struct {
		name    string
		args    args
		wantRes []FailureDomainResource
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRes, err := Dao.ResourceManager().GetFailureDomain(tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFailureDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRes, tt.wantRes) {
				t.Errorf("GetFailureDomain() gotRes = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}

func TestHostStatus_IsAvailable(t *testing.T) {
	tests := []struct {
		name string
		h    resource.Host
		want bool
	}{
		{"normal_online", resource.Host{Status: int32(resource.HOST_ONLINE), Stat: resource.HOST_LOADLESS}, true},
		{"normal_inused", resource.Host{Status: int32(resource.HOST_ONLINE), Stat: resource.HOST_INUSED}, true},
		{"want false", resource.Host{Status: int32(resource.HOST_OFFLINE), Stat: resource.HOST_INUSED}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostStatus_IsInused(t *testing.T) {
	tests := []struct {
		name string
		h    resource.Host
		want bool
	}{
		{"normal_inused", resource.Host{Status: int32(resource.HOST_ONLINE), Stat: resource.HOST_INUSED}, true},
		{"want false", resource.Host{Status: int32(resource.HOST_ONLINE), Stat: resource.HOST_LOADLESS}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.IsInused(); got != tt.want {
				t.Errorf("IsInused() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHostTableName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"normal", "hosts"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TABLE_NAME_HOST; got != tt.want {
				t.Errorf("HostTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHost_AfterDelete(t *testing.T) {
	type fields struct {
		ID        string
		IP        string
		HostName  string
		Status    int32
		OS        string
		Kernel    string
		CpuCores  int32
		Memory    int32
		Spec      string
		Nic       string
		Region    string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []resource.Disk
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &resource.Host{
				ID:           tt.fields.ID,
				IP:           tt.fields.IP,
				HostName:     tt.fields.HostName,
				Status:       tt.fields.Status,
				OS:           tt.fields.OS,
				Kernel:       tt.fields.Kernel,
				FreeCpuCores: tt.fields.CpuCores,
				FreeMemory:   tt.fields.Memory,
				Spec:         tt.fields.Spec,
				Nic:          tt.fields.Nic,
				Region:       tt.fields.Region,
				AZ:           tt.fields.AZ,
				Rack:         tt.fields.Rack,
				Purpose:      tt.fields.Purpose,
				Disks:        tt.fields.Disks,
				CreatedAt:    tt.fields.CreatedAt,
				UpdatedAt:    tt.fields.UpdatedAt,
				DeletedAt:    tt.fields.DeletedAt,
			}
			if err := h.AfterDelete(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("AfterDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHost_AfterFind(t *testing.T) {
	type fields struct {
		ID        string
		IP        string
		HostName  string
		Status    int32
		OS        string
		Kernel    string
		CpuCores  int32
		Memory    int32
		Spec      string
		Nic       string
		Region    string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []resource.Disk
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &resource.Host{
				ID:           tt.fields.ID,
				IP:           tt.fields.IP,
				HostName:     tt.fields.HostName,
				Status:       tt.fields.Status,
				OS:           tt.fields.OS,
				Kernel:       tt.fields.Kernel,
				FreeCpuCores: tt.fields.CpuCores,
				FreeMemory:   tt.fields.Memory,
				Spec:         tt.fields.Spec,
				Nic:          tt.fields.Nic,
				Region:       tt.fields.Region,
				AZ:           tt.fields.AZ,
				Rack:         tt.fields.Rack,
				Purpose:      tt.fields.Purpose,
				Disks:        tt.fields.Disks,
				CreatedAt:    tt.fields.CreatedAt,
				UpdatedAt:    tt.fields.UpdatedAt,
				DeletedAt:    tt.fields.DeletedAt,
			}
			if err := h.AfterFind(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("AfterFind() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHost_BeforeCreate(t *testing.T) {
	type fields struct {
		ID        string
		IP        string
		HostName  string
		Status    int32
		OS        string
		Kernel    string
		CpuCores  int32
		Memory    int32
		Spec      string
		Nic       string
		Region    string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []resource.Disk
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &resource.Host{
				ID:           tt.fields.ID,
				IP:           tt.fields.IP,
				HostName:     tt.fields.HostName,
				Status:       tt.fields.Status,
				OS:           tt.fields.OS,
				Kernel:       tt.fields.Kernel,
				FreeCpuCores: tt.fields.CpuCores,
				FreeMemory:   tt.fields.Memory,
				Spec:         tt.fields.Spec,
				Nic:          tt.fields.Nic,
				Region:       tt.fields.Region,
				AZ:           tt.fields.AZ,
				Rack:         tt.fields.Rack,
				Purpose:      tt.fields.Purpose,
				Disks:        tt.fields.Disks,
				CreatedAt:    tt.fields.CreatedAt,
				UpdatedAt:    tt.fields.UpdatedAt,
				DeletedAt:    tt.fields.DeletedAt,
			}
			if err := h.BeforeCreate(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("BeforeCreate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHost_BeforeDelete(t *testing.T) {
	type fields struct {
		ID        string
		IP        string
		HostName  string
		Status    int32
		OS        string
		Kernel    string
		CpuCores  int32
		Memory    int32
		Spec      string
		Nic       string
		Region    string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []resource.Disk
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
	type args struct {
		tx *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &resource.Host{
				ID:           tt.fields.ID,
				IP:           tt.fields.IP,
				HostName:     tt.fields.HostName,
				Status:       tt.fields.Status,
				OS:           tt.fields.OS,
				Kernel:       tt.fields.Kernel,
				FreeCpuCores: tt.fields.CpuCores,
				FreeMemory:   tt.fields.Memory,
				Spec:         tt.fields.Spec,
				Nic:          tt.fields.Nic,
				Region:       tt.fields.Region,
				AZ:           tt.fields.AZ,
				Rack:         tt.fields.Rack,
				Purpose:      tt.fields.Purpose,
				Disks:        tt.fields.Disks,
				CreatedAt:    tt.fields.CreatedAt,
				UpdatedAt:    tt.fields.UpdatedAt,
				DeletedAt:    tt.fields.DeletedAt,
			}
			if err := h.BeforeDelete(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("BeforeDelete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHost_IsExhaust(t *testing.T) {
	type fields struct {
		ID        string
		IP        string
		HostName  string
		Status    int32
		OS        string
		Kernel    string
		CpuCores  int32
		Memory    int32
		Spec      string
		Nic       string
		Region    string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []resource.Disk
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"normal", fields{CpuCores: 4, Memory: 8, Disks: []resource.Disk{{Status: int32(resource.DISK_AVAILABLE)}}}, false},
		{"without_disk", fields{CpuCores: 4, Memory: 8}, true},
		{"without_cpu", fields{CpuCores: 0, Memory: 8, Disks: []resource.Disk{{Status: int32(resource.DISK_AVAILABLE)}}}, true},
		{"without_momery", fields{CpuCores: 4, Memory: 0, Disks: []resource.Disk{{Status: int32(resource.DISK_AVAILABLE)}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := resource.Host{
				ID:           tt.fields.ID,
				IP:           tt.fields.IP,
				HostName:     tt.fields.HostName,
				Status:       tt.fields.Status,
				OS:           tt.fields.OS,
				Kernel:       tt.fields.Kernel,
				FreeCpuCores: tt.fields.CpuCores,
				FreeMemory:   tt.fields.Memory,
				Spec:         tt.fields.Spec,
				Nic:          tt.fields.Nic,
				Region:       tt.fields.Region,
				AZ:           tt.fields.AZ,
				Rack:         tt.fields.Rack,
				Purpose:      tt.fields.Purpose,
				Disks:        tt.fields.Disks,
				CreatedAt:    tt.fields.CreatedAt,
				UpdatedAt:    tt.fields.UpdatedAt,
				DeletedAt:    tt.fields.DeletedAt,
			}
			if got := h.IsExhaust(); got != tt.want {
				t.Errorf("IsExhaust() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHost_SetDiskStatus(t *testing.T) {
	type fields struct {
		Disks []resource.Disk
	}
	type args struct {
		diskId string
		s      resource.DiskStatus
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wanted resource.DiskStatus
	}{
		{"normal", fields{[]resource.Disk{{ID: "fake1", Status: int32(resource.DISK_AVAILABLE)}, {ID: "fake2", Status: int32(resource.DISK_AVAILABLE)}}}, args{"fake2", resource.DISK_INUSED}, resource.DISK_INUSED},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &resource.Host{
				Disks: tt.fields.Disks,
			}
			h.SetDiskStatus(tt.args.diskId, tt.args.s)
			assert.Equal(t, int32(tt.wanted), h.Disks[1].Status)
		})
	}
}

func CreateTestHost(region, zone, rack, hostName, ip, purpose, diskType string, freeCpuCores, freeMemory, availableDiskCount int32) (id string, err error) {
	h := &resource.Host{
		HostName:     hostName,
		IP:           ip,
		UserName:     "root",
		Passwd:       "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:       0,
		Stat:         0,
		Arch:         "X86",
		OS:           "CentOS",
		Kernel:       "5.0.0",
		FreeCpuCores: freeCpuCores,
		FreeMemory:   freeMemory,
		Nic:          "1GE",
		Region:       region,
		AZ:           zone,
		Rack:         rack,
		Purpose:      purpose,
		DiskType:     diskType,
		Disks: []resource.Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: 1, Type: diskType},
		},
	}
	for i := 0; i < int(availableDiskCount); i++ {
		deviceName := fmt.Sprintf("sd%c", 'b'+i)
		path := fmt.Sprintf("/mnt%d", i+1)
		h.Disks = append(h.Disks, resource.Disk{
			Name:     deviceName,
			Path:     path,
			Capacity: 256,
			Status:   0,
			Type:     diskType,
		})
	}
	return Dao.ResourceManager().CreateHost(h)
}
func TestListHosts(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName1", "111.121.999.132", "TestCompute", string(resource.SSD), 17, 64, 1)
	id2, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName2", "222.121.999.132", "TestCompute", string(resource.SSD), 16, 64, 0)
	id3, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName3", "333.121.999.132", "whatever", string(resource.SSD), 15, 64, 0)
	defer Dao.ResourceManager().DeleteHost(id1)
	defer Dao.ResourceManager().DeleteHost(id2)
	defer Dao.ResourceManager().DeleteHost(id3)

	type args struct {
		req ListHostReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(a args, result []resource.Host) bool
	}{
		{"normal", args{req: ListHostReq{
			Status:  0,
			Purpose: "TestCompute",
			Offset:  0,
			Limit:   2,
		}}, false, []func(a args, result []resource.Host) bool{
			func(a args, result []resource.Host) bool { return len(result) == 2 },
			func(a args, result []resource.Host) bool { return result[1].Purpose == "TestCompute" },
		}},
		{"offset", args{req: ListHostReq{
			Status:  0,
			Purpose: "TestCompute",
			Offset:  1,
			Limit:   2,
		}}, false, []func(a args, result []resource.Host) bool{
			func(a args, result []resource.Host) bool { return len(result) == 1 },
		}},
		{"without Purpose", args{req: ListHostReq{
			Status: resource.HOST_ONLINE,
			Offset: 0,
			Limit:  5,
		}}, false, []func(a args, result []resource.Host) bool{
			func(a args, result []resource.Host) bool { return len(result) >= 2 },
		}},
		{"without status", args{req: ListHostReq{
			Status:  resource.HOST_WHATEVER,
			Purpose: "TestCompute",
			Offset:  0,
			Limit:   5,
		}}, false, []func(a args, result []resource.Host) bool{
			func(a args, result []resource.Host) bool { return len(result) == 2 },
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHosts, err := Dao.ResourceManager().ListHosts(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListHosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(tt.args, gotHosts) {
					t.Errorf("ListHosts() assert false, index = %v, gotHosts %v", i, gotHosts)
				}
			}
		})
	}
}

func TestAllocHosts_3Hosts(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName1", "474.111.111.111", string(resource.General), string(resource.SSD), 17, 64, 1)
	id2, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName2", "474.111.111.112", string(resource.General), string(resource.SSD), 16, 64, 2)
	id3, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName3", "474.111.111.113", string(resource.General), string(resource.SSD), 15, 64, 1)
	// Host Status should be inused or exhausted, so delete would failed
	defer Dao.ResourceManager().DeleteHost(id1)
	defer Dao.ResourceManager().DeleteHost(id2)
	defer Dao.ResourceManager().DeleteHost(id3)

	var m AllocReqs = make(map[string][]*HostAllocReq)
	m["PD"] = append(m["PD"], &HostAllocReq{
		FailureDomain: "Zone1",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["TiDB"] = append(m["TiDB"], &HostAllocReq{
		FailureDomain: "Zone1",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["TiKV"] = append(m["TiKV"], &HostAllocReq{
		FailureDomain: "Zone1",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	type args struct {
		req AllocReqs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{
			req: m,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := Dao.ResourceManager().AllocHosts(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			pdIp := rsp["PD"][0].Ip
			tidbIp := rsp["TiDB"][0].Ip
			tikvIp := rsp["TiDB"][0].Ip
			assert.True(t, pdIp == "474.111.111.111" || pdIp == "474.111.111.112" || pdIp == "474.111.111.113")
			assert.True(t, tidbIp == "474.111.111.111" || tidbIp == "474.111.111.112" || tidbIp == "474.111.111.113")
			assert.True(t, tikvIp == "474.111.111.111" || tikvIp == "474.111.111.112" || tikvIp == "474.111.111.113")
			assert.Equal(t, 4, rsp["PD"][0].CpuCores)
			assert.Equal(t, 8, rsp["PD"][0].Memory)
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "474.111.111.111")
			assert.Equal(t, int32(17-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_EXHAUST))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.112")
			assert.Equal(t, int32(16-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.113")
			assert.Equal(t, int32(15-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_EXHAUST))
		})
	}
}

func TestAllocHosts_1Host(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone99", "3-1", "HostName1", "192.168.56.99", string(resource.General), string(resource.SSD), 17, 64, 3)
	// Host Status should be inused or exhausted, so delete would failed
	defer Dao.ResourceManager().DeleteHost(id1)

	var m AllocReqs = make(map[string][]*HostAllocReq)
	m["PD"] = append(m["PD"], &HostAllocReq{
		FailureDomain: "Zone99",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["TiDB"] = append(m["TiDB"], &HostAllocReq{
		FailureDomain: "Zone99",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["TiKV"] = append(m["TiKV"], &HostAllocReq{
		FailureDomain: "Zone99",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	type args struct {
		req AllocReqs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{
			req: m,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := Dao.ResourceManager().AllocHosts(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			pdIp := rsp["PD"][0].Ip
			tidbIp := rsp["TiDB"][0].Ip
			tikvIp := rsp["TiKV"][0].Ip
			pdDisk := rsp["PD"][0].DiskName
			tidbDisk := rsp["TiDB"][0].DiskName
			tikvDisk := rsp["TiKV"][0].DiskName
			assert.True(t, pdIp == "192.168.56.99" && tidbIp == "192.168.56.99" && tikvIp == "192.168.56.99")
			assert.True(t, pdDisk != tidbDisk && tidbDisk != tikvDisk && tikvDisk != pdDisk)
			assert.True(t, pdDisk == "sdb" || pdDisk == "sdc" || pdDisk == "sdd")
			assert.True(t, tidbDisk == "sdb" || tidbDisk == "sdc" || tidbDisk == "sdd")
			assert.True(t, tikvDisk == "sdb" || tikvDisk == "sdc" || tikvDisk == "sdd")
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "192.168.56.99")
			assert.Equal(t, int32(17-4-4-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_EXHAUST))

		})
	}
}

func TestAllocHosts_1Host_NotEnough(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone100", "3-1", "HostName1", "192.168.56.100", string(resource.General), string(resource.SSD), 17, 64, 3)
	defer Dao.ResourceManager().DeleteHost(id1)

	var m AllocReqs = make(map[string][]*HostAllocReq)
	m["PD"] = append(m["PD"], &HostAllocReq{
		FailureDomain: "Zone100",
		CpuCores:      4,
		Memory:        8,
		Count:         2,
	})
	m["TiDB"] = append(m["TiDB"], &HostAllocReq{
		FailureDomain: "Zone100",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["TiKV"] = append(m["TiKV"], &HostAllocReq{
		FailureDomain: "Zone100",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	type args struct {
		req AllocReqs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{
			req: m,
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Dao.ResourceManager().AllocHosts(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			s, ok := status.FromError(err)
			assert.True(t, true, ok)
			assert.Equal(t, codes.Internal, s.Code())
		})
	}
}

func TestAllocResources_1Requirement_3Hosts(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName1", "474.111.111.108", string(resource.General), string(resource.SSD), 17, 64, 1)
	id2, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName2", "474.111.111.109", string(resource.General), string(resource.SSD), 16, 64, 2)
	id3, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName3", "474.111.111.110", string(resource.General), string(resource.SSD), 15, 64, 1)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed

	//	defer Dao.ResourceManager().DeleteHost(id1)
	//	defer Dao.ResourceManager().DeleteHost(id2)
	//	defer Dao.ResourceManager().DeleteHost(id3)

	loc := new(dbpb.DBLocation)
	loc.Region = "Region1"
	loc.Zone = "Zone1"
	filter := new(dbpb.DBFilter)
	filter.Arch = string(resource.X86)
	filter.DiskType = string(resource.SSD)
	filter.Purpose = string(resource.General)
	require := new(dbpb.DBRequirement)
	require.ComputeReq = new(dbpb.DBComputeRequirement)
	require.ComputeReq.CpuCores = 4
	require.ComputeReq.Memory = 8
	require.DiskReq = new(dbpb.DBDiskRequirement)
	require.DiskReq.Capacity = 256
	require.DiskReq.DiskType = string(resource.SSD)
	require.DiskReq.NeedDisk = true
	require.PortReq = append(require.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10010,
		PortCnt: 5,
	})

	var test_req dbpb.DBAllocRequest
	test_req.Applicant = new(dbpb.DBApplicant)
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location:   loc,
		HostFilter: filter,
		Strategy:   int32(resource.RandomRack),
		Require:    require,
		Count:      3,
	})

	type args struct {
		request *dbpb.DBAllocRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{
			request: &test_req,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := Dao.ResourceManager().AllocResources(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, 3, len(rsp.Results))
			assert.Equal(t, int32(0), rsp.Results[0].Reqseq)
			assert.Equal(t, int32(0), rsp.Results[1].Reqseq)
			assert.Equal(t, int32(0), rsp.Results[2].Reqseq)
			assert.Equal(t, int32(10003), rsp.Results[0].PortRes[0].Ports[3])
			assert.True(t, rsp.Results[0].HostIp != rsp.Results[1].HostIp)
			assert.True(t, rsp.Results[0].HostIp != rsp.Results[2].HostIp)
			assert.True(t, rsp.Results[2].HostIp != rsp.Results[1].HostIp)
			assert.True(t, rsp.Results[0].HostIp == "474.111.111.108" || rsp.Results[0].HostIp == "474.111.111.109" || rsp.Results[0].HostIp == "474.111.111.110")
			assert.True(t, rsp.Results[1].HostIp == "474.111.111.108" || rsp.Results[1].HostIp == "474.111.111.109" || rsp.Results[1].HostIp == "474.111.111.110")
			assert.True(t, rsp.Results[2].HostIp == "474.111.111.108" || rsp.Results[2].HostIp == "474.111.111.109" || rsp.Results[2].HostIp == "474.111.111.110")
			assert.Equal(t, int32(4), rsp.Results[0].ComputeRes.CpuCores)
			assert.Equal(t, int32(8), rsp.Results[0].ComputeRes.Memory)
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "474.111.111.108")
			assert.Equal(t, int32(17-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_EXHAUST))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.109")
			assert.Equal(t, int32(16-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.110")
			assert.Equal(t, int32(15-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_EXHAUST))
		})
	}
}

func TestAllocResources_3Requirement_3Hosts(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone2", "3-1", "HostName1", "474.111.111.114", string(resource.General), string(resource.SSD), 17, 64, 1)
	id2, _ := CreateTestHost("Region1", "Zone2", "3-1", "HostName2", "474.111.111.115", string(resource.General), string(resource.SSD), 16, 64, 2)
	id3, _ := CreateTestHost("Region1", "Zone2", "3-1", "HostName3", "474.111.111.116", string(resource.Storage), string(resource.SSD), 15, 64, 1)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed

	//	defer Dao.ResourceManager().DeleteHost(id1)
	//	defer Dao.ResourceManager().DeleteHost(id2)
	//	defer Dao.ResourceManager().DeleteHost(id3)

	loc := new(dbpb.DBLocation)
	loc.Region = "Region1"
	loc.Zone = "Zone2"
	filter1 := new(dbpb.DBFilter)
	filter1.Arch = string(resource.X86)
	filter1.DiskType = string(resource.SSD)
	filter1.Purpose = string(resource.General)
	filter2 := new(dbpb.DBFilter)
	filter2.Arch = string(resource.X86)
	filter2.DiskType = string(resource.SSD)
	filter2.Purpose = string(resource.Storage)
	require := new(dbpb.DBRequirement)
	require.ComputeReq = new(dbpb.DBComputeRequirement)
	require.ComputeReq.CpuCores = 4
	require.ComputeReq.Memory = 8
	require.DiskReq = new(dbpb.DBDiskRequirement)
	require.DiskReq.Capacity = 256
	require.DiskReq.DiskType = string(resource.SSD)
	require.DiskReq.NeedDisk = true
	require.PortReq = append(require.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10010,
		PortCnt: 5,
	})

	var test_req dbpb.DBAllocRequest
	test_req.Applicant = new(dbpb.DBApplicant)
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   int32(resource.RandomRack),
		Require:    require,
		Count:      2,
	})
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location:   loc,
		HostFilter: filter2,
		Strategy:   int32(resource.RandomRack),
		Require:    require,
		Count:      1,
	})

	type args struct {
		request *dbpb.DBAllocRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{
			request: &test_req,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := Dao.ResourceManager().AllocResources(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, 3, len(rsp.Results))
			assert.Equal(t, int32(0), rsp.Results[0].Reqseq)
			assert.Equal(t, int32(0), rsp.Results[1].Reqseq)
			assert.Equal(t, int32(1), rsp.Results[2].Reqseq)
			assert.Equal(t, int32(10003), rsp.Results[0].PortRes[0].Ports[3])
			assert.True(t, rsp.Results[0].HostIp != rsp.Results[1].HostIp)
			assert.True(t, rsp.Results[0].HostIp == "474.111.111.114" || rsp.Results[0].HostIp == "474.111.111.115")
			assert.True(t, rsp.Results[1].HostIp == "474.111.111.114" || rsp.Results[1].HostIp == "474.111.111.115")
			assert.True(t, rsp.Results[2].HostIp == "474.111.111.116")
			assert.Equal(t, int32(4), rsp.Results[0].ComputeRes.CpuCores)
			assert.Equal(t, int32(8), rsp.Results[0].ComputeRes.Memory)
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "474.111.111.114")
			assert.Equal(t, int32(17-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_EXHAUST))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.115")
			assert.Equal(t, int32(16-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.116")
			assert.Equal(t, int32(15-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_EXHAUST))
			//var usedPorts []int32
			var usedPorts []resource.UsedPort
			MetaDB.Order("port").Model(&resource.UsedPort{}).Where("host_id = ?", host3.ID).Scan(&usedPorts)
			assert.Equal(t, 5, len(usedPorts))
			assert.Equal(t, test_req.Applicant.HolderId, usedPorts[1].HolderId)
			assert.Equal(t, test_req.Applicant.RequestId, usedPorts[2].RequestId)
			for i := 0; i < 5; i++ {
				assert.Equal(t, int32(10000+i), usedPorts[i].Port)
			}
		})
	}
}

func TestAllocResources_3RequestsInBatch_3Hosts(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone3", "3-1", "HostName1", "474.111.111.117", string(resource.General), string(resource.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Region1", "Zone3", "3-1", "HostName2", "474.111.111.118", string(resource.General), string(resource.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Region1", "Zone3", "3-1", "HostName3", "474.111.111.119", string(resource.General), string(resource.SSD), 15, 64, 3)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed
	/*
		defer Dao.ResourceManager().DeleteHost(id1)
		defer Dao.ResourceManager().DeleteHost(id2)
		defer Dao.ResourceManager().DeleteHost(id3)
	*/

	loc := new(dbpb.DBLocation)
	loc.Region = "Region1"
	loc.Zone = "Zone3"
	filter1 := new(dbpb.DBFilter)
	filter1.DiskType = string(resource.SSD)
	filter1.Purpose = string(resource.General)
	filter1.Arch = string(resource.X86)
	require := new(dbpb.DBRequirement)
	require.ComputeReq = new(dbpb.DBComputeRequirement)
	require.ComputeReq.CpuCores = 4
	require.ComputeReq.Memory = 8
	require.DiskReq = new(dbpb.DBDiskRequirement)
	require.DiskReq.Capacity = 256
	require.DiskReq.DiskType = string(resource.SSD)
	require.DiskReq.NeedDisk = true
	require.PortReq = append(require.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req dbpb.DBAllocRequest
	test_req.Applicant = new(dbpb.DBApplicant)
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   int32(resource.RandomRack),
		Require:    require,
		Count:      3,
	})

	var batchReq dbpb.DBBatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	type args struct {
		request *dbpb.DBBatchAllocRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{
			request: &batchReq,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := Dao.ResourceManager().AllocResourcesInBatch(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, 3, len(rsp.BatchResults))
			assert.Equal(t, 3, len(rsp.BatchResults[0].Results))
			assert.Equal(t, int32(0), rsp.BatchResults[0].Results[2].Reqseq)
			assert.Equal(t, int32(10003), rsp.BatchResults[0].Results[0].PortRes[0].Ports[3])
			assert.True(t, rsp.BatchResults[0].Results[0].HostIp != rsp.BatchResults[0].Results[1].HostIp)
			assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "474.111.111.117" || rsp.BatchResults[0].Results[0].HostIp == "474.111.111.118" || rsp.BatchResults[0].Results[0].HostIp == "474.111.111.119")
			assert.True(t, rsp.BatchResults[0].Results[1].HostIp == "474.111.111.117" || rsp.BatchResults[0].Results[1].HostIp == "474.111.111.118" || rsp.BatchResults[0].Results[1].HostIp == "474.111.111.119")
			assert.True(t, rsp.BatchResults[0].Results[2].HostIp == "474.111.111.117" || rsp.BatchResults[0].Results[2].HostIp == "474.111.111.118" || rsp.BatchResults[0].Results[2].HostIp == "474.111.111.119")
			assert.Equal(t, int32(4), rsp.BatchResults[0].Results[0].ComputeRes.CpuCores)
			assert.Equal(t, int32(8), rsp.BatchResults[0].Results[0].ComputeRes.Memory)
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "474.111.111.117")
			assert.Equal(t, int32(17-4-4-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_EXHAUST))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.118")
			assert.Equal(t, int32(16-4-4-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_EXHAUST))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.119")
			assert.Equal(t, int32(15-4-4-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_EXHAUST))
			//var usedPorts []int32
			var usedPorts []resource.UsedPort
			MetaDB.Order("port").Model(&resource.UsedPort{}).Where("host_id = ?", host3.ID).Scan(&usedPorts)
			assert.Equal(t, 15, len(usedPorts))
			assert.Equal(t, test_req.Applicant.HolderId, usedPorts[1].HolderId)
			assert.Equal(t, test_req.Applicant.RequestId, usedPorts[2].RequestId)
			for i := 0; i < 15; i++ {
				assert.Equal(t, int32(10000+i), usedPorts[i].Port)
			}
		})
	}
}

func TestAllocResources_3RequestsInBatch_3Hosts_No_Disk(t *testing.T) {
	id1, _ := CreateTestHost("Region111", "Zone3", "3-1", "HostName1", "474.112.111.117", string(resource.General), string(resource.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Region111", "Zone3", "3-1", "HostName2", "474.112.111.118", string(resource.General), string(resource.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Region111", "Zone3", "3-1", "HostName3", "474.112.111.119", string(resource.General), string(resource.SSD), 15, 64, 3)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed
	/*
		defer Dao.ResourceManager().DeleteHost(id1)
		defer Dao.ResourceManager().DeleteHost(id2)
		defer Dao.ResourceManager().DeleteHost(id3)
	*/

	loc := new(dbpb.DBLocation)
	loc.Region = "Region111"
	loc.Zone = "Zone3"
	filter1 := new(dbpb.DBFilter)
	filter1.DiskType = string(resource.SSD)
	filter1.Purpose = string(resource.General)
	filter1.Arch = string(resource.X86)
	require := new(dbpb.DBRequirement)
	require.ComputeReq = new(dbpb.DBComputeRequirement)
	require.ComputeReq.CpuCores = 4
	require.ComputeReq.Memory = 8
	require.DiskReq = new(dbpb.DBDiskRequirement)
	require.DiskReq.NeedDisk = false
	require.PortReq = append(require.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req dbpb.DBAllocRequest
	test_req.Applicant = new(dbpb.DBApplicant)
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location:   loc,
		HostFilter: filter1,
		Strategy:   int32(resource.RandomRack),
		Require:    require,
		Count:      3,
	})

	var batchReq dbpb.DBBatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	assert.Equal(t, 3, len(batchReq.BatchRequests))

	type args struct {
		request *dbpb.DBBatchAllocRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{
			request: &batchReq,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := Dao.ResourceManager().AllocResourcesInBatch(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, 3, len(rsp.BatchResults))
			assert.Equal(t, 3, len(rsp.BatchResults[0].Results))
			assert.Equal(t, int32(0), rsp.BatchResults[0].Results[2].Reqseq)
			assert.Equal(t, int32(10003), rsp.BatchResults[0].Results[0].PortRes[0].Ports[3])
			assert.True(t, rsp.BatchResults[0].Results[0].HostIp != rsp.BatchResults[0].Results[1].HostIp)
			assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "474.112.111.117" || rsp.BatchResults[0].Results[0].HostIp == "474.112.111.118" || rsp.BatchResults[0].Results[0].HostIp == "474.112.111.119")
			assert.True(t, rsp.BatchResults[0].Results[1].HostIp == "474.112.111.117" || rsp.BatchResults[0].Results[1].HostIp == "474.112.111.118" || rsp.BatchResults[0].Results[1].HostIp == "474.112.111.119")
			assert.True(t, rsp.BatchResults[0].Results[2].HostIp == "474.112.111.117" || rsp.BatchResults[0].Results[2].HostIp == "474.112.111.118" || rsp.BatchResults[0].Results[2].HostIp == "474.112.111.119")
			assert.Equal(t, int32(4), rsp.BatchResults[0].Results[0].ComputeRes.CpuCores)
			assert.Equal(t, int32(8), rsp.BatchResults[0].Results[0].ComputeRes.Memory)
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "474.112.111.117")
			assert.Equal(t, int32(17-4-4-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			assert.True(t, host.Disks[1].Status == int32(resource.DISK_AVAILABLE) || host.Disks[2].Status == int32(resource.DISK_AVAILABLE) || host.Disks[3].Status == int32(resource.DISK_AVAILABLE))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.112.111.118")
			assert.Equal(t, int32(16-4-4-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			assert.True(t, host2.Disks[1].Status == int32(resource.DISK_AVAILABLE) || host2.Disks[2].Status == int32(resource.DISK_AVAILABLE) || host2.Disks[3].Status == int32(resource.DISK_AVAILABLE))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.112.111.119")
			assert.Equal(t, int32(15-4-4-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_INUSED))
			assert.True(t, host3.Disks[1].Status == int32(resource.DISK_AVAILABLE) || host3.Disks[2].Status == int32(resource.DISK_AVAILABLE) || host3.Disks[3].Status == int32(resource.DISK_AVAILABLE))
			//var usedPorts []int32
			var usedPorts []resource.UsedPort
			MetaDB.Order("port").Model(&resource.UsedPort{}).Where("host_id = ?", host3.ID).Scan(&usedPorts)
			assert.Equal(t, 15, len(usedPorts))
			assert.Equal(t, test_req.Applicant.HolderId, usedPorts[1].HolderId)
			assert.Equal(t, test_req.Applicant.RequestId, usedPorts[2].RequestId)
			for i := 0; i < 15; i++ {
				assert.Equal(t, int32(10000+i), usedPorts[i].Port)
			}
		})
	}
}

func TestAllocResources_3RequestsInBatch_SpecifyHost_Strategy(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone4", "3-1", "HostName1", "474.111.111.127", string(resource.General), string(resource.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Region1", "Zone4", "3-1", "HostName2", "474.111.111.128", string(resource.General), string(resource.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Region1", "Zone4", "3-1", "HostName3", "474.111.111.129", string(resource.General), string(resource.SSD), 15, 64, 3)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed
	/*
		defer Dao.ResourceManager().DeleteHost(id1)
		defer Dao.ResourceManager().DeleteHost(id2)
		defer Dao.ResourceManager().DeleteHost(id3)
	*/

	loc1 := new(dbpb.DBLocation)
	loc1.Region = "Region1"
	loc1.Zone = "Zone4"
	loc1.Host = id1

	require1 := new(dbpb.DBRequirement)
	require1.ComputeReq = new(dbpb.DBComputeRequirement)
	require1.ComputeReq.CpuCores = 4
	require1.ComputeReq.Memory = 8
	require1.DiskReq = new(dbpb.DBDiskRequirement)
	require1.DiskReq.Capacity = 256
	require1.DiskReq.DiskType = string(resource.SSD)
	require1.DiskReq.NeedDisk = true
	require1.PortReq = append(require1.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req dbpb.DBAllocRequest
	test_req.Applicant = new(dbpb.DBApplicant)
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location: loc1,
		Strategy: int32(resource.UserSpecifyHost),
		Require:  require1,
		Count:    1,
	})

	loc2 := new(dbpb.DBLocation)
	loc2.Region = "Region1"
	loc2.Zone = "Zone4"
	loc2.Host = id2

	require2 := new(dbpb.DBRequirement)
	require2.ComputeReq = new(dbpb.DBComputeRequirement)
	require2.ComputeReq.CpuCores = 4
	require2.ComputeReq.Memory = 8
	require2.DiskReq = new(dbpb.DBDiskRequirement)
	require2.DiskReq.Capacity = 256
	require2.DiskReq.DiskType = string(resource.SSD)
	require2.DiskReq.NeedDisk = true
	require2.PortReq = append(require2.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location: loc2,
		Strategy: int32(resource.UserSpecifyHost),
		Require:  require2,
		Count:    1,
	})

	loc3 := new(dbpb.DBLocation)
	loc3.Region = "Region1"
	loc3.Zone = "Zone4"
	loc3.Host = id3

	require3 := new(dbpb.DBRequirement)
	require3.ComputeReq = new(dbpb.DBComputeRequirement)
	require3.ComputeReq.CpuCores = 4
	require3.ComputeReq.Memory = 8
	require3.DiskReq = new(dbpb.DBDiskRequirement)
	require3.DiskReq.Capacity = 256
	require3.DiskReq.DiskType = string(resource.SSD)
	require3.DiskReq.NeedDisk = true
	require3.PortReq = append(require3.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location: loc3,
		Strategy: int32(resource.UserSpecifyHost),
		Require:  require3,
		Count:    1,
	})

	var batchReq dbpb.DBBatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	assert.Equal(t, 1, len(batchReq.BatchRequests))

	type args struct {
		request *dbpb.DBBatchAllocRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{
			request: &batchReq,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := Dao.ResourceManager().AllocResourcesInBatch(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, 1, len(rsp.BatchResults))
			assert.Equal(t, 3, len(rsp.BatchResults[0].Results))
			assert.Equal(t, int32(2), rsp.BatchResults[0].Results[2].Reqseq)
			assert.Equal(t, int32(10003), rsp.BatchResults[0].Results[0].PortRes[0].Ports[3])
			assert.True(t, rsp.BatchResults[0].Results[0].HostIp != rsp.BatchResults[0].Results[1].HostIp)
			assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "474.111.111.127")
			assert.True(t, rsp.BatchResults[0].Results[1].HostIp == "474.111.111.128")
			assert.True(t, rsp.BatchResults[0].Results[2].HostIp == "474.111.111.129")
			assert.Equal(t, int32(4), rsp.BatchResults[0].Results[0].ComputeRes.CpuCores)
			assert.Equal(t, int32(8), rsp.BatchResults[0].Results[0].ComputeRes.Memory)
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "474.111.111.127")
			assert.Equal(t, int32(17-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.128")
			assert.Equal(t, int32(16-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.129")
			assert.Equal(t, int32(15-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_INUSED))
			//var usedPorts []int32
			var usedPorts []resource.UsedPort
			MetaDB.Order("port").Model(&resource.UsedPort{}).Where("host_id = ?", host3.ID).Scan(&usedPorts)
			assert.Equal(t, 5, len(usedPorts))
			assert.Equal(t, test_req.Applicant.HolderId, usedPorts[1].HolderId)
			assert.Equal(t, test_req.Applicant.RequestId, usedPorts[2].RequestId)
			for i := 0; i < 5; i++ {
				assert.Equal(t, int32(10000+i), usedPorts[i].Port)
			}
		})
	}
}

func TestAllocResources_SpecifyHost_Strategy_No_Disk(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone5", "3-1", "HostName1", "474.111.111.137", string(resource.General), string(resource.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Region1", "Zone5", "3-1", "HostName2", "474.111.111.138", string(resource.General), string(resource.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Region1", "Zone5", "3-1", "HostName3", "474.111.111.139", string(resource.General), string(resource.SSD), 15, 64, 3)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed
	/*
		defer Dao.ResourceManager().DeleteHost(id1)
		defer Dao.ResourceManager().DeleteHost(id2)
		defer Dao.ResourceManager().DeleteHost(id3)
	*/

	loc1 := new(dbpb.DBLocation)
	loc1.Region = "Region1"
	loc1.Zone = "Zone5"
	loc1.Host = id1

	require1 := new(dbpb.DBRequirement)
	require1.ComputeReq = new(dbpb.DBComputeRequirement)
	require1.ComputeReq.CpuCores = 4
	require1.ComputeReq.Memory = 8
	require1.DiskReq = new(dbpb.DBDiskRequirement)
	require1.DiskReq.NeedDisk = false
	require1.PortReq = append(require1.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req dbpb.DBAllocRequest
	test_req.Applicant = new(dbpb.DBApplicant)
	test_req.Applicant.HolderId = "TestCluster1"
	test_req.Applicant.RequestId = "TestRequestID1"
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location: loc1,
		Strategy: int32(resource.UserSpecifyHost),
		Require:  require1,
		Count:    1,
	})

	loc2 := new(dbpb.DBLocation)
	loc2.Region = "Region1"
	loc2.Zone = "Zone5"
	loc2.Host = id2

	require2 := new(dbpb.DBRequirement)
	require2.ComputeReq = new(dbpb.DBComputeRequirement)
	require2.ComputeReq.CpuCores = 4
	require2.ComputeReq.Memory = 8
	require2.DiskReq = new(dbpb.DBDiskRequirement)
	require2.DiskReq.Capacity = 256
	require2.DiskReq.DiskType = string(resource.SSD)
	require2.DiskReq.NeedDisk = true
	require2.PortReq = append(require2.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location: loc2,
		Strategy: int32(resource.UserSpecifyHost),
		Require:  require2,
		Count:    1,
	})

	loc3 := new(dbpb.DBLocation)
	loc3.Region = "Region1"
	loc3.Zone = "Zone5"
	loc3.Host = id3

	require3 := new(dbpb.DBRequirement)
	require3.ComputeReq = new(dbpb.DBComputeRequirement)
	require3.ComputeReq.CpuCores = 4
	require3.ComputeReq.Memory = 8
	require3.DiskReq = new(dbpb.DBDiskRequirement)
	require3.DiskReq.NeedDisk = false
	require3.PortReq = append(require3.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location: loc3,
		Strategy: int32(resource.UserSpecifyHost),
		Require:  require3,
		Count:    1,
	})

	var batchReq dbpb.DBBatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	assert.Equal(t, 1, len(batchReq.BatchRequests))

	type args struct {
		request *dbpb.DBBatchAllocRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{
			request: &batchReq,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := Dao.ResourceManager().AllocResourcesInBatch(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, 1, len(rsp.BatchResults))
			assert.Equal(t, 3, len(rsp.BatchResults[0].Results))
			assert.Equal(t, int32(2), rsp.BatchResults[0].Results[2].Reqseq)
			assert.Equal(t, int32(10003), rsp.BatchResults[0].Results[0].PortRes[0].Ports[3])
			assert.True(t, rsp.BatchResults[0].Results[0].HostIp != rsp.BatchResults[0].Results[1].HostIp)
			assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "474.111.111.137")
			assert.True(t, rsp.BatchResults[0].Results[1].HostIp == "474.111.111.138")
			assert.True(t, rsp.BatchResults[0].Results[2].HostIp == "474.111.111.139")
			assert.Equal(t, int32(4), rsp.BatchResults[0].Results[0].ComputeRes.CpuCores)
			assert.Equal(t, int32(8), rsp.BatchResults[0].Results[0].ComputeRes.Memory)
			assert.Equal(t, "", rsp.BatchResults[0].Results[0].DiskRes.DiskId)
			assert.Equal(t, "", rsp.BatchResults[0].Results[2].DiskRes.DiskId)
			disk1 := rsp.BatchResults[0].Results[1].DiskRes.DiskId
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "474.111.111.137")
			assert.Equal(t, int32(17-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			assert.True(t, host.Disks[1].Status == int32(resource.DISK_AVAILABLE) || host.Disks[2].Status == int32(resource.DISK_AVAILABLE) || host.Disks[3].Status == int32(resource.DISK_AVAILABLE))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.138")
			assert.Equal(t, int32(16-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			assert.True(t, disk1 == host2.Disks[1].ID || disk1 == host2.Disks[2].ID || disk1 == host2.Disks[3].ID)
			assert.True(t, host2.Disks[1].Status == int32(resource.DISK_EXHAUST) || host2.Disks[2].Status == int32(resource.DISK_EXHAUST) || host2.Disks[3].Status == int32(resource.DISK_EXHAUST), 3)
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.139")
			assert.Equal(t, int32(15-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_INUSED))
			assert.True(t, host3.Disks[1].Status == int32(resource.DISK_AVAILABLE) || host3.Disks[2].Status == int32(resource.DISK_AVAILABLE) || host3.Disks[3].Status == int32(resource.DISK_AVAILABLE))
			//var usedPorts []int32
			var usedPorts []resource.UsedPort
			MetaDB.Order("port").Model(&resource.UsedPort{}).Where("host_id = ?", host3.ID).Scan(&usedPorts)
			assert.Equal(t, 5, len(usedPorts))
			assert.Equal(t, test_req.Applicant.HolderId, usedPorts[1].HolderId)
			assert.Equal(t, test_req.Applicant.RequestId, usedPorts[2].RequestId)
			for i := 0; i < 5; i++ {
				assert.Equal(t, int32(10000+i), usedPorts[i].Port)
			}
		})
	}
}
