package interfaces

import "context"

type DrivenMailer interface {
	SendMail(ctx context.Context, to, from, title, plainBody, htmlBody string) error
}

type DrivenMQ interface {
	Publish(ctx context.Context, topic string, msg []byte) error
}
