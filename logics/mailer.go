package logics

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"html/template"
	"sync"

	"github.com/yyboo586/IAMService/drivenadapters"
	"github.com/yyboo586/IAMService/interfaces"
)

var (
	mailerOnce sync.Once
	m          *mailer

	//go:embed templates
	templateFS embed.FS
)

type mailer struct {
	m      interfaces.DrivenMailer
	sender string

	cache        sync.Map
	cacheMapping map[interfaces.MailOp]string
}

func NewLogicsMailer() *mailer {
	mailerOnce.Do(func() {
		cacheMapping := map[interfaces.MailOp]string{
			interfaces.UserWelcome: "user_welcome.tmpl",
		}
		m = &mailer{
			m:      drivenadapters.NewMailer(),
			sender: "user_test@example.com",

			cacheMapping: cacheMapping,
		}
	})
	return m
}

func (m *mailer) SendMail(ctx context.Context, op interfaces.MailOp, msg *interfaces.MailMessage) (err error) {
	if msg == nil {
		return errors.New("invalid mail message")
	}

	tmpl, err := m.getTemplate(op)
	if err != nil {
		return err
	}

	title := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(title, "subject", msg); err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(plainBody, "plainBody", msg); err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", msg); err != nil {
		return err
	}

	return m.m.SendMail(ctx, msg.To, m.sender, title.String(), plainBody.String(), htmlBody.String())
}

func (m *mailer) getTemplate(op interfaces.MailOp) (*template.Template, error) {
	if value, ok := m.cache.Load(op); ok {
		return value.(*template.Template), nil
	}

	templatePath, ok := m.cacheMapping[op]
	if !ok {
		return nil, errors.New("invalid mail op")
	}

	tmpl, err := template.ParseFS(templateFS, "templates/"+templatePath)
	if err != nil {
		return nil, err
	}
	m.cache.Store(op, tmpl)

	return tmpl, nil
}
