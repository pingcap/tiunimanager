package domain

import (
	"testing"
)

func TestAccount_checkPassword(t *testing.T) {
	ac := Account{}
	ac.GenSaltAndHash("testMyPassword")
	type fields struct {
		Id        string
		TenantId  string
		Name      string
		Salt      string
		FinalHash string
		Status    CommonStatus
	}
	type args struct {
		passwd string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			"testNormal",
			fields{
				Salt:      ac.Salt,
				FinalHash: ac.FinalHash,
			},
			args{
				"testMyPassword",
			},
			true,
			false,
		},
		{
			"testWrongPassword",
			fields{
				Salt:      ac.Salt,
				FinalHash: ac.FinalHash,
			},
			args{
				"wrongPassword",
			},
			false,
			false,
		},
		{
			"testEmptyPassword",
			fields{
				Salt:      ac.Salt,
				FinalHash: ac.FinalHash,
			},
			args{
				"",
			},
			false,
			true,
		},
		{
			"testTooLongPassword",
			fields{
				Salt:      ac.Salt,
				FinalHash: ac.FinalHash,
			},
			args{
				"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &Account{
				Id:        tt.fields.Id,
				TenantId:  tt.fields.TenantId,
				Name:      tt.fields.Name,
				Salt:      tt.fields.Salt,
				FinalHash: tt.fields.FinalHash,
				Status:    tt.fields.Status,
			}
			got, err := account.CheckPassword(tt.args.passwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckPassword() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccount_genSaltAndHash(t *testing.T) {
	type fields struct {
		Id        string
		TenantId  string
		Name      string
		Salt      string
		FinalHash string
		Status    CommonStatus
	}
	type args struct {
		passwd string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"testNormal",
			fields{},
			args{
				"normal",
			},
			false,
		},
		{
			"testPasswordEmpty",
			fields{},
			args{
				"",
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account := &Account{
				Id:        tt.fields.Id,
				TenantId:  tt.fields.TenantId,
				Name:      tt.fields.Name,
				Salt:      tt.fields.Salt,
				FinalHash: tt.fields.FinalHash,
				Status:    tt.fields.Status,
			}
			if err := account.GenSaltAndHash(tt.args.passwd); (err != nil) != tt.wantErr {
				t.Errorf("GenSaltAndHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if account.Salt == "" {
					t.Errorf("GenSaltAndHash() got empty Salt")
					return
				}
				if account.FinalHash == "" {
					t.Errorf("GenSaltAndHash() got empty FinalHash")
					return
				}
			}
		})
	}
}

func Test_finalHash(t *testing.T) {
	type args struct {
		salt   string
		passwd string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{

		{"testNormal",
			args{
				salt:   "1111111",
				passwd: "testMyPassword",
			},
			false},
		{"testEmptyPassword",
			args{
				salt:   "1111111",
				passwd: "",
			},
			true},
		{"testEmptySalt",
			args{
				salt:   "",
				passwd: "testMyPassword",
			},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := finalHash(tt.args.salt, tt.args.passwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("finalHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got == nil {
				t.Errorf("finalHash() got nil")
				return
			}
			if err == nil && len(got) == 0 {
				t.Errorf("finalHash() got empty")
				return
			}
		})
	}
}
