package drivenadapters

import (
	"context"
	"sync"

	gomail "github.com/go-mail/mail/v2"
	"github.com/yyboo586/common/logUtils"
)

var (
	mOnce sync.Once
	m     *mailer
)

type mailer struct {
	dialer *gomail.Dialer
	logger *logUtils.Logger
}

func NewMailer() *mailer {
	mOnce.Do(func() {
		m = &mailer{
			dialer: mailDialer,
			logger: loggerInstance,
		}
	})

	return m
}

func (m *mailer) SendMail(ctx context.Context, to, from, title, plainBody, htmlBody string) (err error) {
	msg := gomail.NewMessage()

	msg.SetHeader("To", to)
	msg.SetHeader("From", from)
	msg.SetHeader("Subject", title)
	msg.SetBody("text/plain", plainBody)
	msg.AddAlternative("text/html", htmlBody)

	return m.dialer.DialAndSend(msg)
}
