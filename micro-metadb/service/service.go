package service

import (
	"github.com/pingcap-inc/tiem/library/framework"
	"github.com/pingcap-inc/tiem/micro-metadb/models"
)

type DBServiceHandler struct{
	dao *models.DAOManager
	log *framework.LogRecord
}

func (handler *DBServiceHandler) Dao() *models.DAOManager {
	return handler.dao
}

func (handler *DBServiceHandler) SetDao(dao *models.DAOManager) {
	handler.dao = dao
}

func (handler *DBServiceHandler) Log() *framework.LogRecord {
	return handler.log
}

func (handler *DBServiceHandler) SetLog(lr *framework.LogRecord) {
	handler.log = lr
}