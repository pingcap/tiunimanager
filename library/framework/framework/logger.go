package framework

import (
	"context"
	logger2 "github.com/pingcap-inc/tiem/library/framework/logger"

	"github.com/pingcap-inc/tiem/library/firstparty/util"
	log "github.com/sirupsen/logrus"
)

type Logger interface {
	GetDefaultLogger() *logger2.LogRecord
	SetDefaultLogger(*logger2.LogRecord)
	// Return the logger which has already stored inside the ctx.
	// If there is no such logger, return the default logger instead.
	WithContext(ctx context.Context) *logger2.LogRecord
	// Return a new context with extra fields added to the previous logger.
	// If there is no such previous logger, use the default logger instead.
	NewContextWithField(ctx context.Context, key string, value interface{}) context.Context
	NewContextWithFields(ctx context.Context, fields log.Fields) context.Context
}

func initLogger() (Logger, error) {
	util.AssertWithInfo(false, "Not Implemented Yet")
	return nil, nil
}
