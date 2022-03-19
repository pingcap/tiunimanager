/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
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

package resourcepool

import (
	"os"
	"testing"

	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/common/errors"
	em_errors "github.com/pingcap-inc/tiem/common/errors"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"github.com/stretchr/testify/assert"
)

func Test_UpdateDisk(t *testing.T) {
	dbPath := "./test_resource_" + uuidutil.ShortId() + ".db"
	db, err := createDB(dbPath)
	assert.Nil(t, err)
	defer func() { _ = os.Remove(dbPath) }()

	host := genFakeHost("Region1", "Region1,Zone1", "Region1,Zone1,Rack1", "TEST_HOST1", "192.168.999.999", 32, 64,
		string(constants.EMProductIDDataMigration), string(constants.PurposeSchedule), string(constants.NVMeSSD))
	db.AutoMigrate(&Host{})
	db.AutoMigrate(&Disk{})
	inUsedDisk := Disk{
		Name:     "sdk",
		Path:     "/mnt/sdk",
		Type:     string(constants.NVMeSSD),
		Capacity: 256,
		Status:   string(constants.DiskExhaust),
	}
	host.Disks = append(host.Disks, inUsedDisk)
	err = db.Model(&Host{}).Create(host).Error
	assert.Nil(t, err)

	err = db.Delete(&Disk{ID: host.Disks[2].ID}).Error
	assert.NotNil(t, err)
	assert.Equal(t, em_errors.NewErrorf(em_errors.TIEM_RESOURCE_HOST_STILL_INUSED, "disk %s is still in used", host.Disks[2].ID).GetMsg(), err.(errors.EMError).GetMsg())

	newDisk := Disk{Name: "sdg", Path: "/mnt/sdg"}
	err = db.Model(&(host.Disks[2])).Updates(newDisk).Error
	assert.NotNil(t, err)
	assert.Equal(t, em_errors.NewErrorf(em_errors.TIEM_RESOURCE_UPDATE_DISK_ERROR, "update path for disk %s is not allowed", host.Disks[2].ID).GetMsg(), err.(em_errors.EMError).GetMsg())

	newDisk = Disk{Name: "sdg", Status: string(constants.DiskError)}
	err = db.Model(&(host.Disks[2])).Updates(newDisk).Error
	assert.Nil(t, err)

	var queryDisk Disk
	err = db.Model(&Disk{}).Find(&queryDisk, "id = ?", host.Disks[2].ID).Error
	assert.Nil(t, err)
	assert.Equal(t, "sdg", queryDisk.Name)
	assert.Equal(t, string(constants.DiskError), queryDisk.Status)
	assert.Equal(t, int32(256), queryDisk.Capacity)
	assert.Equal(t, "/mnt/sdk", queryDisk.Path)

	var disks []Disk
	err = db.Find(&disks, "host_id = ?", host.ID).Error
	assert.Nil(t, err)
	assert.Equal(t, 3, len(disks))

	err = db.Delete(&Host{ID: host.ID}).Error
	assert.Nil(t, err)

	err = db.Find(&disks, "host_id = ?", host.ID).Error
	assert.Nil(t, err)
	assert.Equal(t, 0, len(disks))
}

func Test_ValidateDisk(t *testing.T) {
	fakeHostId := "fake-host-id"
	diskName := "sdb"
	diskPath := "/mnt/sdb"
	var capacity int32 = 256
	type want struct {
		err error
	}
	tests := []struct {
		name     string
		hostId   string
		diskType constants.DiskType
		disk     Disk
		want     want
	}{
		{"normal", fakeHostId, constants.NVMeSSD, Disk{Name: diskName, Path: diskPath, Capacity: capacity, Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{nil}},
		{"normal_without_type", fakeHostId, constants.NVMeSSD, Disk{Name: diskName, Path: diskPath, Capacity: capacity, Status: string(constants.DiskAvailable)}, want{nil}},
		{"normal_without_status", fakeHostId, constants.NVMeSSD, Disk{Name: diskName, Path: diskPath, Capacity: capacity}, want{nil}},
		{"no_name", fakeHostId, constants.NVMeSSD, Disk{Path: diskPath, Capacity: capacity, Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "create disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
				fakeHostId, "", diskPath, capacity),
		}},
		{"no_path", fakeHostId, constants.NVMeSSD, Disk{Name: diskName, Capacity: capacity, Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "create disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
				fakeHostId, diskName, "", capacity),
		}},
		{"no_capacity", fakeHostId, constants.NVMeSSD, Disk{Name: diskName, Path: diskPath, Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "create disk failed for host %s, disk name (%s) or disk path (%s) or disk capacity (%d) invalid",
				fakeHostId, diskName, diskPath, 0),
		}},
		{"hostId_mismatch", fakeHostId, constants.NVMeSSD, Disk{Name: diskName, Path: diskPath, Capacity: capacity, HostID: "fake-host2", Type: string(constants.NVMeSSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "create disk %s %s failed, host id conflict %s vs %s",
				diskName, diskPath, "fake-host2", fakeHostId),
		}},
		{"diskType_mismatch", fakeHostId, constants.NVMeSSD, Disk{Name: diskName, Path: diskPath, Capacity: capacity, Type: string(constants.SSD), Status: string(constants.DiskAvailable)}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "create disk %s %s for host %s failed, disk type conflict %s vs %s",
				diskName, diskPath, fakeHostId, string(constants.SSD), string(constants.NVMeSSD)),
		}},
		{"status_invalid", fakeHostId, constants.NVMeSSD, Disk{Name: diskName, Path: diskPath, Capacity: capacity, Type: string(constants.NVMeSSD), Status: "bad_status"}, want{
			errors.NewErrorf(errors.TIEM_RESOURCE_VALIDATE_DISK_ERROR, "create disk %s %s for host %s specified a invalid status %s, [Available|Reserved]",
				diskName, diskPath, fakeHostId, "bad_status"),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.disk.ValidateDisk(tt.hostId, string(tt.diskType))
			if err == nil {
				assert.Equal(t, err, tt.want.err)
			} else {
				assert.Equal(t, tt.want.err.(errors.EMError).GetCode(), err.(errors.EMError).GetCode())
				assert.Equal(t, tt.want.err.(errors.EMError).GetMsg(), err.(errors.EMError).GetMsg())
			}
		})
	}
}
