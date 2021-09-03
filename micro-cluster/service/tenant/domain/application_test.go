package domain

import "testing"

func TestAccessible(t *testing.T) {
	myToken, _ := Login(testMyName, testMyPassword)
	otherToken, _ := Login(testOtherName, testOtherPassword)
	invalidToken, _ := Login(testMyName, testMyPassword)
	Logout(invalidToken)
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
		{"test accessible normal", args{tokenString: myToken, path: testPath1, pathType: "type1"}, "1", testMyName, false},
		{"test accessible without token", args{path: testPath1, pathType: "type1"}, "", "", true},
		{"test accessible with wrong token", args{tokenString: "wrong token", path: testPath1, pathType: "type1"}, "", "", true},
		{"test accessible with invalid token", args{tokenString: invalidToken, path: testPath1, pathType: "type1"}, "1", testMyName, true},
		{"test accessible without path", args{tokenString: myToken, pathType: "type1"}, "", "", true},
		{"test accessible with wrong path", args{tokenString: myToken, path: "whatever", pathType: "type1"}, "1", testMyName, true},
		{"test accessible without permission1", args{tokenString: myToken, path: testPath2, pathType: "type1"}, "1", testMyName, true},
		{"test accessible without permission2", args{tokenString: otherToken, path: testPath1, pathType: "type1"}, "1", testOtherName, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTenantId, _, gotAccountName, err := Accessible(tt.args.pathType, tt.args.path, tt.args.tokenString)
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
	type args struct {
		userName string
		password string
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
	}{
		{"test login without argument", args{}, true},
		{"test login empty password", args{userName: testMyName, password: ""}, true},
		{"test login wrong password", args{userName: testMyName, password: "sdfa"}, true},
		{"test login normal", args{userName: testMyName, password: testMyPassword}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTokenString, err := Login(tt.args.userName, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			Logout(gotTokenString)
		})
	}
}

func TestLogout(t *testing.T) {
	tokenString, _ := Login(testMyName, testMyPassword)
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
		{"test logout normal", args{tokenString}, testMyName, false},
		{"test logout with wrong token", args{"testtttt"}, "", true},
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
