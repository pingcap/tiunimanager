package account

import (
	"github.com/pingcap-inc/tiem/models/common"
	"testing"
)

func TestAccount_GenSaltAndHash(t *testing.T) {
	type args struct {
		passwd string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"normal1", args{passwd: "Test12345678"}, false},
		{"normal2", args{passwd: "Testttttttt1111"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &Account{
				Entity:    common.Entity{TenantId: "testID", Status: "0"}, //todo: bug
				Name:      "testName",
				Salt:      "123",
				FinalHash: "1234",
			}
			if err := account.GenSaltAndHash(tt.args.passwd); (err != nil) != tt.wantErr {
				t.Errorf("GenSaltAndHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAccount_CheckPassword(t *testing.T) {
	account := &Account{
		Entity:    common.Entity{TenantId: "testID", Status: "0"}, //todo: bug
		Name:      "testName",
		Salt:      "123",
		FinalHash: "1234",
	}
	account.GenSaltAndHash("Test12345678")

	type args struct {
		passwd string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"normal", args{passwd: "Test12345678"}, true, false},
		{"empty password", args{passwd: ""}, false, true},
		{"long password", args{passwd: "Test123456789123456789"}, false, true},
		{"wrong password", args{passwd: "Testtest12345678"}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := account.CheckPassword(tt.args.passwd)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("CheckPassword() error = %v, got = %v, wantErr %v, want %v", err, got, tt.wantErr, tt.want)
				return
			}
		})
	}
}