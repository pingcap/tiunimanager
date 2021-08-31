package service

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
)

type DBServiceHandler struct {
	dao *models.DAOManager
}

func NewDBServiceHandler(dataDir string, fw *framework.BaseFramework) *DBServiceHandler {
	handler := new(DBServiceHandler)
	dao := new(models.DAOManager)
	dao.InitDB(dataDir)
	dao.InitTables()
	dao.InitData()
	handler.SetDao(dao)
	return handler
}

func (handler *DBServiceHandler) Dao() *models.DAOManager {
	return handler.dao
}

func (handler *DBServiceHandler) SetDao(dao *models.DAOManager) {
	handler.dao = dao
}
