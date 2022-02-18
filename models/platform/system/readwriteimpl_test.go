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
 ******************************************************************************/

/*******************************************************************************
 * @File: readwriteimpl_test.go.go
 * @Description:
 * @Author: zhangpeijin@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/2/16
*******************************************************************************/

package system

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/models/common"
	"gorm.io/gorm"
	"reflect"
	"testing"
)

func TestNewSystemReadWrite(t *testing.T) {
	type args struct {
		db *gorm.DB
	}
	tests := []struct {
		name string
		args args
		want ReaderWriter
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSystemReadWrite(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSystemReadWrite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSystemReadWrite_GetHistoryVersions(t *testing.T) {
	type fields struct {
		GormDB common.GormDB
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*VersionInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SystemReadWrite{
				GormDB: tt.fields.GormDB,
			}
			got, err := s.GetHistoryVersions(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHistoryVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHistoryVersions() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSystemReadWrite_GetSystemInfo(t *testing.T) {
	type fields struct {
		GormDB common.GormDB
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SystemInfo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SystemReadWrite{
				GormDB: tt.fields.GormDB,
			}
			got, err := s.GetSystemInfo(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSystemInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSystemInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSystemReadWrite_UpdateStatus(t *testing.T) {
	type fields struct {
		GormDB common.GormDB
	}
	type args struct {
		ctx      context.Context
		original constants.SystemStatus
		target   constants.SystemStatus
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
			s := &SystemReadWrite{
				GormDB: tt.fields.GormDB,
			}
			if err := s.UpdateStatus(tt.args.ctx, tt.args.original, tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSystemReadWrite_UpdateVersion(t *testing.T) {
	type fields struct {
		GormDB common.GormDB
	}
	type args struct {
		ctx    context.Context
		target constants.SystemStatus
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
			s := &SystemReadWrite{
				GormDB: tt.fields.GormDB,
			}
			if err := s.UpdateVersion(tt.args.ctx, tt.args.target); (err != nil) != tt.wantErr {
				t.Errorf("UpdateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
