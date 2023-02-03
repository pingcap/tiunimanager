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
	"fmt"
	"sort"
	"strings"
	"time"

	"database/sql"

	"github.com/pingcap/tiunimanager/message/dataapps/sqleditor"
)

var (
	EXCLUDEDB       = []string{"__TiDB_BR_Temporary_mysql", "tidbcloud_dm_meta", "tidb_cdc", "lightning_task_info"}
	EXCLUDEDBANDSYS = []string{"__TiDB_BR_Temporary_mysql", "tidbcloud_dm_meta", "tidb_cdc", "lightning_task_info", "INFORMATION_SCHEMA", "mysql", "PERFORMANCE_SCHEMA", "METRICS_SCHEMA"}

	tableMetaQuery = "select * from information_schema.columns where table_schema =\"%s\"  and table_name = \"%s\";"
	dbSql          = "select \n  schema_name,\n  case schema_name\n  when 'INFORMATION_SCHEMA' then 1\n  when 'mysql' then 1\n  when 'PERFORMANCE_SCHEMA' then 1\n  else 2\n  end as t1\n  from information_schema.schemata group by schema_name order by t1,lower(schema_name);"
	briefmetaSql   = "select '' as t1,TABLE_SCHEMA,TABLE_NAME from information_schema.columns where table_schema in (%s) group by TABLE_SCHEMA,TABLE_NAME;"
	detailmetaSql  = "select * from information_schema.columns where table_schema in (%s);"
)

type DBAgent struct {
	DB *sql.DB
}

func NewDBAgent(db *sql.DB) *DBAgent {
	return &DBAgent{
		DB: db,
	}
}

func (db *DBAgent) GetTableMetaData(ctx context.Context, clusterID string, dbName string, tableName string) (metaRes *sqleditor.MetaRes, err error) {
	//execu query
	rows, err := db.DB.Query(fmt.Sprintf(tableMetaQuery, dbName, tableName))
	if err != nil {
		return metaRes, err
	}
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return metaRes, err
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	cols := make([]*sqleditor.Columns, 0)
	for rows.Next() {
		// get RawBytes from data
		err = rows.Scan(scanArgs...)
		if err != nil {
			return metaRes, err
		}

		col := &sqleditor.Columns{
			Col:      string(values[3]),
			DataType: string(values[7]),
			Nullable: strings.ToLower(string(values[6])) == "yes",
		}
		cols = append(cols, col)
	}

	return &sqleditor.MetaRes{
		ClusterId: clusterID,
		Database:  dbName,
		Table:     tableName,
		Columns:   cols,
	}, nil
}

func (db *DBAgent) GetClusterMetaData(ctx context.Context, isBrief bool, showSystemDBFlag bool) (dbmetaList []*sqleditor.DBMeta, err error) {
	dbmetaList = make([]*sqleditor.DBMeta, 0)
	detailmetaDict := make(map[string][]*sqleditor.Columns)

	var execludeDB []string
	if showSystemDBFlag {
		execludeDB = EXCLUDEDB
	} else {
		execludeDB = EXCLUDEDBANDSYS
	}

	//query db
	rows, err := db.DB.Query(dbSql)
	if err != nil {
		return dbmetaList, err
	}
	dbArray := make([]string, 0)
	for rows.Next() {
		var schemaName, schemaType string
		// get RawBytes from data
		err = rows.Scan(&schemaName, &schemaType)
		if err != nil {
			return dbmetaList, err
		}
		if !ArrayIn(schemaName, execludeDB) {
			dbArray = append(dbArray, schemaName)
		}
	}
	dbArrayStr := ""
	for _, db := range dbArray {
		dbArrayStr = dbArrayStr + "'" + db + "',"
	}

	metaSql := ""
	if isBrief {
		metaSql = briefmetaSql
	} else {
		metaSql = detailmetaSql
	}

	//execu query
	metaRows, err := db.DB.Query(fmt.Sprintf(metaSql, dbArrayStr[:len(dbArrayStr)-1]))

	if err != nil {
		return dbmetaList, err
	}

	// Get column names
	columns, err := metaRows.Columns()
	if err != nil {
		return dbmetaList, err
	}

	// Make a slice for the values
	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	tableArrayTemp := make(map[string][]string, 0)
	for metaRows.Next() {
		// get RawBytes from data
		err = metaRows.Scan(scanArgs...)
		if err != nil {
			return dbmetaList, err
		}

		if _, ok := tableArrayTemp[string(values[1])]; !ok {
			tableArrayTemp[string(values[1])] = make([]string, 0)
		}

		isExists := false
		for _, t := range tableArrayTemp[string(values[1])] {
			if string(values[2]) == t {
				isExists = true
				break
			}
		}
		if ! isExists {
			tableArrayTemp[string(values[1])] = append(tableArrayTemp[string(values[1])], string(values[2]))
		}


		if isBrief {
			continue
		}

		col := &sqleditor.Columns{
			Col:      string(values[3]),
			DataType: string(values[7]),
			Nullable: strings.ToLower(string(values[6])) == "yes",
		}

		if _, ok := detailmetaDict[string(values[1])+":"+string(values[2])]; !ok {
			detailmetaDict[string(values[1])+":"+string(values[2])] = make([]*sqleditor.Columns, 0)
		}
		detailmetaDict[string(values[1])+":"+string(values[2])] = append(detailmetaDict[string(values[1])+":"+string(values[2])], col)
	}
	for _, v1 := range dbArray {
		dbMeta := &sqleditor.DBMeta{Name: v1}

		tableArray := make([]string, 0)
		if _, ok := tableArrayTemp[v1]; ok {
			tableArray = append(tableArray, tableArrayTemp[v1]...)
		}

		sort.Sort(Alphabetic(tableArray))

		tablesMeta := make([]*sqleditor.TablesMeta, 0)
		for _, table := range tableArray {
			tMeta := &sqleditor.TablesMeta{Name: table}
			tablesMeta = append(tablesMeta, tMeta)
		}

		dbMeta.Tables = tablesMeta
		dbmetaList = append(dbmetaList, dbMeta)
	}
	if !isBrief {
		for i := 0; i < len(dbmetaList); i++ {
			for k := 0; k < len(dbmetaList[i].Tables); k++ {
				if _, ok := detailmetaDict[dbmetaList[i].Name+":"+dbmetaList[i].Tables[k].Name]; ok {
					dbmetaList[i].Tables[k].Columns = detailmetaDict[dbmetaList[i].Name+":"+dbmetaList[i].Tables[k].Name]
				}
			}
		}
	}
	return dbmetaList, nil
}

func (db *DBAgent) CreateSession(ctx context.Context, clusterID string, expireSec uint64, database string) (sessionID string, err error) {

	db.DB.SetMaxOpenConns(100)
	db.DB.SetMaxIdleConns(20)
	db.DB.SetConnMaxLifetime(time.Second * time.Duration(expireSec))
	conn, err := db.DB.Conn(ctx)
	if err != nil {
		return
	}
	sessionID = CreateSessionCache(conn, expireSec)
	return sessionID, nil
}

func (db *DBAgent) CloseSession(ctx context.Context, sessionID string) error {
	conn, err := GetSessionFromCache(sessionID)
	if err != nil {
		return err
	}
	if conn != nil {
		conn.Close()
	}
	return CloseSessionCache(sessionID)
}

func (db *DBAgent) GetSession(ctx context.Context, sessionID string) (*sql.Conn, error) {
	return GetSessionFromCache(sessionID)
}

func (db *DBAgent) ExecSqlWithSession(ctx context.Context, sessionID, sql string) (stmtRes *sqleditor.StatementsRes, err error) {
	conn, err := GetSessionFromCache(sessionID)
	if err != nil {
		return nil, err
	}
	return queryDB(ctx, conn, sql)
}

type Alphabetic []string

func (list Alphabetic) Len() int { return len(list) }

func (list Alphabetic) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

func (list Alphabetic) Less(i, j int) bool {
	var si = list[i]
	var sj = list[j]
	var siLower = strings.ToLower(si)
	var sjLower = strings.ToLower(sj)
	if siLower == sjLower {
		return si < sj
	}
	return siLower < sjLower
}

func ArrayIn(target string, strArray []string) bool {
	sort.Strings(strArray)
	index := sort.SearchStrings(strArray, target)
	if index < len(strArray) && strArray[index] == target {
		return true
	}
	return false
}
