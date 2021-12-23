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

/**
 * @Author: guobob
 * @Description:
 * @File:  common_test.go
 * @Version: 1.0.0
 * @Date: 2021/12/10 11:40
 */

package utils

import "testing"

func TestTypeString(t *testing.T) {
	type args struct {
		t uint64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "EventHandshake",
			args: args{
				t: EventHandshake,
			},
			want: "EventHandshake",
		},
		{
			name: "EventQuit",
			args: args{
				t: EventQuit,
			},
			want: "EventQuit",
		},
		{
			name: "EventQuery",
			args: args{
				t: EventQuery,
			},
			want: "EventQuery",
		},
		{
			name: "EventStmtPrepare",
			args: args{
				t: EventStmtPrepare,
			},
			want: "EventStmtPrepare",
		},
		{
			name: "EventStmtExecute",
			args: args{
				t: EventStmtExecute,
			},
			want: "EventStmtExecute",
		},
		{
			name: "EventStmtClose",
			args: args{
				t: EventStmtClose,
			},
			want: "EventStmtClose",
		},
		{
			name: "UnKnownType",
			args: args{
				t: 100,
			},
			want: "UnKnownType",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TypeString(tt.args.t); got != tt.want {
				t.Errorf("TypeString() = %v, want %v", got, tt.want)
			}
		})
	}
}
