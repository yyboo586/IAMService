package dbaccess

import (
	"database/sql"

	"github.com/yyboo586/common/logUtils"
)

var (
	dbPool         *sql.DB
	loggerInstance *logUtils.Logger
)

func SetDBPool(i *sql.DB) {
	dbPool = i
}

func SetLogger(i *logUtils.Logger) {
	loggerInstance = i
}

// func withTransaction(db *sql.DB, fn func(tx *sql.Tx) error) error {
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("transaction start error: %v", err.Error())
// 	}

// 	defer func() {
// 		if err != nil {
// 			if rErr := tx.Rollback(); rErr != nil {
// 				u.logger.Errorf("transaction rollback error: %v\n", rErr)
// 			}
// 			return
// 		}

// 		if tErr := tx.Commit(); tErr != nil {
// 			u.logger.Errorf("transaction commit error: %v\n", tErr.Error())
// 			err = tErr
// 		}
// 	}()

// 	return fn(tx)
// }
