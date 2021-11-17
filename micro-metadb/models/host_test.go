/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

package models

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/pingcap-inc/tiem/library/framework"

	"github.com/pingcap-inc/tiem/library/client/metadb/dbpb"
	"github.com/pingcap-inc/tiem/library/common"

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
	id, _ := Dao.ResourceManager().CreateHost(context.TODO(), h)
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id) }()
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
			gotId, err := Dao.ResourceManager().CreateHost(context.TODO(), tt.args.host)
			defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), gotId) }()
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
			gotIds, err := Dao.ResourceManager().CreateHostsInBatch(context.TODO(), tt.args.hosts)
			defer func() { _ = Dao.ResourceManager().DeleteHostsInBatch(context.TODO(), gotIds) }()
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
	id, _ := Dao.ResourceManager().CreateHost(context.TODO(), h)
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id) }()
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
			if err := Dao.ResourceManager().DeleteHost(context.TODO(), tt.args.hostId); (err != nil) != tt.wantErr {
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
	id1, _ := Dao.ResourceManager().CreateHost(context.TODO(), h)
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id1) }()
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
	id2, _ := Dao.ResourceManager().CreateHost(context.TODO(), h2)
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id2) }()

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
			if err := Dao.ResourceManager().DeleteHostsInBatch(context.TODO(), tt.args.hostIds); (err != nil) != tt.wantErr {
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
	id1, _ := Dao.ResourceManager().CreateHost(context.TODO(), h)
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id1) }()
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
			got, err := Dao.ResourceManager().FindHostById(context.TODO(), tt.args.hostId)
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
			gotRes, err := Dao.ResourceManager().GetFailureDomain(context.TODO(), tt.args.domain)
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
		{"normal_online", resource.Host{Status: int32(resource.HOST_ONLINE), Stat: int32(resource.HOST_LOADLESS)}, true},
		{"normal_inused", resource.Host{Status: int32(resource.HOST_ONLINE), Stat: int32(resource.HOST_INUSED)}, true},
		{"want false", resource.Host{Status: int32(resource.HOST_OFFLINE), Stat: int32(resource.HOST_INUSED)}, false},
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
		{"normal_inused", resource.Host{Status: int32(resource.HOST_ONLINE), Stat: int32(resource.HOST_INUSED)}, true},
		{"want false", resource.Host{Status: int32(resource.HOST_ONLINE), Stat: int32(resource.HOST_LOADLESS)}, false},
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
	type want struct {
		stat      resource.HostStat
		isExhaust bool
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{"normal", fields{CpuCores: 4, Memory: 8, Disks: []resource.Disk{{Status: int32(resource.DISK_AVAILABLE)}}}, want{resource.HOST_STAT_WHATEVER, false}},
		{"exhaust1", fields{CpuCores: 0, Memory: 0, Disks: []resource.Disk{{Status: int32(resource.DISK_INUSED)}}}, want{resource.HOST_EXHAUST, true}},
		{"exhaust2", fields{CpuCores: 0, Memory: 0, Disks: []resource.Disk{{Status: int32(resource.DISK_EXHAUST)}}}, want{resource.HOST_EXHAUST, true}},
		{"exhaust3", fields{CpuCores: 0, Memory: 0, Disks: []resource.Disk{{Status: int32(resource.DISK_EXHAUST)}, {Status: int32(resource.DISK_INUSED)}}}, want{resource.HOST_EXHAUST, true}},
		{"with_disk", fields{CpuCores: 0, Memory: 0, Disks: []resource.Disk{{Status: int32(resource.DISK_AVAILABLE)}}}, want{resource.HOST_COMPUTE_EXHAUST, true}},
		{"without_disk", fields{CpuCores: 4, Memory: 8}, want{resource.HOST_DISK_EXHAUST, true}},
		{"without_cpu", fields{CpuCores: 0, Memory: 8, Disks: []resource.Disk{{Status: int32(resource.DISK_AVAILABLE)}}}, want{resource.HOST_COMPUTE_EXHAUST, true}},
		{"without_momery", fields{CpuCores: 4, Memory: 0, Disks: []resource.Disk{{Status: int32(resource.DISK_AVAILABLE)}}}, want{resource.HOST_COMPUTE_EXHAUST, true}},
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
			if stat, isExaust := h.IsExhaust(); stat != tt.want.stat || isExaust != tt.want.isExhaust {
				t.Errorf("IsExhaust() = %v, want %v", stat, tt.want)
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

func CreateTestHost(region, zone, rack, hostName, ip, clusterType, purpose, diskType string, freeCpuCores, freeMemory, availableDiskCount int32) (id string, err error) {
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
		CpuCores:     freeCpuCores,
		Memory:       freeCpuCores,
		FreeCpuCores: freeCpuCores,
		FreeMemory:   freeMemory,
		Nic:          "1GE",
		Region:       region,
		AZ:           zone,
		Rack:         rack,
		ClusterType:  clusterType,
		Purpose:      purpose,
		DiskType:     diskType,
		Disks: []resource.Disk{
			{Name: "sda", Path: "/", Capacity: 256, Status: 1, Type: diskType},
		},
	}
	h.BuildDefaultTraits()
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
	return Dao.ResourceManager().CreateHost(context.TODO(), h)
}
func TestListHosts(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName1", "111.121.999.132", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 1)
	id2, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName2", "222.121.999.132", string(resource.Database), string(resource.Compute), string(resource.SSD), 16, 64, 0)
	id3, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName3", "333.121.999.132", string(resource.Database), string(resource.Storage), string(resource.SSD), 15, 64, 0)
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id1) }()
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id2) }()
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id3) }()

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
			Purpose: "Compute",
			Offset:  0,
			Limit:   2,
		}}, false, []func(a args, result []resource.Host) bool{
			func(a args, result []resource.Host) bool { return len(result) == 2 },
			func(a args, result []resource.Host) bool { return result[1].Purpose == "Compute" },
			func(a args, result []resource.Host) bool { return result[0].Traits == 69 },
		}},
		{"offset", args{req: ListHostReq{
			Status:  0,
			Purpose: "Compute",
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
			Purpose: "Compute",
			Offset:  0,
			Limit:   5,
		}}, false, []func(a args, result []resource.Host) bool{
			func(a args, result []resource.Host) bool { return len(result) == 2 },
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHosts, err := Dao.ResourceManager().ListHosts(context.TODO(), tt.args.req)
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
	id1, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName1", "474.111.111.111", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 8, 1)
	id2, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName2", "474.111.111.112", string(resource.Database), string(resource.Compute), string(resource.SSD), 4, 16, 2)
	id3, _ := CreateTestHost("Region1", "Zone1", "3-1", "HostName3", "474.111.111.113", string(resource.Database), string(resource.Compute), string(resource.SSD), 4, 8, 1)
	// Host Status should be inused or exhausted, so delete would failed
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id1) }()
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id2) }()
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id3) }()

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
			rsp, err := Dao.ResourceManager().AllocHosts(context.TODO(), tt.args.req)
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
			assert.Equal(t, int32(0), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			stat, isExaust := host.IsExhaust()
			assert.True(t, stat == resource.HOST_EXHAUST && isExaust == true)
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.112")
			assert.Equal(t, int32(0), host2.FreeCpuCores)
			assert.Equal(t, int32(16-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			stat, isExaust = host2.IsExhaust()
			assert.True(t, stat == resource.HOST_COMPUTE_EXHAUST && isExaust == true)
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.113")
			assert.Equal(t, int32(0), host3.FreeCpuCores)
			assert.Equal(t, int32(0), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_INUSED))
			stat, isExaust = host3.IsExhaust()
			assert.True(t, stat == resource.HOST_EXHAUST && isExaust == true)
		})
	}
}

func TestAllocHosts_1Host(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone99", "3-1", "HostName1", "192.168.56.99", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 3)
	// Host Status should be inused or exhausted, so delete would failed
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id1) }()

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
			rsp, err := Dao.ResourceManager().AllocHosts(context.TODO(), tt.args.req)
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
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))

		})
	}
}

func TestAllocHosts_1Host_NotEnough(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone100", "3-1", "HostName1", "192.168.56.100", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 3)
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id1) }()

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
			_, err := Dao.ResourceManager().AllocHosts(context.TODO(), tt.args.req)
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
	id1, _ := CreateTestHost("Region1", "Region1,Zone1", "Region1,Zone1,3-1", "HostName1", "474.111.111.108", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 1)
	id2, _ := CreateTestHost("Region1", "Region1,Zone1", "Region1,Zone1,3-1", "HostName2", "474.111.111.109", string(resource.Database), string(resource.Compute), string(resource.SSD), 16, 64, 2)
	id3, _ := CreateTestHost("Region1", "Region1,Zone1", "Region1,Zone1,3-1", "HostName3", "474.111.111.110", string(resource.Database), string(resource.Compute), string(resource.SSD), 15, 64, 1)
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
	filter.Purpose = string(resource.Compute)
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
			rsp, err := Dao.ResourceManager().AllocResources(context.TODO(), tt.args.request)
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
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.109")
			assert.Equal(t, int32(16-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.110")
			assert.Equal(t, int32(15-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_INUSED))
		})
	}
}

func TestAllocResources_3Requirement_3Hosts(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Region1,Zone2", "Region1,Zone2,3-1", "HostName1", "474.111.111.114", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 1)
	id2, _ := CreateTestHost("Region1", "Region1,Zone2", "Region1,Zone2,3-1", "HostName2", "474.111.111.115", string(resource.Database), string(resource.Compute), string(resource.SSD), 16, 64, 2)
	id3, _ := CreateTestHost("Region1", "Region1,Zone2", "Region1,Zone2,3-1", "HostName3", "474.111.111.116", string(resource.Database), string(resource.Storage), string(resource.SSD), 15, 64, 1)
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
	filter1.Purpose = string(resource.Compute)
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
			rsp, err := Dao.ResourceManager().AllocResources(context.TODO(), tt.args.request)
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
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.115")
			assert.Equal(t, int32(16-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.116")
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

func TestAllocResources_3RequestsInBatch_3Hosts(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Region1,Zone3", "Region1,Zone3,3-1", "HostName1", "474.111.111.117", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Region1", "Region1,Zone3", "Region1,Zone3,3-1", "HostName2", "474.111.111.118", string(resource.Database), string(resource.Compute), string(resource.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Region1", "Region1,Zone3", "Region1,Zone3,3-1", "HostName3", "474.111.111.119", string(resource.Database), string(resource.Compute), string(resource.SSD), 15, 64, 3)
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
	filter1.Purpose = string(resource.Compute)
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
			rsp, err := Dao.ResourceManager().AllocResourcesInBatch(context.TODO(), tt.args.request)
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
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			stat, isExaust := host.IsExhaust()
			assert.True(t, stat == resource.HOST_DISK_EXHAUST && isExaust == true)
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.118")
			assert.Equal(t, int32(16-4-4-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.119")
			assert.Equal(t, int32(15-4-4-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8-8-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_INUSED))
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
	id1, _ := CreateTestHost("Region111", "Region111,Zone3", "Region111,Zone3,3-1", "HostName1", "474.112.111.117", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Region111", "Region111,Zone3", "Region111,Zone3,3-1", "HostName2", "474.112.111.118", string(resource.Database), string(resource.Compute), string(resource.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Region111", "Region111,Zone3", "Region111,Zone3,3-1", "HostName3", "474.112.111.119", string(resource.Database), string(resource.Compute), string(resource.SSD), 15, 64, 3)
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
	filter1.Purpose = string(resource.Compute)
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
			rsp, err := Dao.ResourceManager().AllocResourcesInBatch(context.TODO(), tt.args.request)
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
	id1, _ := CreateTestHost("Region1", "Region1,Zone4", "Region1,Zone4,3-1", "HostName1", "474.111.111.127", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Region1", "Region1,Zone4", "Region1,Zone4,3-1", "HostName2", "474.111.111.128", string(resource.Database), string(resource.Compute), string(resource.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Region1", "Region1,Zone4", "Region1,Zone4,3-1", "HostName3", "474.111.111.129", string(resource.Database), string(resource.Compute), string(resource.SSD), 15, 64, 3)
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
	loc1.Host = "474.111.111.127"

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
	loc2.Host = "474.111.111.128"

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
	loc3.Host = "474.111.111.129"

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
			rsp, err := Dao.ResourceManager().AllocResourcesInBatch(context.TODO(), tt.args.request)
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
	id1, _ := CreateTestHost("Region1", "Region1,Zone5", "Region1,Zone5,3-1", "HostName1", "474.111.111.137", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Region1", "Region1,Zone5", "Region1,Zone5,3-1", "HostName2", "474.111.111.138", string(resource.Database), string(resource.Compute), string(resource.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Region1", "Region1,Zone5", "Region1,Zone5,3-1", "HostName3", "474.111.111.139", string(resource.Database), string(resource.Compute), string(resource.SSD), 15, 64, 3)
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
	loc1.Host = "474.111.111.137"

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
	loc2.Host = "474.111.111.138"

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
	loc3.Host = "474.111.111.139"

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
			rsp, err := Dao.ResourceManager().AllocResourcesInBatch(context.TODO(), tt.args.request)
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

func TestUpdateHost(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "TestUpdateZone", "3-1", "HostName1", "474.111.111.140", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 3)
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id1) }()
	reserved_req := dbpb.DBReserveHostRequest{
		Reserved: true,
	}
	reserved_req.HostIds = append(reserved_req.HostIds, id1)
	err := Dao.ResourceManager().ReserveHost(context.TODO(), &reserved_req)
	assert.Equal(t, nil, err)
	update_req := dbpb.DBUpdateHostStatusRequest{
		Status: 2,
	}
	update_req.HostIds = append(update_req.HostIds, id1)
	err = Dao.ResourceManager().UpdateHostStatus(context.TODO(), &update_req)
	assert.Equal(t, nil, err)
	var host resource.Host
	MetaDB.First(&host, "IP = ?", "474.111.111.140")
	assert.Equal(t, int32(2), host.Status)
	assert.Equal(t, true, host.Reserved)
}

func TestAllocResources_SpecifyHost_Strategy_TakeOver(t *testing.T) {
	id1, _ := CreateTestHost("Region1", "Zone4", "3-1", "HostName1", "474.111.111.147", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 3)
	reserved_req := dbpb.DBReserveHostRequest{
		Reserved: true,
	}
	reserved_req.HostIds = append(reserved_req.HostIds, id1)
	err := Dao.ResourceManager().ReserveHost(context.TODO(), &reserved_req)
	assert.Equal(t, nil, err)
	var host resource.Host
	MetaDB.First(&host, "IP = ?", "474.111.111.147")
	assert.Equal(t, true, host.Reserved)

	// Host Status should be inused or exhausted, so delete would failed
	/*
		defer Dao.ResourceManager().DeleteHost(id1)
	*/

	loc1 := new(dbpb.DBLocation)
	loc1.Region = "Region1"
	loc1.Zone = "Zone4"
	loc1.Host = "474.111.111.147"

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

	var batchReq dbpb.DBBatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	assert.Equal(t, 1, len(batchReq.BatchRequests))

	rsp, err := Dao.ResourceManager().AllocResourcesInBatch(context.TODO(), &batchReq)
	assert.True(t, nil == rsp && err != nil)
	t.Log(err)
	te, ok := err.(framework.TiEMError)
	assert.Equal(t, true, ok)
	assert.True(t, common.TIEM_RESOURCE_NOT_ALL_SUCCEED.Equal(int32(te.GetCode())))

	batchReq.BatchRequests[0].Applicant.TakeoverOperation = true
	rsp, err2 := Dao.ResourceManager().AllocResourcesInBatch(context.TODO(), &batchReq)
	t.Log(err2)
	assert.Equal(t, nil, err2)
	assert.True(t, rsp.BatchResults[0].Results[0].HostId == id1)
}

func TestGetStocks(t *testing.T) {
	id1, _ := CreateTestHost("Region47", "Region47,Zone15", "Region47,Zone15,3-1", "HostName1", "474.110.115.137", string(resource.Database), string(resource.Compute), string(resource.SSD), 17, 64, 3)
	id2, _ := CreateTestHost("Region47", "Region47,Zone15", "Region47,Zone15,3-1", "HostName2", "474.110.115.138", string(resource.Database), string(resource.Compute), string(resource.SSD), 16, 64, 3)
	id3, _ := CreateTestHost("Region47", "Region47,Zone25", "Region47,Zone25,3-1", "HostName3", "474.110.115.139", string(resource.Database), string(resource.Compute), string(resource.SSD), 15, 64, 3)

	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id1) }()
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id2) }()
	defer func() { _ = Dao.ResourceManager().DeleteHost(context.TODO(), id3) }()
	type args struct {
		req StockCondition
	}
	ssdType := resource.SSD
	archX86 := resource.X86
	diskStatusAvailable := resource.DISK_AVAILABLE
	tests := []struct {
		name    string
		args    args
		wantErr bool
		asserts []func(result []Stock) bool
	}{
		{"Zone25", args{req: StockCondition{
			Location:      resource.Location{Zone: "Region47,Zone25"},
			HostCondition: HostCondition{Arch: (*string)(&archX86)},
			DiskCondition: DiskCondition{Type: (*string)(&ssdType), Status: (*int32)(&diskStatusAvailable)},
		}}, false, []func(result []Stock) bool{
			func(result []Stock) bool { return len(result) == 1 },
			func(result []Stock) bool {
				return result[0].FreeCpuCores == 15 && result[0].FreeMemory == 64 && result[0].FreeDiskCount == 3 && result[0].FreeDiskCapacity == 256*3
			},
		}},
		{"Zone15", args{req: StockCondition{
			Location:      resource.Location{Zone: "Region47,Zone15"},
			HostCondition: HostCondition{Arch: (*string)(&archX86)},
			DiskCondition: DiskCondition{Type: (*string)(&ssdType), Status: (*int32)(&diskStatusAvailable)},
		}}, false, []func(result []Stock) bool{
			func(result []Stock) bool { return len(result) == 2 },
			func(result []Stock) bool {
				var free_cpu_cores int
				var free_memory int
				var free_disk_capacity int
				var free_disk_count int
				for i := 0; i < len(result); i++ {
					free_cpu_cores += result[i].FreeCpuCores
					free_memory += result[i].FreeMemory
					free_disk_count += result[i].FreeDiskCount
					free_disk_capacity += result[i].FreeDiskCapacity
				}
				return free_cpu_cores == 33 && free_memory == 128 && free_disk_count == 6 && free_disk_capacity == 256*2*3
			},
		}},
		{"Region47", args{req: StockCondition{
			Location:      resource.Location{Region: "Region47"},
			DiskCondition: DiskCondition{Status: (*int32)(&diskStatusAvailable)},
		}}, false, []func(result []Stock) bool{
			func(result []Stock) bool { return len(result) == 3 },
			func(result []Stock) bool {
				var free_cpu_cores int
				var free_memory int
				var free_disk_capacity int
				var free_disk_count int
				for i := 0; i < len(result); i++ {
					free_cpu_cores += result[i].FreeCpuCores
					free_memory += result[i].FreeMemory
					free_disk_count += result[i].FreeDiskCount
					free_disk_capacity += result[i].FreeDiskCapacity
				}
				return free_cpu_cores == 48 && free_memory == 192 && free_disk_count == 9 && free_disk_capacity == 256*3*3
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stocks, err := Dao.ResourceManager().GetStocks(context.TODO(), tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListHosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, assert := range tt.asserts {
				if !assert(stocks) {
					t.Errorf("ListHosts() assert false, index = %v, stocks %v", i, stocks)
				}
			}
		})
	}
}

func TestAllocResources_1Requirement_3Hosts_Filted_by_Label(t *testing.T) {
	id1, _ := CreateTestHost("Region59", "Region59,Zone59", "Region59,Zone59,3-1", "HostName1", "474.111.111.158", string(resource.Database), string(resource.Compute)+","+string(resource.Storage), string(resource.SSD), 17, 64, 1)
	id2, _ := CreateTestHost("Region59", "Region59,Zone59", "Region59,Zone59,3-1", "HostName2", "474.111.111.159", string(resource.Database), string(resource.Compute), string(resource.SSD), 16, 64, 2)
	id3, _ := CreateTestHost("Region59", "Region59,Zone59", "Region59,Zone59,3-1", "HostName3", "474.111.111.150", string(resource.Database), string(resource.Compute), string(resource.SSD), 15, 64, 1)
	t.Log(id1, id2, id3)
	// Host Status should be inused or exhausted, so delete would failed

	//	defer Dao.ResourceManager().DeleteHost(id1)
	//	defer Dao.ResourceManager().DeleteHost(id2)
	//	defer Dao.ResourceManager().DeleteHost(id3)

	loc := new(dbpb.DBLocation)
	loc.Region = "Region59"
	loc.Zone = "Zone59"
	filter := new(dbpb.DBFilter)
	filter.Arch = string(resource.X86)
	//filter.DiskType = string(resource.SSD)
	//filter.Purpose = string(resource.Compute)
	filter.HostTraits = 5
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
	test_req.Applicant.HolderId = "TestCluster59"
	test_req.Applicant.RequestId = "TestRequestID59"
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
			rsp, err := Dao.ResourceManager().AllocResources(context.TODO(), tt.args.request)
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
			assert.True(t, rsp.Results[0].HostIp == "474.111.111.158" || rsp.Results[0].HostIp == "474.111.111.159" || rsp.Results[0].HostIp == "474.111.111.150")
			assert.True(t, rsp.Results[1].HostIp == "474.111.111.158" || rsp.Results[1].HostIp == "474.111.111.159" || rsp.Results[1].HostIp == "474.111.111.150")
			assert.True(t, rsp.Results[2].HostIp == "474.111.111.158" || rsp.Results[2].HostIp == "474.111.111.159" || rsp.Results[2].HostIp == "474.111.111.150")
			assert.Equal(t, int32(4), rsp.Results[0].ComputeRes.CpuCores)
			assert.Equal(t, int32(8), rsp.Results[0].ComputeRes.Memory)
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "474.111.111.158")
			assert.Equal(t, int32(17-4), host.FreeCpuCores)
			assert.Equal(t, int32(64-8), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "474.111.111.159")
			assert.Equal(t, int32(16-4), host2.FreeCpuCores)
			assert.Equal(t, int32(64-8), host2.FreeMemory)
			assert.True(t, host2.Stat == int32(resource.HOST_INUSED))
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "474.111.111.150")
			assert.Equal(t, int32(15-4), host3.FreeCpuCores)
			assert.Equal(t, int32(64-8), host3.FreeMemory)
			assert.True(t, host3.Stat == int32(resource.HOST_INUSED))
		})
	}
}

func TestAllocResources_Host_Recycle_Strategy(t *testing.T) {
	id1, _ := CreateTestHost("Region37", "Region37,Zone41", "Region37,Zone41,3-1", "HostName1", "437.111.111.127", string(resource.Database), string(resource.Compute), string(resource.SSD), 32, 64, 3)
	t.Log(id1)
	// Host Status should be inused or exhausted, so delete would failed
	/*
		defer Dao.ResourceManager().DeleteHost(id1)
		defer Dao.ResourceManager().DeleteHost(id2)
	*/

	loc1 := new(dbpb.DBLocation)
	loc1.Region = "Region37"
	loc1.Zone = "Zone41"
	loc1.Host = "437.111.111.127"

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
	test_req.Applicant.HolderId = "TestCluster11"
	test_req.Applicant.RequestId = "TestRequestID11"
	test_req.Requires = append(test_req.Requires, &dbpb.DBAllocRequirement{
		Location: loc1,
		Strategy: int32(resource.UserSpecifyHost),
		Require:  require1,
		Count:    1,
	})

	loc2 := new(dbpb.DBLocation)
	loc2.Region = "Region37"
	loc2.Zone = "Zone41"
	loc2.Host = "437.111.111.127"

	require2 := new(dbpb.DBRequirement)
	require2.ComputeReq = new(dbpb.DBComputeRequirement)
	require2.ComputeReq.CpuCores = 16
	require2.ComputeReq.Memory = 16
	require2.DiskReq = new(dbpb.DBDiskRequirement)
	require2.DiskReq.Capacity = 256
	require2.DiskReq.DiskType = string(resource.SSD)
	require2.DiskReq.NeedDisk = true
	require2.PortReq = append(require2.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req2 dbpb.DBAllocRequest
	test_req2.Applicant = new(dbpb.DBApplicant)
	test_req2.Applicant.HolderId = "TestCluster12"
	test_req2.Applicant.RequestId = "TestRequestID12"
	test_req2.Requires = append(test_req2.Requires, &dbpb.DBAllocRequirement{
		Location: loc2,
		Strategy: int32(resource.UserSpecifyHost),
		Require:  require2,
		Count:    1,
	})

	loc3 := new(dbpb.DBLocation)
	loc3.Region = "Region37"
	loc3.Zone = "Zone41"
	loc3.Host = "437.111.111.127"

	require3 := new(dbpb.DBRequirement)
	require3.ComputeReq = new(dbpb.DBComputeRequirement)
	require3.ComputeReq.CpuCores = 8
	require3.ComputeReq.Memory = 32
	require3.DiskReq = new(dbpb.DBDiskRequirement)
	require3.DiskReq.Capacity = 256
	require3.DiskReq.DiskType = string(resource.SSD)
	require3.DiskReq.NeedDisk = true
	require3.PortReq = append(require3.PortReq, &dbpb.DBPortRequirement{
		Start:   10000,
		End:     10015,
		PortCnt: 5,
	})

	var test_req3 dbpb.DBAllocRequest
	test_req3.Applicant = new(dbpb.DBApplicant)
	test_req3.Applicant.HolderId = "TestCluster11"
	test_req3.Applicant.RequestId = "TestRequestID13"
	test_req3.Requires = append(test_req3.Requires, &dbpb.DBAllocRequirement{
		Location: loc3,
		Strategy: int32(resource.UserSpecifyHost),
		Require:  require3,
		Count:    1,
	})

	var batchReq dbpb.DBBatchAllocRequest
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req2)
	batchReq.BatchRequests = append(batchReq.BatchRequests, &test_req3)
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
			rsp, err := Dao.ResourceManager().AllocResourcesInBatch(context.TODO(), tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("AllocHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, 3, len(rsp.BatchResults))
			assert.True(t, len(rsp.BatchResults[0].Results) == 1 && len(rsp.BatchResults[1].Results) == 1 && len(rsp.BatchResults[2].Results) == 1)
			assert.Equal(t, int32(10008), rsp.BatchResults[1].Results[0].PortRes[0].Ports[3])

			assert.True(t, rsp.BatchResults[0].Results[0].HostIp == "437.111.111.127" && rsp.BatchResults[1].Results[0].HostIp == "437.111.111.127" && rsp.BatchResults[2].Results[0].HostIp == "437.111.111.127")
			var host resource.Host
			MetaDB.First(&host, "IP = ?", "437.111.111.127")
			assert.Equal(t, int32(32-4-16-8), host.FreeCpuCores)
			assert.Equal(t, int32(64-8-16-32), host.FreeMemory)
			assert.True(t, host.Stat == int32(resource.HOST_INUSED))

			//var usedPorts []int32
			var usedPorts []resource.UsedPort
			MetaDB.Order("port").Model(&resource.UsedPort{}).Where("host_id = ?", host.ID).Scan(&usedPorts)
			assert.Equal(t, 15, len(usedPorts))
			for i := 0; i < 15; i++ {
				assert.Equal(t, int32(10000+i), usedPorts[i].Port)
			}

			type UsedCompute struct {
				TotalCpuCores int
				TotalMemory   int
				HolderId      string
			}
			var usedCompute []UsedCompute
			MetaDB.Model(&resource.UsedCompute{}).Select("sum(cpu_cores) as TotalCpuCores, sum(memory) as TotalMemory, holder_id").Group("holder_id").Having("holder_id == ? or holder_id == ?", "TestCluster11", "TestCluster12").Order("holder_id").Scan(&usedCompute)
			assert.Equal(t, 2, len(usedCompute))
			assert.True(t, usedCompute[0].TotalCpuCores == 12 && usedCompute[0].TotalMemory == 40 && usedCompute[0].HolderId == "TestCluster11")
			assert.True(t, usedCompute[1].TotalCpuCores == 16 && usedCompute[1].TotalMemory == 16 && usedCompute[1].HolderId == "TestCluster12")

			diskId1 := rsp.BatchResults[0].Results[0].DiskRes.DiskId // used by cluster1
			diskId2 := rsp.BatchResults[1].Results[0].DiskRes.DiskId // used by cluster2
			diskId3 := rsp.BatchResults[2].Results[0].DiskRes.DiskId // used by cluster3
			var disks []resource.Disk
			MetaDB.Model(&resource.Disk{}).Where("id = ? or id = ? or id = ?", diskId1, diskId2, diskId3).Scan(&disks)

			assert.Equal(t, int32(resource.DISK_EXHAUST), disks[0].Status)
			assert.Equal(t, int32(resource.DISK_EXHAUST), disks[1].Status)
			assert.Equal(t, int32(resource.DISK_EXHAUST), disks[2].Status)

			// recycle part 1
			var request dbpb.DBRecycleRequest
			var recycleRequire dbpb.DBRecycleRequire
			recycleRequire.HostId = id1
			recycleRequire.HolderId = "TestCluster11"
			recycleRequire.ComputeReq = new(dbpb.DBComputeRequirement)
			recycleRequire.ComputeReq.CpuCores = 4
			recycleRequire.ComputeReq.Memory = 8
			recycleRequire.DiskReq = append(recycleRequire.DiskReq, &dbpb.DBDiskResource{
				DiskId: diskId1,
			})
			recycleRequire.PortReq = append(recycleRequire.PortReq, &dbpb.DBPortResource{
				Ports: []int32{10000, 10001, 10002},
			})
			recycleRequire.RecycleType = int32(resource.RecycleHost)
			request.RecycleReqs = append(request.RecycleReqs, &recycleRequire)
			err = Dao.resourceManager.RecycleAllocResources(context.TODO(), &request)
			assert.Nil(t, err)
			var host2 resource.Host
			MetaDB.First(&host2, "IP = ?", "437.111.111.127")
			assert.Equal(t, int32(32-16-8), host2.FreeCpuCores)
			assert.Equal(t, int32(64-16-32), host2.FreeMemory)
			var usedCompute2 []UsedCompute
			MetaDB.Model(&resource.UsedCompute{}).Select("sum(cpu_cores) as TotalCpuCores, sum(memory) as TotalMemory, holder_id").Group("holder_id").Having("holder_id == ? or holder_id == ?", "TestCluster11", "TestCluster12").Order("holder_id").Scan(&usedCompute2)
			assert.Equal(t, 2, len(usedCompute2))
			assert.True(t, usedCompute2[0].TotalCpuCores == 8 && usedCompute2[0].TotalMemory == 32 && usedCompute2[0].HolderId == "TestCluster11")
			assert.True(t, usedCompute2[1].TotalCpuCores == 16 && usedCompute2[1].TotalMemory == 16 && usedCompute2[1].HolderId == "TestCluster12")
			var usedPorts2 []resource.UsedPort
			MetaDB.Order("port").Model(&resource.UsedPort{}).Where("host_id = ?", host.ID).Scan(&usedPorts2)
			assert.Equal(t, 12, len(usedPorts2))
			for i := 3; i < 15; i++ {
				assert.Equal(t, int32(10000+i), usedPorts2[i-3].Port)
			}
			var disk1 resource.Disk
			MetaDB.Model(&resource.Disk{}).Where("id = ?", diskId1).Scan(&disk1)
			assert.Equal(t, int32(resource.DISK_AVAILABLE), disk1.Status)

			// recycle part 2
			var request2 dbpb.DBRecycleRequest
			var recycleRequire2 dbpb.DBRecycleRequire
			recycleRequire2.HostId = id1
			recycleRequire2.HolderId = "TestCluster11"
			recycleRequire2.ComputeReq = new(dbpb.DBComputeRequirement)
			recycleRequire2.ComputeReq.CpuCores = 8
			recycleRequire2.ComputeReq.Memory = 32
			recycleRequire2.DiskReq = append(recycleRequire2.DiskReq, &dbpb.DBDiskResource{
				DiskId: diskId3,
			})

			recycleRequire2.PortReq = append(recycleRequire2.PortReq, &dbpb.DBPortResource{
				Ports: []int32{10003, 10004, 10010, 10011, 10012, 10013, 10014},
			})
			recycleRequire2.RecycleType = int32(resource.RecycleHost)
			request2.RecycleReqs = append(request2.RecycleReqs, &recycleRequire2)
			err = Dao.resourceManager.RecycleAllocResources(context.TODO(), &request2)
			assert.Nil(t, err)
			var host3 resource.Host
			MetaDB.First(&host3, "IP = ?", "437.111.111.127")
			assert.Equal(t, int32(32-16), host3.FreeCpuCores)
			assert.Equal(t, int32(64-16), host3.FreeMemory)
			var usedCompute3 []UsedCompute
			MetaDB.Model(&resource.UsedCompute{}).Select("sum(cpu_cores) as TotalCpuCores, sum(memory) as TotalMemory, holder_id").Group("holder_id").Having("holder_id == ? or holder_id == ?", "TestCluster11", "TestCluster12").Order("holder_id").Scan(&usedCompute3)
			assert.Equal(t, 1, len(usedCompute3))
			assert.True(t, usedCompute3[0].TotalCpuCores == 16 && usedCompute3[0].TotalMemory == 16 && usedCompute3[0].HolderId == "TestCluster12")
			var usedPorts3 []resource.UsedPort
			MetaDB.Order("port").Model(&resource.UsedPort{}).Where("host_id = ?", host.ID).Scan(&usedPorts3)
			assert.Equal(t, 5, len(usedPorts3))
			for i := 5; i < 10; i++ {
				assert.Equal(t, int32(10000+i), usedPorts3[i-5].Port)
			}
			var disk2 resource.Disk
			MetaDB.Model(&resource.Disk{}).Where("id = ?", diskId3).Scan(&disk2)
			assert.Equal(t, int32(resource.DISK_AVAILABLE), disk2.Status)

			// recycle all
			var request3 dbpb.DBRecycleRequest
			var recycleRequire3 dbpb.DBRecycleRequire
			recycleRequire3.HostId = id1
			recycleRequire3.HolderId = "TestCluster12"
			recycleRequire3.ComputeReq = new(dbpb.DBComputeRequirement)
			recycleRequire3.ComputeReq.CpuCores = 16
			recycleRequire3.ComputeReq.Memory = 16
			recycleRequire3.DiskReq = append(recycleRequire3.DiskReq, &dbpb.DBDiskResource{
				DiskId: diskId2,
			})
			recycleRequire3.PortReq = append(recycleRequire3.PortReq, &dbpb.DBPortResource{
				Ports: []int32{10005, 10006, 10007, 10008, 10009},
			})
			recycleRequire3.RecycleType = int32(resource.RecycleHost)
			request3.RecycleReqs = append(request3.RecycleReqs, &recycleRequire3)
			err = Dao.resourceManager.RecycleAllocResources(context.TODO(), &request3)
			assert.Nil(t, err)
			var host4 resource.Host
			MetaDB.First(&host4, "IP = ?", "437.111.111.127")
			assert.Equal(t, int32(32), host4.FreeCpuCores)
			assert.Equal(t, int32(64), host4.FreeMemory)
			//assert.Equal(t, int32(resource.HOST_LOADLESS), host4.Stat)
			var usedCompute4 []UsedCompute
			MetaDB.Model(&resource.UsedCompute{}).Select("sum(cpu_cores) as TotalCpuCores, sum(memory) as TotalMemory, holder_id").Group("holder_id").Having("holder_id == ? or holder_id == ?", "TestCluster11", "TestCluster12").Order("holder_id").Scan(&usedCompute4)
			assert.Equal(t, 0, len(usedCompute4))
			var usedPorts4 []resource.UsedPort
			MetaDB.Order("port").Model(&resource.UsedPort{}).Where("host_id = ?", host.ID).Scan(&usedPorts4)
			assert.Equal(t, 0, len(usedPorts4))
			var disk3 resource.Disk
			MetaDB.Model(&resource.Disk{}).Where("id = ?", diskId2).Scan(&disk3)
			assert.Equal(t, int32(resource.DISK_AVAILABLE), disk3.Status)
		})
	}
}
