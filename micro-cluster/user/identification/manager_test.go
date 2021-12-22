package identification

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/micro-cluster/user/userinfo"
	"github.com/pingcap-inc/tiem/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

var manager = &Manager{}
var ma = &userinfo.Manager{}

func TestManager_Login_v1(t *testing.T) {
	te, _ := models.GetTenantReaderWriter().AddTenant(context.TODO(), "tenant", 0, 0)
	ma.CreateAccount(context.TODO(), te, "testName", "123456789")
	type args struct {
		ctx     context.Context
		request message.LoginReq
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
	}{
		{"normal", args{context.TODO(), message.LoginReq{"testName", "123456789"}}, false},
		{"wrong username", args{context.TODO(), message.LoginReq{"name", "123456789"}}, true},
		{"wrong password", args{context.TODO(), message.LoginReq{"testName", "12345"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResp, err := manager.Login(tt.args.ctx, tt.args.request)
			fmt.Println(err)
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
	ma.CreateAccount(context.TODO(), te, "testName", "123456789")
	tokenString, _ := manager.Login(context.TODO(), message.LoginReq{"testName", "123456789"})
	type args struct {
		ctx context.Context
		req message.LogoutReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), message.LogoutReq{tokenString.TokenString}}, false},
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