
/******************************************************************************
 * Copyright (c)  2021 PingCAP                                                *
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
	"github.com/micro/cli/v2"
	"testing"
)

func TestAllFlags(t *testing.T) {
	type args struct {
		receiver *ClientArgs
	}
	tests := []struct {
		name string
		args args
		asserts []func(receiver *ClientArgs, fs []cli.Flag) bool
	}{
		{"normal", args{&ClientArgs{}}, []func(receiver *ClientArgs, fs []cli.Flag) bool{
			func(receiver *ClientArgs, fs []cli.Flag) bool {return len(fs) > 4},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllFlags(tt.args.receiver)
			for i,v := range tt.asserts {
				if !v(tt.args.receiver, got) {
					t.Errorf("AllFlags assert false, assert index = %v, got = %v", i, got)
				}
			}

		})
	}
}
