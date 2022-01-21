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
 ******************************************************************************/

package common

import (
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/util/uuidutil"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

type TestEntity struct {
	Entity
	Name string `gorm:"uniqueIndex:myIndex"`
	DeleteTime int64 `gorm:"uniqueIndex:myIndex"`
	Password Password  `gorm:"password"`
}

func (e *TestEntity) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Model(e).Update("delete_time", time.Now().Unix())
	return nil
}

var baseDB *gorm.DB

func TestMain(m *testing.M) {
	testFilePath := "testdata/" + uuidutil.ShortId()
	os.MkdirAll(testFilePath, 0755)
	logins := framework.LogForkFile(constants.LogFileSystem)

	defer func() {
		os.RemoveAll(testFilePath)
		os.Remove(testFilePath)
	}()

	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			dbFile := testFilePath + constants.DBDirPrefix + constants.DatabaseFileName
			db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})

			if err != nil || db.Error != nil {
				logins.Fatalf("open database failed, filepath: %s database error: %s, meta database error: %v", dbFile, err, db.Error)
			} else {
				logins.Infof("open database successful, filepath: %s", dbFile)
			}

			baseDB = db
			db.Migrator().CreateTable(new(TestEntity))

			return nil
		},
	)
	os.Exit(m.Run())
}

func TestUniqueIndex(t *testing.T)  {
	entity := &TestEntity{
		Entity: Entity{
			TenantId: "111",
		},
		Name: "aaa",
	}

	err := baseDB.Create(entity).Error
	assert.NoError(t, err)

	err = baseDB.Create(&TestEntity{
		Entity: Entity{
			TenantId: "111",
		},
		Name: "aaa",
	}).Error
	assert.Error(t, err)

	baseDB.Delete(entity)

	err = baseDB.Create(&TestEntity{
		Entity: Entity{
			TenantId: "111",
		},
		Name: "aaa",
	}).Error
	assert.NoError(t, err)
}

func TestFinalHash(t *testing.T) {
	type args struct {
		salt   string
		passwd string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"normal", args{salt: "&shgdjsdfgjhfgksdh", passwd: "Test12345678"}, false},
		{"empty password", args{salt: "&shgdjsdfgjhfgksdh", passwd: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FinalHash(tt.args.salt, tt.args.passwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("FinalHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got)
			}
		})
	}
}

func TestPassword(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		err := baseDB.Create(&TestEntity{
			Entity: Entity{
				TenantId: "111",
			},
			Name: "createpassword",
			Password: "N&HIO(*(&#Y*&HNS&D*#*GF*RS*FY&DF",
		}).Error
		assert.NoError(t, err)

		result := &TestEntity{}
		err = baseDB.Model(&TestEntity{}).Where("name = ?", "createpassword").First(result).Error
		assert.NoError(t, err)
		assert.Equal(t, "N&HIO(*(&#Y*&HNS&D*#*GF*RS*FY&DF", string(result.Password))
	})
	t.Run("update", func(t *testing.T) {
		a := &TestEntity{
			Entity: Entity{
				TenantId: "111",
			},
			Name: "updatepassword",
			Password: "abcd",
		}
		baseDB.Create(a)
		a.Password = "dddd"
		err := baseDB.Model(a).Save(a).Error
		assert.NoError(t, err)

		result := &TestEntity{}
		err = baseDB.Model(&TestEntity{}).Where("name = ?", "updatepassword").First(result).Error
		assert.NoError(t, err)
		assert.Equal(t, "dddd", string(result.Password))
	})
}