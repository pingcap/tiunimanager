package service

import (
	"github.com/pingcap-inc/tiem/library/thirdparty/logger"
)

var TiEMMetaDBServiceName = "go.micro.tiem.db"

type DBServiceHandler struct{}

var log *logger.LogRecord

func InitLogger(defaultLogger *logger.LogRecord) {
	log = defaultLogger
}
