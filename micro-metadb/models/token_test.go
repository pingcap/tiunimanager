
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

package models

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestAddToken(t *testing.T) {
	accountManager := Dao.AccountManager()
	accountManager.AddToken(context.TODO(), "existingToken", "", "", "111", time.Now().Add(10*time.Second))

	type args struct {
		tokenString    string
		//accountName    string
		//accountId      string
		tenantId       string
		expirationTime time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wants   []func(a args, token Token) bool
	}{
		{"normal", args{tokenString: "aaa", tenantId: "111", expirationTime: time.Now().Add(5 * time.Second)}, false, []func(a args, token Token) bool{
			func(a args, token Token) bool { return token.ID > 0 },
			func(a args, token Token) bool { return token.TokenString == "aaa" },
			func(a args, token Token) bool { return token.TenantId == "111" },
			func(a args, token Token) bool { return token.Status == 0 },
			func(a args, token Token) bool { return token.ExpirationTime.After(time.Now().Add(4 * time.Second)) },
		}},
		{"existed", args{tokenString: "existingToken", tenantId: "111", expirationTime: time.Now().Add(100 * time.Second)}, false, []func(a args, token Token) bool{
			func(a args, token Token) bool { return token.ID > 0 },
			func(a args, token Token) bool { return token.TokenString == "existingToken" },
			func(a args, token Token) bool { return token.TenantId == "111" },
			func(a args, token Token) bool { return token.Status == 0 },
			func(a args, token Token) bool { return token.ExpirationTime.After(time.Now().Add(50 * time.Second)) },
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//TODO
			/*gotToken, err := accountManager.AddToken(tt.args.tokenString, tt.args.accountName, tt.args.accountId, tt.args.tenantId, tt.args.expirationTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, *gotToken) {
					t.Errorf("AddToken() test error, testname = %v, assert %v, args = %v, gotToken = %v", tt.name, i, tt.args, gotToken)
				}
			}*/

		})
	}
}

func TestFindToken(t *testing.T) {
	type args struct {
		tokenString string
	}
	tests := []struct {
		name      string
		args      args
		wantToken Token
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	accountManager := Dao.AccountManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := accountManager.FindToken(context.TODO(), tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotToken, tt.wantToken) {
				t.Errorf("FindToken() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}
