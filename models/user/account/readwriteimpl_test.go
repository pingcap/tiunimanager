package account

import (
	"context"
	"github.com/pingcap-inc/tiem/models/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccountReadWrite_AddAccount(t *testing.T) {
	type args struct {
		ctx       context.Context
		tenantId  string
		name      string
		salt      string
		finalHash string
		status    int8
	}
	tests := []struct {
		name    string
		args    args
		wantErr    bool
	}{
		{"normal", args{context.TODO(), "testID1", "testName1", "123", "15", 0}, false},
		{"without tenantID", args{context.TODO(), "", "testName2", "12345", "234", 0}, true},
		{"without name", args{context.TODO(), "testID3", "", "12345", "234", 0}, true},
		{"without salt", args{context.TODO(), "testID4", "testName4", "", "234", 0}, true},
		{"without finalHash", args{context.TODO(), "testID5", "testName5", "12345", "", 0}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := testRW.AddAccount(tt.args.ctx, tt.args.tenantId, tt.args.name, tt.args.salt, tt.args.finalHash, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestAccountReadWrite_FindAccountByName(t *testing.T) {
	account := &Account{
		Entity:    common.Entity{TenantId: "testID", Status: "0"}, //todo: bug
		Name:      "testName",
		Salt:      "123",
		FinalHash: "1234",
	}
	testRW.DB(context.TODO()).Create(account)
	defer testRW.DB(context.TODO()).Delete(account)

	type args struct {
		ctx context.Context
		name  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), account.Name}, false},
		{"no record", args{context.TODO(), "findName"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := testRW.FindAccountByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAccountById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestAccountReadWrite_FindAccountById(t *testing.T) {
	account := &Account{
		Entity:    common.Entity{TenantId: "testID", Status: "0"}, //todo: bug
		Name:      "testName",
		Salt:      "123",
		FinalHash: "1234",
	}
	testRW.DB(context.TODO()).Create(account)
	defer testRW.DB(context.TODO()).Delete(account)

	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal", args{context.TODO(), account.ID}, false},
		{"no record", args{context.TODO(), "findID"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := testRW.FindAccountById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAccountById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				assert.NotEmpty(t, got.ID)
			}
		})
	}
}

func TestAccountReadWrite_FindAccountById_v2(t *testing.T) {
	account := &Account{
		Entity:    common.Entity{TenantId: "testID", Status: "0"}, //todo: bug
		Name:      "testName",
		Salt:      "123",
		FinalHash: "1234",
	}
	testRW.DB(context.TODO()).Create(account)
	defer testRW.DB(context.TODO()).Delete(account)


	t.Run("normal", func(t *testing.T) {
		err := testRW.DB(context.TODO()).Where("id = ?", account.ID).First(account).Error
		assert.NoError(t, err)
		_, err = testRW.FindAccountById(context.TODO(), account.ID)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := testRW.FindAccountById(context.TODO(), "whatever")
		assert.Error(t, err)

		_, err = testRW.FindAccountById(context.TODO(), "")
		assert.Error(t, err)
	})
}
