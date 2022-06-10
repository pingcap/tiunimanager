package sql

import (
	"context"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pingcap-inc/tiunimanager/common/constants"
	"github.com/pingcap-inc/tiunimanager/micro-cluster/cluster/management/meta"
	"github.com/pingcap-inc/tiunimanager/models/cluster/management"
	"github.com/pingcap-inc/tiunimanager/models/common"
	"strings"
	"testing"
	"time"
)

func TestExecCommandThruSQL(t *testing.T) {
	sqlCommand := "ALTER USER 'root'@'%' IDENTIFIED BY '12345678'"
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec(sqlCommand).
		WillReturnError(fmt.Errorf("some error"))
	mock.ExpectRollback()

	err = ExecCommandThruSQL(context.TODO(), db, sqlCommand)
	if err == nil || !strings.Contains(err.Error(), "some error") {
		t.Errorf("err(%s) should contain 'some error'", err.Error())
	}
}

var dbConnParam1 DbConnParam
var dbConnParam2 DbConnParam

func init() {
	dbConnParam1 = DbConnParam{
		Username: "root",
		Password: "",
		IP:       "127.0.0.1",
		Port:     "4000",
	}
	dbConnParam2 = DbConnParam{
		Username: "root",
		Password: "12345678",
		IP:       "127.0.0.1",
		Port:     "4000",
	}
}

func TestDBUserManager_CreateDBUser(t *testing.T) {
	UpdateDBUserPassword(context.TODO(), dbConnParam1, "root", "12345678", "testworkflownodeid")
	dbConnParam1.Password = "12345678"
	user1 := &management.DBUser{
		ClusterID: "clusterID",
		Name:      "backup",
		Password:  common.PasswordInExpired{Val: meta.GetRandomString(10), UpdateTime: time.Now()},
		RoleType:  string(constants.DBUserBackupRestore),
	}
	type args struct {
		ctx            context.Context
		connec         DbConnParam
		user           *management.DBUser
		workFlowNodeID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"normal", args{context.TODO(), dbConnParam1, user1, "testworkflownodeid"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateDBUser(tt.args.ctx, tt.args.connec, tt.args.user, tt.args.workFlowNodeID); (err != nil) != tt.wantErr {
				t.Errorf("CreateDBUser() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				fmt.Println(err)
			}
		})
	}
}

func TestDBUserManager_UpdateDBUserPassword(t *testing.T) {
	err := UpdateDBUserPassword(context.TODO(), dbConnParam1, "root", "12345678", "22")
	fmt.Println(err)
	if err == nil {
		t.Error("err nil")
	}
}

func TestDBUserManager_DeleteDBUser(t *testing.T) {
	user1 := &management.DBUser{
		ClusterID: "clusterID",
		Name:      "backup",
		Password:  common.PasswordInExpired{Val: meta.GetRandomString(10)},
		RoleType:  string(constants.DBUserBackupRestore),
	}
	CreateDBUser(context.TODO(), dbConnParam1, user1, "22")
	err := DeleteDBUser(context.TODO(), dbConnParam1, "test", "22")
	fmt.Println(err)
	if err == nil {
		t.Error("err nil")
	}
}
