/******************************************************************************
 * Copyright (c)  2022 PingCAP                                                *
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

package sql

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/pingcap/tiunimanager/common/errors"
	"github.com/pingcap/tiunimanager/message/dataapps/sqleditor"
	"github.com/pingcap/tiunimanager/micro-cluster/cluster/management/meta"
)

var (
	EXCLUDEDB       = []string{"__TiDB_BR_Temporary_mysql", "tidbcloud_dm_meta", "tidb_cdc", "lightning_task_info"}
	EXCLUDEDBANDSYS = []string{"__TiDB_BR_Temporary_mysql", "tidbcloud_dm_meta", "tidb_cdc", "lightning_task_info", "INFORMATION_SCHEMA", "mysql", "PERFORMANCE_SCHEMA", "METRICS_SCHEMA"}
)

var (
	tableMetaQuery = "select * from information_schema.columns where table_schema =\"%s\"  and table_name = \"%s\";"
	dbSql          = "select \n  schema_name,\n  case schema_name\n  when 'INFORMATION_SCHEMA' then 1\n  when 'mysql' then 1\n  when 'PERFORMANCE_SCHEMA' then 1\n  else 2\n  end as t1\n  from information_schema.schemata group by schema_name order by t1,lower(schema_name);"
	briefmetaSql   = "select '' as t1,TABLE_SCHEMA,TABLE_NAME from information_schema.columns where table_schema in (%s) group by TABLE_SCHEMA,TABLE_NAME;"
	detailmetaSql  = "select * from information_schema.columns where table_schema in (%s);"
)

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

func GetTableMetaData(ctx context.Context, clusterID string, dbName string, tableName string) (metaRes *sqleditor.MetaRes, err error) {
	// connect database
	clusterMeta, err := meta.Get(ctx, clusterID)
	if err != nil {
		return metaRes, err
	}

	db, err := meta.CreateSQLLink(ctx, clusterMeta)

	if err != nil {
		return metaRes, errors.WrapError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, err.Error(), err)
	}
	defer db.Close()

	//execu query
	rows, err := db.Query(fmt.Sprintf(tableMetaQuery, dbName, tableName))
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

func GetClusterMetaData(ctx context.Context, clusterID string, isBrief bool, showSystemDBFlag bool) (dbmetaList []*sqleditor.DBMeta, err error) {
	dbmetaList = make([]*sqleditor.DBMeta, 0)
	detailmetaDict := make(map[string][]*sqleditor.Columns)

	var execludeDB []string
	if showSystemDBFlag {
		execludeDB = EXCLUDEDB
	} else {
		execludeDB = EXCLUDEDBANDSYS
	}

	// connect database
	clusterMeta, err := meta.Get(ctx, clusterID)
	if err != nil {
		return dbmetaList, err
	}
	db, err := meta.CreateSQLLink(ctx, clusterMeta)
	if err != nil {
		return dbmetaList, errors.WrapError(errors.TIUNIMANAGER_CONNECT_TIDB_ERROR, err.Error(), err)
	}
	defer db.Close()

	//query db
	rows, err := db.Query(fmt.Sprintf(dbSql))
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
	metaRows, err := db.Query(fmt.Sprintf(metaSql, dbArrayStr[:len(dbArrayStr)-1]))
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

	for _, v1 := range dbArray {
		tableArrayTemp := make(map[string]bool, 0)
		dbMeta := &sqleditor.DBMeta{Name: v1}

		for metaRows.Next() {
			// get RawBytes from data
			err = metaRows.Scan(scanArgs...)
			if err != nil {
				return dbmetaList, err
			}
			if v1 == string(values[1]) {
				tableArrayTemp[string(values[2])] = true
			}

			if isBrief {
				continue
			}

			col := &sqleditor.Columns{
				Col:      string(values[3]),
				DataType: string(values[7]),
				Nullable: strings.ToLower(string(values[6])) == "yes",
			}

			if _, ok := detailmetaDict[v1+":"+string(values[2])]; !ok {
				detailmetaDict[v1+":"+string(values[2])] = make([]*sqleditor.Columns, 0)
			}
			detailmetaDict[v1+":"+string(values[2])] = append(detailmetaDict[v1+":"+string(values[2])], col)
		}

		tableArray := make([]string, 0)
		for k, _ := range tableArrayTemp {
			tableArray = append(tableArray, k)
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
