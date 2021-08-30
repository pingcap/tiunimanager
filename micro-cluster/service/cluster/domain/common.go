package domain

import (
	"github.com/pingcap-inc/tiem/library/framework"
)

func getLogger() *framework.LogRecord {
	return framework.GetLogger()
}

