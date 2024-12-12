package interfaces

import (
	"context"
	"sync"

	"github.com/go-jose/go-jose/v4"
)

//go:generate mockgen -source=./logics.go -destination=mock/logics_mock.go -package=mock

type User struct {
	ID       string
	Name     string
	Password string
	Email    string
}

var free = sync.Pool{
	New: func() any {
		return &User{}
	},
}

func NewUser() *User {
	return free.Get().(*User)
}

func FreeUser(user *User) {
	free.Put(user)
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

type LogicsJWT interface {
	Sign(userID string, claims map[string]interface{}, setID, alg string) (string, error)
	Verify(jwtTokenStr string) (map[string]interface{}, error)
	GetPublicKey(kid string) (key *jose.JSONWebKey, err error)
}
