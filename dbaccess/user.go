package dbaccess

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/logUtils"
)

var (
	uOnce sync.Once
	u     *user
)

type user struct {
	db     *sql.DB
	logger *logUtils.Logger
}

func NewUser() interfaces.DBUser {
	uOnce.Do(func() {
		u = &user{
			db:     dbPool,
			logger: loggerInstance,
		}
	})

	return u
}

func (u *user) Create(user *interfaces.User) (err error) {
	sqlStr := "insert into t_user (id, name, password) values(?, ?, ?)"

	if _, err := u.db.Exec(sqlStr, user.ID, user.Name, user.Password); err != nil {
		return fmt.Errorf("dbaccess: create user error: %w", err)
	}

	return nil
}

func (u *user) GetUserInfoByID(id string) (user *interfaces.User, exists bool, err error) {
	user = &interfaces.User{}
	sqlStr := "select id, name from t_user where id = ?"

	if err = u.db.QueryRow(sqlStr, id).Scan(&user.ID, &user.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("dbaccess: GetUserInfoByID error: %w", err)
	}

	return user, true, nil
}

func (u *user) GetUserInfoByName(name string) (user *interfaces.User, exists bool, err error) {
	user = &interfaces.User{}
	sqlStr := "select id, name, password from t_user where name = ?"

	if err = u.db.QueryRow(sqlStr, name).Scan(&user.ID, &user.Name, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("dbaccess: GetUserInfoByName error: %w", err)
	}

	return user, true, nil
}

func (u *user) UpdateLoginTime(id string) (err error) {
	sqlStr := "update t_user set last_login_at = ? where id = ?"

	if _, err = u.db.Exec(sqlStr, time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05"), id); err != nil {
		return fmt.Errorf("dbaccess: UpdateLoginTime error: %w", err)
	}

	return nil
}
