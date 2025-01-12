package logics

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
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
	o      interfaces.LogicsOutbox
	m      interfaces.DrivenMailer
	sender string

	cache        sync.Map
	cacheMapping map[interfaces.MailOp]string
}

func NewLogicsMailer() interfaces.LogicsMailer {
	mailerOnce.Do(func() {
		cacheMapping := map[interfaces.MailOp]string{
			interfaces.UserWelcome: "user_welcome.tmpl",
		}
		m = &mailer{
			o:      NewOutbox(),
			m:      drivenadapters.NewMailer(),
			sender: "user_test@example.com",

			cacheMapping: cacheMapping,
		}
	})

	m.o.RegisterHandler(interfaces.UserCreatedEMAIL, m.Handle)

	return m
}

func (m *mailer) Handle(ctx context.Context, msg *interfaces.OutboxMessage) error {
	var data interface{}
	err := json.Unmarshal(msg.Msg, &data)
	if err != nil {
		return fmt.Errorf("mailer: failed to unmarshal message: %w, msgID: %s", err, msg.ID)
	}

	switch msg.Op {
	case interfaces.UserCreatedEMAIL:
		mailMsg := &interfaces.MailMessage{
			ID: data.(map[string]interface{})["ID"].(string),
			To: data.(map[string]interface{})["To"].(string),
		}
		return m.SendMail(ctx, interfaces.UserWelcome, mailMsg)
	default:
		return fmt.Errorf("mailer: invalid operation: %d, msgID: %s", msg.Op, msg.ID)
	}
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
