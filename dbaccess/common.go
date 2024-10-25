package dbaccess

import "database/sql"

var db *sql.DB

func SetDBPool(i *sql.DB) {
	db = i
}
