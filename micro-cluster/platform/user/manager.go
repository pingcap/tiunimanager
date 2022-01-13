package user

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/library/secondparty"
	"github.com/pingcap-inc/tiem/micro-cluster/platform/user/handler"
	"github.com/pingcap-inc/tiem/models/platform/user"
	"strings"
)

type DBUserManager struct{}

func NewDBUserManager() *DBUserManager {
	return &DBUserManager{}
}

func (manager *DBUserManager) CreateDBUser(ctx context.Context, connec secondparty.DbConnParam, user user.DBUser, workFlowNodeID string) error {
	logInFunc := framework.LogWithContext(ctx).WithField("bizid", workFlowNodeID)
	logInFunc.Infof("createDBUser, user: %v, bizId: %s", user, workFlowNodeID)

	// connect database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", connec.Username, connec.Password, connec.IP, connec.Port))
	if err != nil {
		logInFunc.Error("conn tidb error", err)
		return err
	}
	defer db.Close()

	// execute sql command of creating user
	createSqlCommand := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'", user.Name, "%", user.Password)
	err = handler.ExecCommandThruSQL(ctx, db, createSqlCommand)
	if err != nil {
		return err
	}

	//	execute sql command of granting privileges to user
	grantSqlCommand := fmt.Sprintf("GRANT %s ON %s.%s TO %s@%s IDENTIFIED BY \"%s\"",
		strings.Join(user.Role.Permission, ","), user.Name, "%", "*", "*", user.Password)
	err = handler.ExecCommandThruSQL(ctx, db, grantSqlCommand)
	if err != nil {
		return err
	}

	return nil
}


func (manager *DBUserManager) UpdateDBUserPassword(ctx context.Context, connec secondparty.DbConnParam, name string, passwprd string, workFlowNodeID string) error {
	logInFunc := framework.LogWithContext(ctx).WithField("bizid", workFlowNodeID)
	logInFunc.Infof("UpdateDBUserPassword, name: %v, bizId: %s", name, workFlowNodeID)

	// connect database
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql", connec.Username, connec.Password, connec.IP, connec.Port))
	if err != nil {
		logInFunc.Error("conn tidb error", err)
		return err
	}
	defer db.Close()

	//execute sql command
	sqlCommand := fmt.Sprintf("ALTER USER '%s'@'%s' IDENTIFIED BY '%s'", name, "%", passwprd)
	err = handler.ExecCommandThruSQL(ctx, db, sqlCommand)

	if err != nil {
		return err
	}

	return nil
}


func (manager *DBUserManager) DeleteDBUser(ctx context.Context, connec secondparty.DbConnParam, name string, workFlowNodeID string) error {
	// todo
	return nil
}

