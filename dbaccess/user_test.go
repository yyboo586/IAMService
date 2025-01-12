package dbaccess

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/logUtils"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var (
	errDatabase = errors.New("database error")
)

func setupUserTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, interfaces.DBUser) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	logger, _ := logUtils.NewLogger("debug")

	return db, mock, &user{db: db, logger: logger}
}

func TestCreate(t *testing.T) {
	convey.Convey("Test DBUser Create()", t, func() {
		db, mock, dbUser := setupUserTest(t)
		defer db.Close()

		u := &interfaces.User{
			ID:       "id",
			Name:     "test",
			Password: "123456",
		}
		convey.Convey("数据库错误", func() {
			// 设置期望
			mock.ExpectBegin()
			mock.ExpectExec("insert into t_user").
				WithArgs("id", "test", "123456").
				WillReturnError(errDatabase)
			mock.ExpectRollback()

			// 开启事务并执行
			tx, _ := db.Begin()
			err := dbUser.Create(tx, u)
			tx.Rollback()

			// 断言结果
			assert.Equal(t, fmt.Errorf("dbaccess: create user error: %w", errDatabase), err)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
		convey.Convey("数据插入成功", func() {
			// 设置期望
			mock.ExpectBegin()
			mock.ExpectExec("insert into t_user").
				WithArgs(u.ID, u.Name, u.Password).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			// 开启事务并执行
			tx, _ := db.Begin()
			err := dbUser.Create(tx, u)
			tx.Commit()

			// 断言结果
			assert.Equal(t, nil, err)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
	})
}

func TestGetUserInfoByID(t *testing.T) {
	convey.Convey("Test DBUser GetUserInfoByID()", t, func() {
		db, mock, dbUser := setupUserTest(t)
		defer db.Close()

		convey.Convey("数据库错误", func() {
			mock.ExpectQuery("select").WithArgs("dbError").WillReturnError(errDatabase)

			_, _, err := dbUser.GetUserInfoByID("dbError")

			assert.Equal(t, fmt.Errorf("dbaccess: GetUserInfoByID error: %w", errDatabase), err)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
		convey.Convey("数据不存在", func() {
			mock.ExpectQuery("select").WithArgs("NonExistUser").WillReturnError(sql.ErrNoRows)

			_, exists, err := dbUser.GetUserInfoByID("NonExistUser")

			assert.Equal(t, nil, err)
			assert.Equal(t, false, exists)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
		convey.Convey("数据存在，读取成功", func() {
			u := &interfaces.User{
				ID:   "id",
				Name: "tom",
			}
			rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(u.ID, u.Name)
			mock.ExpectQuery("select").WithArgs(u.ID).WillReturnRows(rows)

			userInfo, exists, err := dbUser.GetUserInfoByID(u.ID)

			assert.Equal(t, nil, err)
			assert.Equal(t, true, exists)
			assert.Equal(t, u, userInfo)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
	})
}

func TestGetUserInfoByName(t *testing.T) {
	convey.Convey("Test DBUser GetUserInfoByName()", t, func() {
		db, mock, dbUser := setupUserTest(t)
		defer db.Close()

		convey.Convey("数据不存在", func() {
			mock.ExpectQuery("select").WithArgs("NonExistUser").WillReturnError(sql.ErrNoRows)

			_, exists, err := dbUser.GetUserInfoByName("NonExistUser")

			assert.Equal(t, nil, err)
			assert.Equal(t, false, exists)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
		convey.Convey("数据库错误", func() {
			mock.ExpectQuery("select").WithArgs("dbError").WillReturnError(errDatabase)

			_, _, err := dbUser.GetUserInfoByName("dbError")

			assert.Equal(t, fmt.Errorf("dbaccess: GetUserInfoByName error: %w", errDatabase), err)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
		convey.Convey("数据存在，读取成功", func() {
			u := &interfaces.User{
				ID:       "id",
				Name:     "tom",
				Password: "12345678",
			}
			rows := sqlmock.NewRows([]string{"id", "name", "password"}).AddRow(u.ID, u.Name, u.Password)
			mock.ExpectQuery("select").WithArgs(u.Name).WillReturnRows(rows)

			userInfo, exists, err := dbUser.GetUserInfoByName(u.Name)

			assert.Equal(t, nil, err)
			assert.Equal(t, true, exists)
			assert.Equal(t, u, userInfo)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
	})
}

func TestUpdateLoginTime(t *testing.T) {
	convey.Convey("Test DBUser UpdateLoginTime()", t, func() {
		db, mock, dbUser := setupUserTest(t)
		defer db.Close()

		convey.Convey("数据更新失败", func() {
			mock.ExpectExec("update").WithArgs(time.Now().Format("2006-01-02 15:04:05"), "id").WillReturnError(errDatabase)

			err := dbUser.UpdateLoginTime("id")

			assert.Equal(t, fmt.Errorf("dbaccess: UpdateLoginTime error: %w", errDatabase), err)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
		convey.Convey("数据更新成功", func() {
			mock.ExpectExec("update").WithArgs(time.Now().Format("2006-01-02 15:04:05"), "id").WillReturnResult(sqlmock.NewResult(1, 1))

			err := dbUser.UpdateLoginTime("id")

			assert.Equal(t, nil, err)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
	})
}
