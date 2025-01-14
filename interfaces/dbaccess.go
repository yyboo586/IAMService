package interfaces

import (
	"context"
	"database/sql"

	"github.com/go-jose/go-jose/v4"
)

//go:generate mockgen -source=./dbaccess.go -destination=mock/dbaccess_mock.go -package=mock

type DBUser interface {
	// Create creates a new user.
	Create(tx *sql.Tx, user *User) error
	// GetUserInfoByID fetches a user by id.
	GetUserInfoByID(id string) (*User, bool, error)
	// GetUserInfoByName fetches a user by name.
	GetUserInfoByName(name string) (*User, bool, error)
	// UpdateLoginTime updates the last login time of a user.
	UpdateLoginTime(id string) error
}

type KeyStatus int

const (
	KeyStatusValid   KeyStatus = 1
	KeyStatusExpired KeyStatus = 2
)

type DBJWT interface {
	AddKeySet(setID string, keySet *jose.JSONWebKeySet) error
	GetKeySet(setID string) (kSet *jose.JSONWebKeySet, err error)
	GetKey(kid string) (key *jose.JSONWebKey, err error)
	SetKeyStatus(kid string, status KeyStatus) error

	AddBlacklist(id string) error
	GetBlacklist(id string) (exists bool, err error)
}

type DBOutbox interface {
	// 添加消息
	Add(ctx context.Context, tx *sql.Tx, msg *OutboxMessage) error
	// 获取消息
	Get(ctx context.Context, status OutboxMessageStatus) (msg *OutboxMessage, exists bool, err error)
	// 更新消息状态
	Update(ctx context.Context, tx *sql.Tx, id string, status OutboxMessageStatus) error
	// 批量删除已正常消费掉的消息
	Delete(ctx context.Context, status OutboxMessageStatus, batchSize int) (rowsAffected int64, err error)
}
