package application

import (
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/adapt"
	"github.com/pingcap-inc/tiem/micro-cluster/service/user/domain"
	"reflect"
	"testing"
)

func TestAccessible(t *testing.T) {
	myToken, _ := authManager.Login(adapt.TestMyName, adapt.TestMyPassword)
	otherToken, _ := authManager.Login(adapt.TestOtherName, adapt.TestOtherPassword)
	invalidToken, _ := authManager.Login(adapt.TestMyName, adapt.TestMyPassword)
	authManager.Logout(invalidToken)
	SkipAuth = false

	type args struct {
		pathType    string
		path        string
		tokenString string
	}
	tests := []struct {
		name            string
		args            args
		wantTenantId    string
		wantAccountName string
		wantErr         bool
	}{
		{"test accessible normal", args{tokenString: myToken, path: adapt.TestPath1, pathType: "type1"}, "1", adapt.TestMyName, false},
		{"test accessible without token", args{path: adapt.TestPath1, pathType: "type1"}, "", "", true},
		{"test accessible with wrong token", args{tokenString: "wrong token", path: adapt.TestPath1, pathType: "type1"}, "", "", true},
		{"test accessible with invalid token", args{tokenString: invalidToken, path: adapt.TestPath1, pathType: "type1"}, "1", adapt.TestMyName, true},
		{"test accessible without path", args{tokenString: myToken, pathType: "type1"}, "", "", true},
		{"test accessible with wrong path", args{tokenString: myToken, path: "whatever", pathType: "type1"}, "1", adapt.TestMyName, true},
		{"test accessible without permission1", args{tokenString: myToken, path: adapt.TestPath2, pathType: "type1"}, "1", adapt.TestMyName, true},
		{"test accessible without permission2", args{tokenString: otherToken, path: adapt.TestPath1, pathType: "type1"}, "1", adapt.TestOtherName, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTenantId, _, gotAccountName, err := authManager.Accessible(tt.args.pathType, tt.args.path, tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("Accessible() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTenantId != tt.wantTenantId {
				t.Errorf("Accessible() gotTenantId = %v, want %v", gotTenantId, tt.wantTenantId)
			}
			if gotAccountName != tt.wantAccountName {
				t.Errorf("Accessible() gotAccountName = %v, want %v", gotAccountName, tt.wantAccountName)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"testError", &domain.UnauthorizedError{}, "Unauthorized"},
		{"testError", &domain.ForbiddenError{}, "Access Forbidden"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	type args struct {
		userName string
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"test login without argument", args{}, true},
		{"test login empty password", args{userName: adapt.TestMyName, password: ""}, true},
		{"test login wrong password", args{userName: adapt.TestMyName, password: "sdfa"}, true},
		{"test login normal", args{userName: adapt.TestMyName, password: adapt.TestMyPassword}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTokenString, err := authManager.Login(tt.args.userName, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			authManager.Logout(gotTokenString)
		})
	}
}

func TestLogout(t *testing.T) {
	tokenString, _ := authManager.Login(adapt.TestMyName, adapt.TestMyPassword)
	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"test logout without token", args{}, "", true},
		{"test logout normal", args{tokenString}, adapt.TestMyName, false},
		{"test logout with wrong token", args{"testtttt"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := authManager.Logout(tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("Logout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Logout() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkAuth(t *testing.T) {
	type args struct {
		account    *domain.AccountAggregation
		permission *domain.PermissionAggregation
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
				&domain.AccountAggregation{
					Roles: []domain.Role{
						{
							Id: "1",
						},
						{
							Id: "2",
						},
					},
				},
				&domain.PermissionAggregation{
					Roles: []domain.Role{
						{
							Id: "1",
						},
						{
							Id: "5",
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
				&domain.AccountAggregation{
					Roles: []domain.Role{
						{
							Id: "4",
						},
						{
							Id: "3",
						},
						{
							Id: "2",
						},
						{
							Id: "1",
						},
					},
				},
				&domain.PermissionAggregation{
					Roles: []domain.Role{
						{
							Id: "9",
						},
						{
							Id: "1",
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
				&domain.AccountAggregation{
					Roles: []domain.Role{
						{
							Id: "1",
						},
						{
							Id: "2",
						},
					},
				},
				&domain.PermissionAggregation{
					Roles: []domain.Role{
						{
							Id: "3",
						},
						{
							Id: "4",
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
				&domain.AccountAggregation{
					Roles: []domain.Role{},
				},
				&domain.PermissionAggregation{
					Roles: []domain.Role{
						{
							Id: "1",
						},
						{
							Id: "5",
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
				&domain.AccountAggregation{
					Roles: nil,
				},
				&domain.PermissionAggregation{
					Roles: []domain.Role{
						{
							Id: "1",
						},
						{
							Id: "5",
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
				&domain.AccountAggregation{
					Roles: []domain.Role{
						{
							Id: "1",
						},
						{
							Id: "2",
						},
					},
				},
				&domain.PermissionAggregation{
					Roles: []domain.Role{},
				},
			},
			false,
			false,
		},
		{
			"permissionWithoutRole2",
			args{
				&domain.AccountAggregation{
					Roles: []domain.Role{
						{
							Id: "1",
						},
						{
							Id: "2",
						},
					},
				},
				&domain.PermissionAggregation{
					Roles: nil,
				},
			},
			false,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := authManager.checkAuth(tt.args.account, tt.args.permission)
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

func Test_findAccountAggregation(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.AccountAggregation
		wantErr bool
	}{
		{name: "test empty name", args: args{""}, want: nil, wantErr: true},
		{name: "test noaccount", args: args{"noaccount"}, want: nil, wantErr: true},
		{name: "test account", args: args{"testMyName"}, want: &domain.AccountAggregation{
			Account: domain.Account{
				Name:     "testMyName",
				Id:       "1",
				TenantId: "1",
			},
			Roles: []domain.Role{
				{Id: "1", Name: "admin", TenantId: "1"},
				{Id: "2", Name: "dba", TenantId: "1"},
			},
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := userManager.findAccountAggregation(tt.args.name)
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