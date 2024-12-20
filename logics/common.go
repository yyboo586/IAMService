package logics

import (
	"github.com/yyboo586/common/logUtils"
)

var (
	loggerInstance *logUtils.Logger
)

func SetLogger(i *logUtils.Logger) {
	loggerInstance = i
}
