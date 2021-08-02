package service

import "github.com/pingcap/ticp/addon/logger"

var TiCPMetaDBServiceName = "go.micro.ticp.db"

type DBServiceHandler struct{}

var log *logger.LogRecord

func InitLogger() {
	log = logger.GetLogger()
}
