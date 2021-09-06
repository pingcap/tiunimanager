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
	handler.SetDao(dao)

	dao.InitDB(dataDir)
	if dao.Db().Migrator().HasTable(&models.Tenant{}) {
		framework.LogWithCaller().Warn("data existed, skip initialization")
		return handler
	} else {
		dao.InitTables()
		dao.InitData()
	}
	return handler
}

func (handler *DBServiceHandler) Dao() *models.DAOManager {
	return handler.dao
}

func (handler *DBServiceHandler) SetDao(dao *models.DAOManager) {
	handler.dao = dao
}
