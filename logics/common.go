package logics

import (
	"crypto/rsa"
	"database/sql"
)

var (
	privateKey *rsa.PrivateKey
	dbPool     *sql.DB
)

func SetPrivateKey(key *rsa.PrivateKey) {
	privateKey = key
}

func SetDBPool(i *sql.DB) {
	dbPool = i
}
