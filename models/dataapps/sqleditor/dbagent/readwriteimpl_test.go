package dbagent

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDbagent_GetTableMetaData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := &gin.Context{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	dbagent := DBAgent{
		DB: db,
	}
	dbName := "test"
	tableName := "test"
	clusterID := "test"
	columns := []string{"col1", "col2", "col3", "col4", "col5", "col6", "col7", "col8"}
	rows := sqlmock.NewRows(columns).AddRow("value1", "value2", "value3", "value4", "value5", "value6", "value7", "value8")
	rows = rows.AddRow("value12", "value22", "value32", "value42", "value52", "value62", "value72", "value82")
	mock.ExpectQuery(fmt.Sprintf(tableMetaQuery, dbName, tableName)).WillReturnRows(rows)
	res, err := dbagent.GetTableMetaData(c, clusterID, dbName, tableName)
	assert.Equal(t, len(res.Columns), 2)
	assert.Equal(t, res.Columns[0].Col, "value4")
	assert.NoError(t, err)
}

func TestDbagent_GetClusterMetaData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := &gin.Context{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	dbagent := DBAgent{
		DB: db,
	}

	t.Run("brief", func(t *testing.T) {
		columns := []string{"schema_name", "t1"}
		rows := sqlmock.NewRows(columns).AddRow("INFORMATION_SCHEMA", "1")
		rows = rows.AddRow("test", "2")
		rows = rows.AddRow("test2", "2")
		mock.ExpectQuery(dbSql).WillReturnRows(rows)
		columns2 := []string{"col1", "col2", "col3", "col4", "col5", "col6", "col7", "col8"}
		rows2 := sqlmock.NewRows(columns2).AddRow("value1", "value2", "value3", "value4", "value5", "value6", "value7", "value8")
		rows2 = rows2.AddRow("value12", "value22", "value32", "value42", "value52", "value62", "value72", "value82")
		mock.ExpectQuery(fmt.Sprintf(briefmetaSql, "'test','test2'")).WillReturnRows(rows2)
		res, err := dbagent.GetClusterMetaData(c, true, false)
		assert.Equal(t, res[0].Name, "test")
		assert.NoError(t, err)
	})

	t.Run("detail", func(t *testing.T) {
		columns := []string{"schema_name", "t1"}
		rows := sqlmock.NewRows(columns).AddRow("INFORMATION_SCHEMA", "1")
		rows = rows.AddRow("test", "2")
		rows = rows.AddRow("test2", "2")
		mock.ExpectQuery(dbSql).WillReturnRows(rows)
		columns2 := []string{"col1", "col2", "col3", "col4", "col5", "col6", "col7", "col8"}
		rows2 := sqlmock.NewRows(columns2).AddRow("value1", "value2", "value3", "value4", "value5", "value6", "value7", "value8")
		rows2 = rows2.AddRow("value12", "value22", "value32", "value42", "value52", "value62", "value72", "value82")
		mock.ExpectQuery(fmt.Sprintf(detailmetaSql, "'INFORMATION_SCHEMA','test','test2'")).WillReturnRows(rows2)
		res, err := dbagent.GetClusterMetaData(c, false, true)
		assert.Equal(t, res[0].Name, "INFORMATION_SCHEMA")
		assert.NoError(t, err)
	})
}

func TestDbagent_Session(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	c := &gin.Context{}

	db, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	dbagent := DBAgent{
		DB: db,
	}

	clusterID := "test"
	var sessionID string

	t.Run("createSession", func(t *testing.T) {
		sessionID, err = dbagent.CreateSession(c, clusterID, 60, "test")
		assert.NoError(t, err)
		assert.NotEmpty(t, sessionID)
	})

	t.Run("getSession", func(t *testing.T) {
		conn, err := dbagent.GetSession(c, sessionID)
		assert.NoError(t, err)
		assert.NotNil(t, conn)
	})

	t.Run("closeSession", func(t *testing.T) {
		err = dbagent.CloseSession(c, sessionID)
		assert.NoError(t, err)
	})

}
