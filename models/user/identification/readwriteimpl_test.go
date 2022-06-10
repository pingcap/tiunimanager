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

package identification

import (
	"context"
	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTokenReadWrite_CreateToken(t *testing.T) {
	token := &Token{
		TokenString: "testToken",
		UserID:      "accountID",
		TenantID:    "tenantID",
	}
	testRW.DB(context.TODO()).Create(token)
	defer testRW.DB(context.TODO()).Delete(token)

	type args struct {
		ctx            context.Context
		tokenString    string
		accountName    string
		accountId      string
		tenantId       string
		expirationTime time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), "testToken1", "testName1", "accountID", "tenantID", time.Unix(2, 56).Add(constants.DefaultTokenValidPeriod)}, false},
		{"without tokenString", args{context.TODO(), "", "testName2", "accountID", "tenantID", time.Unix(2, 56).Add(constants.DefaultTokenValidPeriod)}, true},
		{"already exist", args{context.TODO(), "testToken", "accountName", "accountID", "tenantID", time.Unix(2, 56).Add(constants.DefaultTokenValidPeriod)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testRW.CreateToken(tt.args.ctx, tt.args.tokenString, tt.args.accountId, tt.args.tenantId, tt.args.expirationTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestTokenReadWrite_GetToken(t *testing.T) {
	token := &Token{
		TokenString: "token",
		UserID:      "accountID",
		TenantID:    "tenantID",
	}
	testRW.DB(context.TODO()).Create(token)
	defer testRW.DB(context.TODO()).Delete(token)

	type args struct {
		ctx         context.Context
		tokenString string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), "token"}, false},
		{"null tokenString, no record", args{context.TODO(), ""}, true},
		{"no record", args{context.TODO(), "tokentest"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testRW.GetToken(tt.args.ctx, tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}
