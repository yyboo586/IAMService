package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func NewDB(user, passwd, host string, port int, dbName string) (dbPool *sql.DB, err error) {
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", user, passwd, host, port, dbName)

	if dbPool, err = sql.Open("mysql", dsn); err != nil {
		log.Printf("Failed to open database, error: %v\n", err)
		return nil, err
	}

	if err = dbPool.Ping(); err != nil {
		log.Printf("DB ping failed, error: %v\n", err)
		return nil, err
	}

	return dbPool, nil
}
