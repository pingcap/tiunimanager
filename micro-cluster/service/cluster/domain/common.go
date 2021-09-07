package domain

import (
	"github.com/pingcap-inc/tiem/library/framework"
	log "github.com/sirupsen/logrus"
)

func getLogger() *log.Entry {
	return framework.Log()
}
