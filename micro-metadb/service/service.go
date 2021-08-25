package service

import (
	config2 "github.com/pingcap-inc/tiem/library/framework/config"
	logger2 "github.com/pingcap-inc/tiem/library/framework/logger"
)

var TiEMMetaDBServiceName = "go.micro.tiem.db"

type DBServiceHandler struct{}

var log *logger2.LogRecord

func InitLogger(key config2.Key) {
	log = logger2.GetLogger(key)
}
