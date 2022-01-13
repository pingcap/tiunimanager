package handler

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pingcap-inc/tiem/library/framework"
	"math/rand"
)

// GetRandomString get random password
func GetRandomString(n int) string {
	randBytes := make([]byte, n/2)
	rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}


func ExecCommandThruSQL(ctx context.Context, db *sql.DB, sqlCommand string) error {
	logInFunc := framework.LogWithContext(ctx)
	_, err := db.Exec(sqlCommand)
	if err != nil {
		logInFunc.Errorf("execute sql command %s error: %s", sqlCommand, err.Error())
		return err
	}
	return nil
}