package service

import (
	"github.com/pingcap/tiem/library/thirdparty/logger"
)

var TiEMMetaDBServiceName = "go.micro.tiem.db"

type DBServiceHandler struct{}

var log *logger.LogRecord

func InitLogger() {
	log = logger.GetLogger()
}
