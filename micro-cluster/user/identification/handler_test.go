package identification

import (
	"context"
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/message"
	"github.com/pingcap-inc/tiem/models"
	"github.com/pingcap-inc/tiem/models/user/identification"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

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

func TestProvide(t *testing.T) {
	token := identification.Token{
		AccountName: "testName",
		AccountId: "bdhjdbjfsjbd",
		TenantId: "hdsfksdj",
		ExpirationTime: time.Now().Add(constants.DefaultTokenValidPeriod),
	}
	type args struct {
		ctx     context.Context
		request message.ProvideTokenReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), message.ProvideTokenReq{Token: &token}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Provide(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.TokenString)
			}
		})
	}
}

func TestModify(t *testing.T) {
	models.GetTokenReaderWriter().AddToken(context.TODO(),
		"66d0bcea-30da-4313-9f47-8adf60e3115c",
		"accoutName",
		"accountID",
		"tenantID",
		time.Now().Add(constants.DefaultTokenValidPeriod),
		)

	type args struct {
		ctx     context.Context
		request message.ModifyTokenReq
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), message.ModifyTokenReq{
			Token: &identification.Token{
				TokenString: "66d0bcea-30da-4313-9f47-8adf60e3115c",
				TenantId: "tenant2",
			},
		}},
		false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Modify(tt.args.ctx, tt.args.request); (err != nil) != tt.wantErr {
				t.Errorf("Modify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetToken(t *testing.T) {
	models.GetTokenReaderWriter().AddToken(context.TODO(),
		"66d0bcea-30da-4313-9f47-8adf60e3115c",
		"accoutName",
		"accountID",
		"tenantID",
		time.Now().Add(constants.DefaultTokenValidPeriod),
	)

	type args struct {
		ctx     context.Context
		request message.GetTokenReq
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
	}{
		{"normal", args{context.TODO(), message.GetTokenReq{TokenString: "66d0bcea-30da-4313-9f47-8adf60e3115c"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetToken(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.NotEmpty(t, got.TokenString)
			}
		})
	}
}