package domain

import (
	"context"
	"github.com/pingcap-inc/tiem/library/framework"
	log "github.com/sirupsen/logrus"
)

func getLogger() *log.Entry {
	return framework.Log()
}

func getLoggerWithContext(ctx context.Context) *log.Entry {
	return framework.LogWithContext(ctx)
}
