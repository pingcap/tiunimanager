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

package framework

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateInstance(t *testing.T) {
	type args struct {
		key   Key
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want Instance
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateInstance(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	LocalConfig = map[Key]Instance{
		"existed": {"existed", 1, 1},
	}
	type args struct {
		key Key
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{"normal", args{"existed"}, 1, false},
		{"not existed", args{"not existed"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetInstance(t *testing.T) {
	LocalConfig = map[Key]Instance{
		"existed": {"existed", 1, 1},
	}
	type args struct {
		key Key
	}
	tests := []struct {
		name    string
		args    args
		want    Instance
		wantErr bool
	}{
		{"normal", args{"existed"}, Instance{"existed", 1, 1}, false},
		{"not existed", args{"not existed"}, Instance{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetInstance(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInstance() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIntegerWithDefault(t *testing.T) {
	LocalConfig = map[Key]Instance{
		"existed": {"existed", 1, 1},
	}
	type args struct {
		key   Key
		value int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"normal", args{"existed", 5}, 1},
		{"not existed", args{"not existed", 5}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetIntegerWithDefault(tt.args.key, tt.args.value); got != tt.want {
				t.Errorf("GetIntegerWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStringWithDefault(t *testing.T) {
	LocalConfig = map[Key]Instance{
		"existed": {"existed", "111", 1},
	}
	type args struct {
		key   Key
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{"existed", "555"}, "111"},
		{"not existed", args{"not existed", "555"}, "555"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStringWithDefault(tt.args.key, tt.args.value); got != tt.want {
				t.Errorf("GetStringWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWithDefault(t *testing.T) {
	LocalConfig = map[Key]Instance{
		"existed":    {"existed", "111", 1},
		"existedInt": {"existedInt", 222, 1},
	}
	type args struct {
		key   Key
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{"normal", args{"existed", "555"}, "111"},
		{"not existed", args{"not existed", "555"}, "555"},

		{"existedInt", args{"existedInt", 999}, 222},
		{"not existedInt", args{"not existedInt", 999}, 999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetWithDefault(tt.args.key, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWithDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModifyGlobalConfig(t *testing.T) {
	type args struct {
		key   Key
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"normal", args{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ModifyGlobalConfig(tt.args.key, tt.args.value); got != tt.want {
				t.Errorf("ModifyGlobalConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModifyLocalServiceConfig(t *testing.T) {
	type args struct {
		key   Key
		value interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"normal", args{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ModifyLocalServiceConfig(tt.args.key, tt.args.value); got != tt.want {
				t.Errorf("ModifyLocalServiceConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateLocalConfig(t *testing.T) {
	type args struct {
		key        Key
		value      interface{}
		newVersion int
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 int
	}{
		{"normal", args{"existed", 3, 3}, true, 3},
		{"version changed", args{"existed", 1, 1}, false, 2},
		{"not existed", args{"not existed", 3, 3}, false, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LocalConfig = map[Key]Instance{
				"existed": {"existed", 2, 2},
			}
			got, got1 := UpdateLocalConfig(tt.args.key, tt.args.value, tt.args.newVersion)
			if got != tt.want {
				t.Errorf("UpdateLocalConfig() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("UpdateLocalConfig() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_CheckAndSetInEMTiupProcess(t *testing.T) {
	InitBaseFrameworkForUt(ClusterService)
	ok := CheckAndSetInEMTiupProcess()
	assert.True(t, ok)
	assert.True(t, GetBoolWithDefault(DuringEMTiupProcess, false))
	ok = CheckAndSetInEMTiupProcess()
	assert.False(t, ok)
	assert.True(t, GetBoolWithDefault(DuringEMTiupProcess, false))
	UnsetInEmTiupProcess()
	assert.False(t, GetBoolWithDefault(DuringEMTiupProcess, false))
	ok = CheckAndSetInEMTiupProcess()
	assert.True(t, ok)
	assert.True(t, GetBoolWithDefault(DuringEMTiupProcess, false))
	UnsetInEmTiupProcess()
	assert.False(t, GetBoolWithDefault(DuringEMTiupProcess, false))
}
