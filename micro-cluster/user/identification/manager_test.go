package identification

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-cluster/user/userinfo"
	"github.com/pingcap-inc/tiem/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var manager = &Manager{}
var ma = &userinfo.Manager{}

func TestManager_Login_v1(t *testing.T) {
	te, _ := models.GetTenantReaderWriter().AddTenant(context.TODO(), "tenant", 0, 0)
	_, err := ma.CreateAccount(context.TODO(), te, "testName", "123456789")
	assert.Nil(t, err)

	type args struct {
		ctx     context.Context
		request message.LoginReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "normal", args: args{ctx: context.TODO(), request: message.LoginReq{UserName: "testName", Password: "123456789"}}, wantErr: false},
		{name: "wrong username", args: args{ctx: context.TODO(), request: message.LoginReq{UserName: "name", Password: "123456789"}}, wantErr: true},
		{name: "wrong password", args: args{ctx: context.TODO(), request: message.LoginReq{UserName: "testName", Password: "12345"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResp, err := manager.Login(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, gotResp.TokenString)
			}
		})
	}
}

func TestManager_Logout(t *testing.T) {
	te, _ := models.GetTenantReaderWriter().AddTenant(context.TODO(), "tenant", 0, 0)
	_, err := ma.CreateAccount(context.TODO(), te, "test", "123456789")
	assert.Nil(t, err)

	tokenString, _ := manager.Login(context.TODO(), message.LoginReq{UserName: "test", Password: "123456789"})
	type args struct {
		ctx context.Context
		req message.LogoutReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "normal", args: args{ctx: context.TODO(), req: message.LogoutReq{TokenString: tokenString.TokenString}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.Logout(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Logout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.AccountName)
			}
		})
	}
}

func TestManager_Accessible(t *testing.T) {
	models.GetTokenReaderWriter().AddToken(context.TODO(), "&vhgjkgsjksdas", "account", "accountID", "tenantID", time.Unix(2, 56).Add(constants.DefaultTokenValidPeriod))
	models.GetTokenReaderWriter().AddToken(context.TODO(), "token", "account1", "accountID1", "tenantID1", time.Now().Add(constants.DefaultTokenValidPeriod))
	type args struct {
		ctx     context.Context
		request message.AccessibleReq
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
	}{
		{"normal", args{ctx: context.TODO(), request: message.AccessibleReq{
			PathType: "type",
			Path: "path",
			TokenString: "token",
		}}, false},
		{"invalid token", args{ctx: context.TODO(), request: message.AccessibleReq{
			PathType: "type",
			Path: "path",
			TokenString: "&vhgjkgsjksdas",
		}}, true},
		{"token not found", args{ctx: context.TODO(), request: message.AccessibleReq{
			PathType: "type",
			Path: "path",
			TokenString: "&vhgjs",
		}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResp, err := manager.Accessible(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Accessible() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, gotResp.AccountID)
				assert.NotEmpty(t, gotResp.TenantID)
			}
		})
	}
}