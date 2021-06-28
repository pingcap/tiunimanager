package domain

import "testing"

func TestAccessible(t *testing.T) {
	setupMockAdapter()

	type args struct {
		pathType    string
		path        string
		tokenString string
	}
	tests := []struct {
		name            string
		args            args
		wantTenantId    uint
		wantAccountName string
		wantErr         bool
		prepareType 	int // 0 什么都不做，1 登录，2 退出，3登录再退出
	}{
		{"test accessible without token", args{path: testPath1, pathType: "type1"}, 0, "", true, 0},
		{"test accessible with wrong token", args{tokenString: "wrong token", path: testPath1, pathType: "type1"}, 0, "", true, 0},
		{"test accessible with invalid token", args{tokenString: testMyName, path: testPath1, pathType: "type1"}, 1, testMyName, true, 3},
		{"test accessible normal", args{tokenString: testMyName, path: testPath1, pathType: "type1"}, 1, testMyName, false, 1},

		{"test accessible without path", args{tokenString: testMyName, pathType: "type1"}, 0, "", true, 1},
		{"test accessible with wrong path", args{tokenString: testMyName, path: "whatever", pathType: "type1"}, 1, testMyName, true, 1},
		{"test accessible without permission1", args{tokenString: testMyName, path: testPath2, pathType: "type1"}, 1, testMyName, true, 1},
		{"test accessible without permission2", args{tokenString: testOtherName, path: testPath1, pathType: "type1"}, 1, testOtherName, true, 1},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.prepareType {
			case 1:
				if tt.args.tokenString == testMyName {
					Login(testMyName, testMyPassword)
				}
				if tt.args.tokenString == testOtherName {
					Login(testOtherName, testOtherPassword)
				}
			case 2:
				Logout(testMyName)
			case 3:
				Login(testMyName, testMyPassword)
				Logout(testMyName)
			default:
			}
			gotTenantId, gotAccountName, err := Accessible(tt.args.pathType, tt.args.path, tt.args.tokenString)
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
		err error
		want string
	}{
		{"testError", &UnauthorizedError{}, "Unauthorized"},
		{"testError", &ForbiddenError{}, "Access Forbidden"},
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
	setupMockAdapter()
	type args struct {
		userName string
		password string
	}
	tests := []struct {
		name            string
		args            args
		want			string
		wantErr         bool
	}{
		{"test login without argument", args{}, "", true},
		{"test login empty password", args{userName: testMyName, password: ""}, "", true},
		{"test login wrong password", args{userName: testMyName, password: "sdfa"}, "", true},
		{"test login normal", args{userName: testMyName, password: testMyPassword}, testMyName, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTokenString, err := Login(tt.args.userName, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotTokenString != tt.want {
				t.Errorf("Login() got = %v, want %v", gotTokenString, tt.want)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	setupMockAdapter()
	Login(testMyName, testMyPassword)
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
		{"test logout normal", args{testMyName}, testMyName, false},
		{"test logout with wrong token", args{"testtttt"}, "", true},
		{"test logout with wrong token", args{testMyName}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Logout(tt.args.tokenString)
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
