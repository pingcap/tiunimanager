package service

import (
	"github.com/pingcap-inc/tiem/library/framework"
)

type DBServiceHandler struct{}

func getLogger() *framework.LogRecord {
	return framework.GetLogger()
}