package drivenadapters

import gomail "github.com/go-mail/mail/v2"

var (
	mailDialer *gomail.Dialer
)

func SetMailDialer(i *gomail.Dialer) {
	mailDialer = i
}
