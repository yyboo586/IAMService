package logics

import (
	"database/sql"
	"fmt"

	"github.com/yyboo586/common/logUtils"
)

var (
	loggerInstance *logUtils.Logger
	dbPoolInstance *sql.DB
)

func SetLogger(i *logUtils.Logger) {
	loggerInstance = i
}

func SetDB(i *sql.DB) {
	dbPoolInstance = i
}

// 无论回滚失败，还是提交失败，数据库引擎都会自动回滚事务中的所有更改。
func withTransaction(db *sql.DB, fn func(tx *sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("transaction start error: %v", err.Error())
	}

	defer func() {
		if err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				loggerInstance.Errorf("transaction rollback error: %v\n", rErr)
			}
			return
		}

		if tErr := tx.Commit(); tErr != nil {
			loggerInstance.Errorf("transaction commit error: %v\n", tErr.Error())
			err = tErr
		}
	}()

	return fn(tx)
}
