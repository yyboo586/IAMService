package drivenadapters

import (
	gomail "github.com/go-mail/mail/v2"
	"github.com/yyboo586/common/logUtils"
)

var (
	mailDialer     *gomail.Dialer
	loggerInstance *logUtils.Logger
)

func SetMailDialer(i *gomail.Dialer) {
	mailDialer = i
}

func SetLogger(i *logUtils.Logger) {
	loggerInstance = i
}
