package dbaccess

import (
	"ServiceA/interfaces"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

var (
	errDatabase = errors.New("database error")
)

func newUser(db *sql.DB) *user {
	return &user{
		db: db,
	}
}

func TestCreate(t *testing.T) {
	Convey("Test DBUser Create()", t, func() {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		dbUser := newUser(db)

		Convey("数据插入成功", func() {
			u := &interfaces.User{
				ID:       "id",
				Name:     "test",
				Password: "123456",
			}
			mock.ExpectExec("insert").WithArgs(u.ID, u.Name, u.Password).WillReturnResult(sqlmock.NewResult(1, 1))

			err = dbUser.Create(u)

			assert.Equal(t, nil, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})

}

func TestFetchByName(t *testing.T) {
	Convey("Test DBUser FetchByName()", t, func() {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		dbUser := newUser(db)

		Convey("数据不存在", func() {
			mock.ExpectQuery("select").WithArgs("NonExistUser").WillReturnError(sql.ErrNoRows)

			_, exists, err := dbUser.FetchByName("NonExistUser")

			assert.Equal(t, nil, err)
			assert.Equal(t, false, exists)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("数据库错误", func() {
			mock.ExpectQuery("select").WithArgs("dbError").WillReturnError(errDatabase)

			_, _, err := dbUser.FetchByName("dbError")

			assert.Equal(t, errDatabase, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("数据存在，读取成功", func() {
			u := &interfaces.User{
				ID:       "id",
				Name:     "tom",
				Password: "12345678",
			}
			rows := sqlmock.NewRows([]string{"id", "name", "password"}).AddRow(u.ID, u.Name, u.Password)
			mock.ExpectQuery("select").WithArgs(u.Name).WillReturnRows(rows)

			_, exists, err := dbUser.FetchByName(u.Name)

			assert.Equal(t, nil, err)
			assert.Equal(t, true, exists)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func TestUpdateLoginTime(t *testing.T) {
	Convey("Test DBUser UpdateLoginTime()", t, func() {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
		}
		defer db.Close()

		dbUser := newUser(db)

		Convey("数据更新失败", func() {
			mock.ExpectExec("update").WithArgs(time.Now().Format("2006-01-02 15:04:05"), "id").WillReturnError(errDatabase)

			err = dbUser.UpdateLoginTime("id")

			assert.Equal(t, errDatabase, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})

		Convey("数据更新成功", func() {
			mock.ExpectExec("update").WithArgs(time.Now().Format("2006-01-02 15:04:05"), "id").WillReturnResult(sqlmock.NewResult(1, 1))

			err = dbUser.UpdateLoginTime("id")

			assert.Equal(t, nil, err)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
