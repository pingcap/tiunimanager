package domain
import (
	"reflect"
	"testing"
)


func TestAccount_checkPassword(t *testing.T) {
	ac := Account{
	}
	ac.genSaltAndHash(testMyPassword)
	type fields struct {
		Id        uint
		TenantId  uint
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
				Salt: ac.Salt,
				FinalHash: ac.FinalHash,
			},
			args{
				testMyPassword,
			},
			true,
			false,
		},
		{
			"testWrongPassword",
			fields{
				Salt: ac.Salt,
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
				Salt: ac.Salt,
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
				Salt: ac.Salt,
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
			got, err := account.checkPassword(tt.args.passwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkPassword() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccount_genSaltAndHash(t *testing.T) {
	type fields struct {
		Id        uint
		TenantId  uint
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
			if err := account.genSaltAndHash(tt.args.passwd); (err != nil) != tt.wantErr {
				t.Errorf("genSaltAndHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if account.Salt == "" {
					t.Errorf("genSaltAndHash() got empty Salt")
					return
				}
				if account.FinalHash == "" {
					t.Errorf("genSaltAndHash() got empty FinalHash")
					return
				}
			}
		})
	}
}

func TestAccount_persist(t *testing.T) {
	type fields struct {
		Id        uint
		TenantId  uint
		Name      string
		Salt      string
		FinalHash string
		Status    CommonStatus
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
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
			if err := account.persist(); (err != nil) != tt.wantErr {
				t.Errorf("persist() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateAccount(t *testing.T) {
	type args struct {
		tenant *Tenant
		name   string
		passwd string
	}
	tests := []struct {
		name    string
		args    args
		want    *Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateAccount(tt.args.tenant, tt.args.name, tt.args.passwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateAccount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkAuth(t *testing.T) {
	type args struct {
		account    *AccountAggregation
		permission *PermissionAggregation
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			"withPermission1",
			args{
				&AccountAggregation{
					Roles: []Role{
						{
							Id: 1,
						},
						{
							Id: 2,
						},
					},
				},
				&PermissionAggregation{
					Roles: []Role{
						{
							Id: 1,
						},
						{
							Id: 5,
						},
					},
				},
			},
			true,
			false,
		},
		{
			"withPermission2",
			args{
				&AccountAggregation{
					Roles: []Role{
						{
							Id: 4,
						},
						{
							Id: 3,
						},
						{
							Id: 2,
						},
						{
							Id: 1,
						},
					},
				},
				&PermissionAggregation{
					Roles: []Role{
						{
							Id: 9,
						},
						{
							Id: 1,
						},
					},
				},
			},
			true,
			false,
		},
		{
			"noPermission",
			args{
				&AccountAggregation{
					Roles: []Role{
						{
							Id: 1,
						},
						{
							Id: 2,
						},
					},
				},
				&PermissionAggregation{
					Roles: []Role{
						{
							Id: 3,
						},
						{
							Id: 4,
						},
					},
				},
			},
			false,
			false,
		},
		{
			"userWithoutRole1",
			args{
				&AccountAggregation{
					Roles: []Role{},
				},
				&PermissionAggregation{
					Roles: []Role{
						{
							Id: 1,
						},
						{
							Id: 5,
						},
					},
				},
			},
			false,
			false,
		},
		{
			"userWithoutRole2",
			args{
				&AccountAggregation{
					Roles: nil,
				},
				&PermissionAggregation{
					Roles: []Role{
						{
							Id: 1,
						},
						{
							Id: 5,
						},
					},
				},
			},
			false,
			false,
		},
		{
			"permissionWithoutRole1",
			args{
				&AccountAggregation{
					Roles: []Role{
						{
							Id: 1,
						},
						{
							Id: 2,
						},
					},
				},
				&PermissionAggregation{
					Roles: []Role{

					},
				},
			},
			false,
			false,
		},
		{
			"permissionWithoutRole2",
			args{
				&AccountAggregation{
					Roles: []Role{
						{
							Id: 1,
						},
						{
							Id: 2,
						},
					},
				},
				&PermissionAggregation{
					Roles: nil,
				},
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkAuth(tt.args.account, tt.args.permission)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkAuth() got = %v, want %v", got, tt.want)
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
				passwd: testMyPassword,
			},
			false},
		{"testEmptyPassword",
			args{
				salt: "1111111",
				passwd: "",
			},
			true},
		{"testEmptySalt",
			args{
				salt:   "",
				passwd: testMyPassword,
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

func Test_findAccountAggregation(t *testing.T) {
	setupMockAdapter()
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *AccountAggregation
		wantErr bool
	}{
		{name: "test empty name", args: args{""}, want: nil, wantErr: true},
		{name: "test noaccount", args: args{"noaccount"}, want: nil, wantErr: true},
		{name: "test account", args: args{testMyName}, want: &AccountAggregation{
			Account: Account{
				Name:     testMyName,
				Id:       1,
				TenantId: 1,
			},
			Roles: []Role{
				{Id: 1, Name: "admin",TenantId: 1},
				{Id: 2, Name: "dba",TenantId: 1},
			},
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findAccountAggregation(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("findAccountAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}
			if tt.want.Account.Id != got.Account.Id ||
				tt.want.Account.Name != got.Account.Name ||
				tt.want.Account.TenantId != got.Account.TenantId ||
				!reflect.DeepEqual(got.Roles, tt.want.Roles) {
				t.Errorf("findAccountAggregation() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findAccountByName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *Account
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findAccountByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("findAccountByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findAccountByName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
