package interfaces

import (
	"context"
	"database/sql"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

//go:generate mockgen -source=./logics.go -destination=mock/logics_mock.go -package=mock

type User struct {
	ID       string
	Name     string
	Password string
	Email    string
}

type LogicsUser interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error
	// Login validate password and return user id, jwtToken
	Login(name, password string) (string, string, error)
	// GetUserInfo return user info by id
	GetUserInfo(ctx context.Context, id string) (*User, error)
}

type MailOp string

const (
	UserWelcome MailOp = "user_welcome"
)

type MailMessage struct {
	ID string `json:"id"`
	To string
}

type LogicsMailer interface {
	SendMail(ctx context.Context, op MailOp, msg *MailMessage) error
}

type CustomClaims struct {
	jwt.Claims

	ExtClaims map[string]interface{}
}

type LogicsJWT interface {
	Sign(userID string, claims map[string]interface{}, setID, alg string) (string, error)
	Verify(jwtTokenStr string) (*CustomClaims, error)
	GetPublicKey(kid string) (key *jose.JSONWebKey, err error)

	RevokeToken(jwtTokenStr string) error
}

type OutboxBussinessType int

const (
	_                OutboxBussinessType = iota
	UserCreatedEMAIL OutboxBussinessType = iota
	UserCreatedMQ
)

type OutboxMessageStatus int

const (
	_                            OutboxMessageStatus = iota
	OutboxMessageStatusUnhandled                     // 未处理
	OutboxMessageStatusHandled                       // 已处理
	OutboxMessageStatusError                         // 处理失败
)

type OutboxMessage struct {
	ID     string
	Op     OutboxBussinessType
	Msg    []byte
	Status OutboxMessageStatus
}

type OutboxHandler func(ctx context.Context, msg *OutboxMessage) error

type LogicsOutbox interface {
	// 添加消息
	AddMessage(ctx context.Context, tx *sql.Tx, op OutboxBussinessType, data []byte) error
	// 注册消息处理函数
	RegisterHandler(op OutboxBussinessType, handler OutboxHandler)
	// 唤醒推送 goroutine
	Notify()
}
