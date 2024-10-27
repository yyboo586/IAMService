package interfaces

import "context"

//go:generate mockgen -source=./logics.go -destination=mock/logics_mock.go -package=mock

type contextKey string

const (
	TokenKey contextKey = "token"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LogicsUser interface {
	// Create creates a new user
	Create(user *User) error
	// Login validate password and return user id, jwtToken
	Login(name, password string) (string, string, error)
	// GetUserInfo return user info by id
	GetUserInfo(ctx context.Context, id string) (*User, error)
}
