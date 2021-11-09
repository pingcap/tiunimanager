
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

package application

import (
	"context"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
	"reflect"
	"testing"
)

func TestCreateTenant(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.Tenant
		wantErr bool
	}{
		{"normal", args{"notExisted"}, &domain.Tenant{Name: "notExisted", Type: domain.InstanceWorkspace}, false},
		{"existed", args{"existed"}, &domain.Tenant{Name: "existed"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tenantManager.CreateTenant(context.TODO(), tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateTenant() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindTenant(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.Tenant
		wantErr bool
	}{
		{"normal", args{name: "111"}, &domain.Tenant{Name: "111"}, false},
		{"notExisted", args{name: "notExisted"}, &domain.Tenant{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tenantManager.FindTenant(context.TODO(), tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindTenant() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindTenantById(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.Tenant
		wantErr bool
	}{
		{"normal", args{id: "111"}, &domain.Tenant{Id: "111"}, false},
		{"notExisted", args{id: "notExisted"}, &domain.Tenant{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tenantManager.FindTenantById(context.TODO(), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindTenantById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindTenantById() got = %v, want %v", got, tt.want)
			}
		})
	}
}
