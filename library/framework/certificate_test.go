
/******************************************************************************
 * Copyright (c)  2021 PingCAP                                               **
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
	"github.com/pingcap/tiunimanager/common/constants"
	"reflect"
	"testing"
)

func TestNewCertificateFromArgs(t *testing.T) {
	type args struct {
		args *ClientArgs
	}
	tests := []struct {
		name string
		args args
		want *CertificateInfo
	}{
		{"normal", args{&ClientArgs{DeployDir: "aaaa"}}, &CertificateInfo{
			CertificateCrtFilePath: "aaaa" + constants.CertDirPrefix + constants.CertFileName,
			CertificateKeyFilePath: "aaaa" + constants.CertDirPrefix + constants.KeyFileName,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCertificateFromArgs(tt.args.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCertificateFromArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
