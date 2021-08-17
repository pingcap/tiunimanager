package domain

import "testing"

var dashboardUrl string = "http://127.0.0.1:2379/dashboard/"
var usrName string = "root"
var password string = ""

func TestGetLoginToken(t *testing.T) {
	token, err := getLoginToken(dashboardUrl, usrName, password)
	if err != nil {
		t.Errorf("TestGetLoginToken failed, %s", err.Error())
	}
	t.Logf("TestGetLoginToken success, token: %s", token)
}

func TestGenerateShareCode(t *testing.T) {
	token, err := getLoginToken(dashboardUrl, usrName, password)
	if err != nil {
		t.Errorf("TestGenerateShareCode GetLoginToken failed, %s", err.Error())
	}
	t.Logf("token: %s", token)
	code, err := generateShareCode(dashboardUrl, token)
	if err != nil {
		t.Errorf("TestGenerateShareCode failed, %s", err.Error())
	}
	t.Logf("TestGenerateShareCode success, code: %s", code)
}