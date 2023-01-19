/******************************************************************************
 * Copyright (c)  2023 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 ******************************************************************************/

package dbagent

import (
	"context"
	"database/sql"

	"github.com/pingcap/tiunimanager/message/dataapps/sqleditor"
)

type ReaderWriter interface {
	GetTableMetaData(ctx context.Context, clusterID string, dbName string, tableName string) (metaRes *sqleditor.MetaRes, err error)
	GetClusterMetaData(ctx context.Context, clusterID *sql.DB, isBrief bool, showSystemDBFlag bool) (dbmetaList []*sqleditor.DBMeta, err error)
	CreateSession(ctx context.Context, clusterID string, expireSec uint64, database string) (sessionID string, err error)
	CloseSession(ctx context.Context, sessionID string) error
	GetSession(ctx context.Context, sessionID string) (*sql.Conn, error)
}
