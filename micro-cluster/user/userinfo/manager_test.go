package userinfo

import (
	"context"
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/user/tenant"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var manager = &Manager{}

func TestMain(m *testing.M) {
	var testFilePath string
	framework.InitBaseFrameworkForUt(framework.ClusterService,
		func(d *framework.BaseFramework) error {
			testFilePath = d.GetDataDir()
			os.MkdirAll(testFilePath, 0755)

			return models.Open(d, false)
		},
	)
	code := m.Run()
	os.RemoveAll(testFilePath)

	os.Exit(code)
}

func TestManager_CreateAccount(t *testing.T) {
	te, _ := models.GetTenantReaderWriter().AddTenant(context.TODO(), "tenant", 0, 0)
	type args struct {
		ctx    context.Context
		tenant *tenant.Tenant
		name   string
		passwd string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), te, "testName", "123456789"}, false},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.CreateAccount(tt.args.ctx, tt.args.tenant, tt.args.name, tt.args.passwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestManager_FindAccountByName(t *testing.T) {
	te, _ := models.GetTenantReaderWriter().AddTenant(context.TODO(), "tenant", 0, 0)
	manager.CreateAccount(context.TODO(), te, "testName", "123456789")
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), "testName"}, false},
		{"no record", args{context.TODO(), "ttt"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := manager.FindAccountByName(tt.args.ctx, tt.args.name)
			fmt.Println(err)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAccountByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}