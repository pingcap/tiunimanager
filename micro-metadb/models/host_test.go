package models

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func TestCreateHost(t *testing.T) {
	h := &Host{
		HostName: "TestCreateHostRepeated",
		IP:       "111.111.111.111",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}
	id, _ := Dao.ResourceManager().CreateHost(h)
	defer Dao.ResourceManager().DeleteHost(id)
	type args struct {
		host *Host
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		assert  func(result string) bool
	}{
		{"normal", args{host: &Host{
			HostName: "TestCreateHost1",
			IP:       "192.168.125.132",
			Status:   0,
			OS:       "CentOS",
			Kernel:   "5.0.0",
			CpuCores: 5,
			Memory:   8,
			Nic:      "1GE",
			AZ:       "Zone1",
			Rack:     "3-1",
			Purpose:  "Compute",
			Disks: []Disk{
				{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
			},
		}}, false, func(result string) bool {
			return len(result) > 0
		}},
		{"same name", args{host: &Host{
			HostName: "TestCreateHostRepeated",
			IP:       "111.111.111.111",
			Status:   0,
			OS:       "CentOS",
			Kernel:   "5.0.0",
			CpuCores: 5,
			Memory:   8,
			Nic:      "1GE",
			AZ:       "Zone1",
			Rack:     "3-1",
			Purpose:  "Compute",
			Disks: []Disk{
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
		hosts []*Host
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(result []string) bool
	}{
		{"normal", args{hosts: []*Host{
			{
				HostName: "TestCreateHostsInBatch1",
				IP:       "192.168.11.111",
				Status:   0,
				OS:       "CentOS",
				Kernel:   "5.0.0",
				CpuCores: 5,
				Memory:   8,
				Nic:      "1GE",
				AZ:       "Zone1",
				Rack:     "3-1",
				Purpose:  "Compute",
				Disks:    []Disk{{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1}},
			},
			{
				HostName: "TestCreateHostsInBatch2",
				IP:       "192.111.125.111",
				Status:   0,
				OS:       "CentOS",
				Kernel:   "5.0.0",
				CpuCores: 5,
				Memory:   8,
				Nic:      "1GE",
				AZ:       "Zone1",
				Rack:     "3-1",
				Purpose:  "Compute",
				Disks:    []Disk{{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1}},
			},
		}}, false, []func(result []string) bool{
			func(result []string) bool { return len(result) == 2 },
			func(result []string) bool { return len(result[0]) > 0 },
		}},
		{"same name", args{hosts: []*Host{
			{
				HostName: "TestCreateHostsInBatchRepeated",
				IP:       "444.555.666.777",
				Status:   0,
				OS:       "CentOS",
				Kernel:   "5.0.0",
				CpuCores: 5,
				Memory:   8,
				Nic:      "1GE",
				AZ:       "Zone1",
				Rack:     "3-1",
				Purpose:  "Compute",
				Disks:    []Disk{{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1}},
			},
			{
				HostName: "TestCreateHostsInBatchRepeated",
				IP:       "444.555.666.777",
				Status:   0,
				OS:       "CentOS",
				Kernel:   "5.0.0",
				CpuCores: 5,
				Memory:   8,
				Nic:      "1GE",
				AZ:       "Zone1",
				Rack:     "3-1",
				Purpose:  "Compute",
				Disks:    []Disk{{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1}},
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
	h := &Host{
		HostName: "TestDeleteHost1",
		IP:       "192.99.999.132",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
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
	h := &Host{
		HostName: "主机1",
		IP:       "111.99.999.132",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}
	id1, _ := Dao.ResourceManager().CreateHost(h)
	defer Dao.ResourceManager().DeleteHost(id1)
	h2 := &Host{
		HostName: "主机1",
		IP:       "222.99.999.132",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
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
		s    DiskStatus
		want bool
	}{
		{"normal", DISK_AVAILABLE, true},
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
		s    DiskStatus
		want bool
	}{
		{"normal", DISK_INUSED, true},
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
			d := &Disk{
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

func TestDisk_TableName(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Disk{
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
			if got := d.TableName(); got != tt.want {
				t.Errorf("TableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindHostById(t *testing.T) {
	h := &Host{
		HostName: "主机1",
		IP:       "876.111.111.111",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
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
		asserts []func(result *Host) bool
	}{
		{"normal", args{hostId: id1}, false, []func(result *Host) bool{
			func(result *Host) bool { return id1 == result.ID },
			func(result *Host) bool { return h.IP == result.IP },
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
		s    HostStatus
		want bool
	}{
		{"normal_online", HOST_ONLINE, true},
		{"normal_inused", HOST_INUSED, true},
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

func TestHostStatus_IsInused(t *testing.T) {
	tests := []struct {
		name string
		s    HostStatus
		want bool
	}{
		{"normal_inused", HOST_INUSED, true},
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

func TestHostTableName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"normal", "hosts"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HostTableName(); got != tt.want {
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
		CpuCores  int
		Memory    int
		Spec      string
		Nic       string
		DC        string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []Disk
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
			h := &Host{
				ID:        tt.fields.ID,
				IP:        tt.fields.IP,
				HostName:  tt.fields.HostName,
				Status:    tt.fields.Status,
				OS:        tt.fields.OS,
				Kernel:    tt.fields.Kernel,
				CpuCores:  tt.fields.CpuCores,
				Memory:    tt.fields.Memory,
				Spec:      tt.fields.Spec,
				Nic:       tt.fields.Nic,
				DC:        tt.fields.DC,
				AZ:        tt.fields.AZ,
				Rack:      tt.fields.Rack,
				Purpose:   tt.fields.Purpose,
				Disks:     tt.fields.Disks,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
				DeletedAt: tt.fields.DeletedAt,
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
		CpuCores  int
		Memory    int
		Spec      string
		Nic       string
		DC        string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []Disk
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
			h := &Host{
				ID:        tt.fields.ID,
				IP:        tt.fields.IP,
				HostName:  tt.fields.HostName,
				Status:    tt.fields.Status,
				OS:        tt.fields.OS,
				Kernel:    tt.fields.Kernel,
				CpuCores:  tt.fields.CpuCores,
				Memory:    tt.fields.Memory,
				Spec:      tt.fields.Spec,
				Nic:       tt.fields.Nic,
				DC:        tt.fields.DC,
				AZ:        tt.fields.AZ,
				Rack:      tt.fields.Rack,
				Purpose:   tt.fields.Purpose,
				Disks:     tt.fields.Disks,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
				DeletedAt: tt.fields.DeletedAt,
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
		CpuCores  int
		Memory    int
		Spec      string
		Nic       string
		DC        string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []Disk
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
			h := &Host{
				ID:        tt.fields.ID,
				IP:        tt.fields.IP,
				HostName:  tt.fields.HostName,
				Status:    tt.fields.Status,
				OS:        tt.fields.OS,
				Kernel:    tt.fields.Kernel,
				CpuCores:  tt.fields.CpuCores,
				Memory:    tt.fields.Memory,
				Spec:      tt.fields.Spec,
				Nic:       tt.fields.Nic,
				DC:        tt.fields.DC,
				AZ:        tt.fields.AZ,
				Rack:      tt.fields.Rack,
				Purpose:   tt.fields.Purpose,
				Disks:     tt.fields.Disks,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
				DeletedAt: tt.fields.DeletedAt,
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
		CpuCores  int
		Memory    int
		Spec      string
		Nic       string
		DC        string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []Disk
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
			h := &Host{
				ID:        tt.fields.ID,
				IP:        tt.fields.IP,
				HostName:  tt.fields.HostName,
				Status:    tt.fields.Status,
				OS:        tt.fields.OS,
				Kernel:    tt.fields.Kernel,
				CpuCores:  tt.fields.CpuCores,
				Memory:    tt.fields.Memory,
				Spec:      tt.fields.Spec,
				Nic:       tt.fields.Nic,
				DC:        tt.fields.DC,
				AZ:        tt.fields.AZ,
				Rack:      tt.fields.Rack,
				Purpose:   tt.fields.Purpose,
				Disks:     tt.fields.Disks,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
				DeletedAt: tt.fields.DeletedAt,
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
		CpuCores  int
		Memory    int
		Spec      string
		Nic       string
		DC        string
		AZ        string
		Rack      string
		Purpose   string
		Disks     []Disk
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt gorm.DeletedAt
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"normal", fields{CpuCores: 4, Memory: 8, Disks: []Disk{{Status: int32(DISK_AVAILABLE)}}}, false},
		{"without_disk", fields{CpuCores: 4, Memory: 8}, true},
		{"without_cpu", fields{CpuCores: 0, Memory: 8, Disks: []Disk{{Status: int32(DISK_AVAILABLE)}}}, true},
		{"without_momery", fields{CpuCores: 4, Memory: 0, Disks: []Disk{{Status: int32(DISK_AVAILABLE)}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Host{
				ID:        tt.fields.ID,
				IP:        tt.fields.IP,
				HostName:  tt.fields.HostName,
				Status:    tt.fields.Status,
				OS:        tt.fields.OS,
				Kernel:    tt.fields.Kernel,
				CpuCores:  tt.fields.CpuCores,
				Memory:    tt.fields.Memory,
				Spec:      tt.fields.Spec,
				Nic:       tt.fields.Nic,
				DC:        tt.fields.DC,
				AZ:        tt.fields.AZ,
				Rack:      tt.fields.Rack,
				Purpose:   tt.fields.Purpose,
				Disks:     tt.fields.Disks,
				CreatedAt: tt.fields.CreatedAt,
				UpdatedAt: tt.fields.UpdatedAt,
				DeletedAt: tt.fields.DeletedAt,
			}
			if got := h.IsExhaust(); got != tt.want {
				t.Errorf("IsExhaust() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHost_SetDiskStatus(t *testing.T) {
	type fields struct {
		Disks []Disk
	}
	type args struct {
		diskId string
		s      DiskStatus
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wanted DiskStatus
	}{
		{"normal", fields{[]Disk{{ID: "fake1", Status: int32(DISK_AVAILABLE)}, {ID: "fake2", Status: int32(DISK_AVAILABLE)}}}, args{"fake2", DISK_INUSED}, DISK_INUSED},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Host{
				Disks: tt.fields.Disks,
			}
			h.SetDiskStatus(tt.args.diskId, tt.args.s)
			assert.Equal(t, int32(tt.wanted), h.Disks[1].Status)
		})
	}
}

func TestListHosts(t *testing.T) {
	h := &Host{
		HostName: "TestListHosts1",
		IP:       "111.121.999.132",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}

	id1, _ := Dao.ResourceManager().CreateHost(h)
	defer Dao.ResourceManager().DeleteHost(id1)

	h2 := &Host{
		HostName: "TestListHosts2",
		IP:       "222.121.999.132",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}
	id2, _ := Dao.ResourceManager().CreateHost(h2)
	defer Dao.ResourceManager().DeleteHost(id2)

	h3 := &Host{
		HostName: "TestListHosts3",
		IP:       "333.121.999.132",
		Status:   3,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 5,
		Memory:   8,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "whatever",
		Disks: []Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 1},
		},
	}
	id3, _ := Dao.ResourceManager().CreateHost(h3)
	defer Dao.ResourceManager().DeleteHost(id3)

	type args struct {
		req ListHostReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(a args, result []Host) bool
	}{
		{"normal", args{req: ListHostReq{
			Status:  0,
			Purpose: "Compute",
			Offset:  0,
			Limit:   2,
		}}, false, []func(a args, result []Host) bool{
			func(a args, result []Host) bool { return len(result) == 2 },
			func(a args, result []Host) bool { return result[1].Purpose == "Compute" },
		}},
		{"offset", args{req: ListHostReq{
			Status:  0,
			Purpose: "Compute",
			Offset:  1,
			Limit:   2,
		}}, false, []func(a args, result []Host) bool{
			func(a args, result []Host) bool { return len(result) == 1 },
		}},
		{"without Purpose", args{req: ListHostReq{
			Status: HOST_ONLINE,
			Offset: 0,
			Limit:  5,
		}}, false, []func(a args, result []Host) bool{
			func(a args, result []Host) bool { return len(result) == 2 },
		}},
		{"without status", args{req: ListHostReq{
			Status: HOST_WHATEVER,
			Offset: 0,
			Limit:  5,
		}}, false, []func(a args, result []Host) bool{
			func(a args, result []Host) bool { return len(result) == 3 },
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
	h := &Host{
		HostName: "主机1",
		IP:       "474.111.111.111",
		UserName: "root",
		Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 17,
		Memory:   64,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: 1},
			{Name: "sdb", Path: "/pd", Capacity: 256, Status: 0},
		},
	}
	h2 := &Host{
		HostName: "主机2",
		IP:       "474.111.111.112",
		UserName: "root",
		Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 16,
		Memory:   64,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/tidb", Capacity: 256, Status: 0},
			{Name: "sdc", Path: "/tidb2", Capacity: 256, Status: 0},
		},
	}
	h3 := &Host{
		HostName: "主机3",
		IP:       "474.111.111.113",
		UserName: "root",
		Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 15,
		Memory:   64,
		Nic:      "1GE",
		AZ:       "Zone1",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sdb", Path: "/tikv", Capacity: 256, Status: 0},
		},
	}
	id1, _ := Dao.ResourceManager().CreateHost(h)
	id2, _ := Dao.ResourceManager().CreateHost(h2)
	id3, _ := Dao.ResourceManager().CreateHost(h3)
	// Host Status should be inused or exhausted, so delete would failed
	defer Dao.ResourceManager().DeleteHost(id1)
	defer Dao.ResourceManager().DeleteHost(id2)
	defer Dao.ResourceManager().DeleteHost(id3)

	var m AllocReqs = make(map[string][]*HostAllocReq)
	m["pd"] = append(m["pd"], &HostAllocReq{
		FailureDomain: "Zone1",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["tidb"] = append(m["tidb"], &HostAllocReq{
		FailureDomain: "Zone1",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["tikv"] = append(m["tikv"], &HostAllocReq{
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
			pdIp := rsp["pd"][0].Ip
			tidbIp := rsp["tidb"][0].Ip
			tikvIp := rsp["tikv"][0].Ip
			assert.True(t, pdIp != tidbIp && tidbIp != tikvIp && tikvIp != pdIp)
			assert.True(t, pdIp == "474.111.111.111" || pdIp == "474.111.111.112" || pdIp == "474.111.111.113")
			assert.True(t, tidbIp == "474.111.111.111" || tidbIp == "474.111.111.112" || tidbIp == "474.111.111.113")
			assert.True(t, tikvIp == "474.111.111.111" || tikvIp == "474.111.111.112" || tikvIp == "474.111.111.113")
			assert.Equal(t, 4, rsp["pd"][0].CpuCores)
			assert.Equal(t, 8, rsp["pd"][0].Memory)
			var host Host
			MetaDB.First(&host, "IP = ?", "474.111.111.111")
			assert.Equal(t, 17-4, host.CpuCores)
			assert.Equal(t, 64-8, host.Memory)
			assert.True(t, host.Status == int32(HOST_EXHAUST))
			var host2 Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.112")
			assert.Equal(t, 16-4, host2.CpuCores)
			assert.Equal(t, 64-8, host2.Memory)
			assert.True(t, host2.Status == int32(HOST_INUSED))
			var host3 Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.113")
			assert.Equal(t, 15-4, host3.CpuCores)
			assert.Equal(t, 64-8, host3.Memory)
			assert.True(t, host3.Status == int32(HOST_EXHAUST))
		})
	}
}

func TestAllocHosts_1Host(t *testing.T) {
	h := &Host{
		HostName: "主机1",
		IP:       "192.168.56.99",
		UserName: "root",
		Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 17,
		Memory:   64,
		Nic:      "1GE",
		AZ:       "Zone99",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: 1},
			{Name: "sdb", Path: "/mnt1", Capacity: 256, Status: 0},
			{Name: "sdc", Path: "/mnt2", Capacity: 256, Status: 0},
			{Name: "sdd", Path: "/mnt3", Capacity: 256, Status: 0},
		},
	}

	id1, _ := Dao.ResourceManager().CreateHost(h)
	defer Dao.ResourceManager().DeleteHost(id1)

	var m AllocReqs = make(map[string][]*HostAllocReq)
	m["pd"] = append(m["pd"], &HostAllocReq{
		FailureDomain: "Zone99",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["tidb"] = append(m["tidb"], &HostAllocReq{
		FailureDomain: "Zone99",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["tikv"] = append(m["tikv"], &HostAllocReq{
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
			pdIp := rsp["pd"][0].Ip
			tidbIp := rsp["tidb"][0].Ip
			tikvIp := rsp["tikv"][0].Ip
			pdDisk := rsp["pd"][0].DiskName
			tidbDisk := rsp["tidb"][0].DiskName
			tikvDisk := rsp["tikv"][0].DiskName
			assert.True(t, pdIp == "192.168.56.99" && tidbIp == "192.168.56.99" && tikvIp == "192.168.56.99")
			assert.True(t, pdDisk != tidbDisk && tidbDisk != tikvDisk && tikvDisk != pdDisk)
			assert.True(t, pdDisk == "sdb" || pdDisk == "sdc" || pdDisk == "sdd")
			assert.True(t, tidbDisk == "sdb" || tidbDisk == "sdc" || tidbDisk == "sdd")
			assert.True(t, tikvDisk == "sdb" || tikvDisk == "sdc" || tikvDisk == "sdd")
			var host Host
			MetaDB.First(&host, "IP = ?", "192.168.56.99")
			assert.Equal(t, 17-4-4-4, host.CpuCores)
			assert.Equal(t, 64-8-8-8, host.Memory)
			assert.True(t, host.Status == int32(HOST_EXHAUST))

		})
	}
}

func TestAllocHosts_1Host_NotEnough(t *testing.T) {
	h := &Host{
		HostName: "主机2",
		IP:       "192.168.56.100",
		UserName: "root",
		Passwd:   "4bc5947d63aab7ad23cda5ca33df952e9678d7920428",
		Status:   0,
		OS:       "CentOS",
		Kernel:   "5.0.0",
		CpuCores: 17,
		Memory:   64,
		Nic:      "1GE",
		AZ:       "Zone100",
		Rack:     "3-1",
		Purpose:  "Compute",
		Disks: []Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: 1},
			{Name: "sdb", Path: "/mnt1", Capacity: 256, Status: 0},
			{Name: "sdc", Path: "/mnt2", Capacity: 256, Status: 0},
			{Name: "sdd", Path: "/mnt3", Capacity: 256, Status: 0},
		},
	}

	id1, _ := Dao.ResourceManager().CreateHost(h)
	defer Dao.ResourceManager().DeleteHost(id1)

	var m AllocReqs = make(map[string][]*HostAllocReq)
	m["pd"] = append(m["pd"], &HostAllocReq{
		FailureDomain: "Zone100",
		CpuCores:      4,
		Memory:        8,
		Count:         2,
	})
	m["tidb"] = append(m["tidb"], &HostAllocReq{
		FailureDomain: "Zone100",
		CpuCores:      4,
		Memory:        8,
		Count:         1,
	})
	m["tikv"] = append(m["tikv"], &HostAllocReq{
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
