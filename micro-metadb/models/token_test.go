package models

import (
	"reflect"
	"testing"
	"time"
)

func TestAddToken(t *testing.T) {
	AddToken("existingToken", "", "", "111", time.Now().Add(10 * time.Second))

	type args struct {
		tokenString    string
		accountName    string
		accountId      string
		tenantId       string
		expirationTime time.Time
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wants 	  []func(a args, token Token) bool
	}{
		{"normal", args{tokenString: "aaa", tenantId: "111", expirationTime: time.Now().Add(5 * time.Second)}, false, []func(a args, token Token) bool{
			func(a args, token Token) bool{return token.ID > 0},
			func(a args, token Token) bool{return token.TokenString == "aaa"},
			func(a args, token Token) bool{return token.TenantId == "111"},
			func(a args, token Token) bool{return token.Status == 0},
			func(a args, token Token) bool{return token.ExpirationTime.After(time.Now().Add(4 * time.Second))},
		}},
		{"existed", args{tokenString: "existingToken", tenantId: "111", expirationTime: time.Now().Add(100 * time.Second)}, false, []func(a args, token Token) bool{
			func(a args, token Token) bool{return token.ID > 0},
			func(a args, token Token) bool{return token.TokenString == "existingToken"},
			func(a args, token Token) bool{return token.TenantId == "111"},
			func(a args, token Token) bool{return token.Status == 0},
			func(a args, token Token) bool{return token.ExpirationTime.After(time.Now().Add(50 * time.Second))},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := AddToken(tt.args.tokenString, tt.args.accountName, tt.args.accountId, tt.args.tenantId, tt.args.expirationTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for i, assert := range tt.wants {
				if !assert(tt.args, gotToken) {
					t.Errorf("AddToken() test error, testname = %v, assert %v, args = %v, gotToken = %v", tt.name, i, tt.args, gotToken)
				}
			}

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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := FindToken(tt.args.tokenString)
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
