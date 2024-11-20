package logics

import (
	"crypto/rsa"

	"github.com/yyboo586/common/logUtils"
)

var (
	privateKey     *rsa.PrivateKey
	loggerInstance *logUtils.Logger
)

func SetPrivateKey(key *rsa.PrivateKey) {
	privateKey = key
}

func SetLogger(i *logUtils.Logger) {
	loggerInstance = i
}
