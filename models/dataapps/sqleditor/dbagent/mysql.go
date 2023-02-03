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
	"fmt"
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	_ "github.com/pingcap/tidb/parser/test_driver"
	"github.com/pingcap/tiunimanager/message/dataapps/sqleditor"
	"reflect"
	"time"
	"unsafe"
)

const (
	SELECT  = "select"
	INSERT  = "insert"
	UPDATE  = "update"
	DELETE  = "delete"
	DROP    = "drop"
	USE     = "use"
	SHOW    = "show"
	EXPLAIN = "explain"
	OTHER   = "other"
	LIMIT   = 2000
)

func queryDB(ctx context.Context, conn *sql.Conn, sql string) (stmtRes *sqleditor.StatementsRes, err error) {
	dataRes := &sqleditor.DataRes{}

	stmtNode, err := parser.New().ParseOneStmt(sql, "", "")
	if err != nil {
		return nil, err
	}
	sqlType := GetStmtType(ctx, stmtNode)
	//databases := GetDatabases(ctx, stmtNode)
	startTime := time.Now()
	startMs := startTime.UnixNano() / int64(time.Millisecond)
	rows, affect, err := executeSQL(ctx, conn, sql, sqlType)
	endTime := time.Now()
	endMs := endTime.UnixNano() / int64(time.Millisecond)
	if err != nil {
		return nil, err
	}

	dataRes.StartTime = startTime.Format("2006-01-02 15:04:05")
	dataRes.EndTime = endTime.Format("2006-01-02 15:04:05")
	dataRes.ExecuteTime = fmt.Sprint(endMs - startMs)
	dataRes.Limit = LIMIT

	if sqlType == SELECT || sqlType == SHOW || sqlType == EXPLAIN {
		dataRes.Query = "Query OK!"
	} else if sqlType == USE {
		dataRes.Query = "Database changed!"
	} else {
		dataRes.Query = fmt.Sprintf("Query OK, %v rows affected (%.3f sec)", affect, float64(endMs-startMs)/1000.0)
	}

	scanRows(rows, uint64(LIMIT), dataRes)
	if rows != nil {
		dataRes.RowCount = int64(len(dataRes.Rows))
		rows.Close()
	}
	return &sqleditor.StatementsRes{
		Data: dataRes,
	}, nil
}

func executeSQL(ctx context.Context, conn *sql.Conn, sql string, sqlType string) (*sql.Rows, int64, error) {
	affect := int64(0)
	if sqlType == SELECT || sqlType == SHOW || sqlType == EXPLAIN {
		rows, err := conn.QueryContext(ctx, sql)
		if err != nil {
			return nil, 0, err
		}
		return rows, 0, nil
	} else {
		exec, err := conn.ExecContext(ctx, sql)
		if err != nil {
			return nil, 0, err
		}
		affect, err = exec.RowsAffected()
		return nil, affect, nil
	}
}

func GetStmtType(ctx context.Context, stmtNode ast.StmtNode) string {
	if stmtNode != nil {
		switch stmtNode.(type) { //nolint:gosimple // ignore
		case *ast.InsertStmt:
			return INSERT
		case *ast.SelectStmt, *ast.SetOprStmt:
			return SELECT
		case *ast.UpdateStmt:
			return UPDATE
		case *ast.DeleteStmt:
			return DELETE
		case *ast.ShowStmt:
			return SHOW
		case *ast.ExplainStmt:
			return EXPLAIN
		case *ast.UseStmt:
			return USE
		default:
			return OTHER
		}
	}
	return OTHER
}

func GetDatabases(ctx context.Context, stmtNode ast.StmtNode) []string {
	if stmtNode == nil {
		return nil
	}
	databases := getSchema(stmtNode)
	exists := make(map[string]struct{})
	result := make([]string, 0, len(databases))
	for _, database := range databases {
		if _, ok := exists[database]; ok {
			continue
		}
		exists[database] = struct{}{}
		result = append(result, database)
	}
	return result
}

func getSchema(node ast.Node) []string {
	if node == nil {
		return nil
	}
	var ans []string
	switch astNode := node.(type) {

	case *ast.SelectStmt:
		if astNode.From == nil {
			return nil
		}
		ans = append(ans, getSchema(astNode.From.TableRefs)...)
	case *ast.SetOprStmt:
		if astNode.SelectList == nil || len(astNode.SelectList.Selects) == 0 {
			return nil
		}
		for _, selectNode := range astNode.SelectList.Selects {
			ans = append(ans, getSchema(selectNode)...)
		}
	case *ast.SetOprSelectList:
		if len(astNode.Selects) == 0 {
			return nil
		}
		for _, selectNode := range astNode.Selects {
			ans = append(ans, getSchema(selectNode)...)
		}

	case *ast.Join:
		ans = append(ans, getSchema(astNode.Left)...)
		ans = append(ans, getSchema(astNode.Right)...)
	case *ast.TableSource:
		ans = append(ans, getSchema(astNode.Source)...)
	case *ast.TableName:
		ans = append(ans, astNode.Schema.O)
	}
	return ans
}

func scanRows(rows *sql.Rows, limits uint64, result *sqleditor.DataRes) {
	rowNum := int64(0)
	if rows == nil {
		return
	}
	//res := make([]map[string]string, 0)
	array := make([][]string, 0)
	colTypes, _ := rows.ColumnTypes()
	var rowParam = make([]interface{}, len(colTypes))
	var rowValue = make([]interface{}, len(colTypes))

	for i, colType := range colTypes {
		rowValue[i] = reflect.New(colType.ScanType())
		rowParam[i] = reflect.ValueOf(&rowValue[i]).Interface()

	}
	var columns []*sqleditor.Columns

	//columns
	for _, col := range colTypes {
		nullable, _ := col.Nullable()
		param := &sqleditor.Columns{
			Col:      col.Name(),
			DataType: col.DatabaseTypeName(),
			Nullable: nullable,
		}
		columns = append(columns, param)
	}

	// rows
	for rows.Next() {
		rowNum++
		if limits > 0 && rowNum > int64(limits) {
			continue
		}
		rows.Scan(rowParam...)
		//record := make(map[string]string)
		row := make([]string, 0)
		for i, _ := range colTypes {

			if rowValue[i] == nil {
				//record[colType.Name()] = ""
				row = append(row, "")
			} else {
				switch f := rowValue[i].(type) {
				case time.Time:
					row = append(row, f.Format("2006-01-02 15:04:05"))
				default:
					row = append(row, Byte2Str(f.([]byte)))
				}

			}
		}
		array = append(array, row)

	}
	result.Columns = columns
	result.Rows = array
	return
}

// Byte2Str []byte to string
func Byte2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
