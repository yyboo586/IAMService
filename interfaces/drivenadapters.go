package interfaces

import "context"

type DrivenMailer interface {
	SendMail(ctx context.Context, to, from, title, plainBody, htmlBody string) error
}
