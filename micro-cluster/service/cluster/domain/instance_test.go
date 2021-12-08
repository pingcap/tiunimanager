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

package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComponentInstance_AcceptPortInfo(t *testing.T) {
	type args struct {
		portInfo string
	}
	tests := []struct {
		name   string
		args   args
		want   []int
	}{
		{"normal", args{"[1,2,3]"}, []int{1,2,3}},
		{"empty", args{"[]"}, []int{}},
		{"empty", args{""}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ComponentInstance{}
			c.DeserializePortInfo(tt.args.portInfo)
			if tt.args.portInfo == "[]" || tt.args.portInfo == ""{
				assert.NotNil(t, c.PortList)
				assert.Equal(t, 0, len(c.PortList))
			} else {
				assert.Equal(t, c.PortList[0], tt.want[0])
			}
		})
	}
}

func TestComponentInstance_ExtractPortInfo(t *testing.T) {
	type args struct {
		portList []int
	}
	tests := []struct {
		name   string
		args   args
		want   string
	}{
		{"normal", args{[]int{1,2,3}}, "[1,2,3]"},
		{"normal", args{[]int{}}, "[]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c := &ComponentInstance{
				PortList: tt.args.portList,
			}
			wants := c.SerializePortInfo()
			assert.Equal(t, wants, tt.want)
		})
	}
}