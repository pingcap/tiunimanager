package service

import (
	"github.com/pingcap/ticp/library/thirdparty/logger"
)

var TiCPMetaDBServiceName = "go.micro.ticp.db"

type DBServiceHandler struct{}

var log *logger.LogRecord

func InitLogger() {
	log = logger.GetLogger()
}
