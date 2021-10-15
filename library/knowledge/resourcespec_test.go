
/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package knowledge

import (
	"reflect"
	"testing"
)

func TestGenSpecCode(t *testing.T) {
	type args struct {
		cpuCores int32
		mem      int32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{3,6}, "3C6G"},
		{"normal", args{999,999}, "999C999G"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenSpecCode(tt.args.cpuCores, tt.args.mem); got != tt.want {
				t.Errorf("GenSpecCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCpu(t *testing.T) {
	type args struct {
		specCode string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"normal", args{"3C6G"}, 3},
		{"normal", args{"999C6G"}, 999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCpu(tt.args.specCode); got != tt.want {
				t.Errorf("ParseCpu() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseMemory(t *testing.T) {
	type args struct {
		specCode string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"normal", args{"3C6G"}, 6},
		{"normal", args{"3C999G"}, 999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseMemory(tt.args.specCode); got != tt.want {
				t.Errorf("ParseMemory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceSpec_GetAttributeValue(t *testing.T) {
	type fields struct {
		SpecItems []ResourceSpecItem
	}
	type args struct {
		attribute ResourceSpecAttribute
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := &ResourceSpec{
				SpecItems: tt.fields.SpecItems,
			}
			got, err := spec.GetAttributeValue(tt.args.attribute)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAttributeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAttributeValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}
