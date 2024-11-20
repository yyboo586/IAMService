package interfaces

import "github.com/go-jose/go-jose/v4"

//go:generate mockgen -source=./dbaccess.go -destination=mock/dbaccess_mock.go -package=mock

type DBUser interface {
	// Create creates a new user.
	Create(user *User) error
	// GetUserInfoByID fetches a user by id.
	GetUserInfoByID(id string) (*User, bool, error)
	// FetchByName fetches a user by name.
	FetchByName(name string) (*User, bool, error)
	// UpdateLoginTime updates the last login time of a user.
	UpdateLoginTime(id string) error
}

type DBJWT interface {
	AddKeySet(setID string, keySet *jose.JSONWebKeySet) error
	GetKeySet(setID string) (kSet *jose.JSONWebKeySet, err error)
	GetKey(kid string) (key *jose.JSONWebKey, err error)
}
