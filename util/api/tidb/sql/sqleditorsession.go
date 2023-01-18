package sql

import (
	"context"
	"database/sql"
	"time"

	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
)

func CreateSession(ctx context.Context, clusterID string, expireSec uint64, database string) (sessionID string, err error) {
	clusterMeta, err := meta.Get(ctx, clusterID)
	if err != nil {
		return

	}
	db, err := meta.CreateSQLLinkWithDatabase(ctx, clusterMeta, database)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Second * time.Duration(expireSec))
	conn, err := db.Conn(ctx)
	if err != nil {
		return
	}
	sessionID = CreateSessionCache(conn, expireSec)
	return sessionID, nil
}

func CloseSession(ctx context.Context, sessionID string) error {
	conn, err := GetSessionFromCache(sessionID)
	if err != nil {
		return err
	}
	if conn != nil {
		conn.Close()
	}
	return CloseSessionCache(sessionID)
}

func GetSession(ctx context.Context, sessionID string) (*sql.Conn, error) {
	return GetSessionFromCache(sessionID)
}
