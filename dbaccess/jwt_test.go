package dbaccess

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-jose/go-jose/v4"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/logUtils"
)

func setupJWTTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, interfaces.DBJWT) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	logger, _ := logUtils.NewLogger("debug")
	SetLogger(logger)

	return db, mock, &dbJWT{dbPool: db}
}

func TestAddKeySet(t *testing.T) {
	convey.Convey("Test DBJWT AddKeySet()", t, func() {
		db, mock, jwt := setupJWTTest(t)
		defer db.Close()

		keySet := &jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{
				{
					KeyID: "key1",
					Key:   []byte("test-key-1"),
				},
				{
					KeyID: "key2",
					Key:   []byte("test-key-2"),
				},
			},
		}

		convey.Convey("空keySet", func() {
			err := jwt.AddKeySet("test-set", &jose.JSONWebKeySet{})

			assert.Equal(t, nil, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
		convey.Convey("database error", func() {
			mock.ExpectBegin()
			mock.ExpectPrepare("INSERT INTO t_jwt_keys").WillReturnError(errDatabase)
			mock.ExpectRollback()

			err := jwt.AddKeySet("test-set", keySet)

			assert.Equal(t, errDatabase, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
		convey.Convey("添加成功", func() {
			// 添加事务相关的期望
			mock.ExpectBegin()
			mock.ExpectPrepare("INSERT INTO t_jwt_keys").
				ExpectExec().
				WithArgs("key1", sqlmock.AnyArg(), "test-set").
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("INSERT INTO t_jwt_keys").
				WithArgs("key2", sqlmock.AnyArg(), "test-set").
				WillReturnResult(sqlmock.NewResult(2, 1))
			mock.ExpectCommit()

			err := jwt.AddKeySet("test-set", keySet)

			assert.Equal(t, nil, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func TestGetKeySet(t *testing.T) {
	db, mock, jwt := setupJWTTest(t)
	defer db.Close()

	convey.Convey("Test DBJWT GetKeySet()", t, func() {
		convey.Convey("数据库错误", func() {
			mock.ExpectQuery("SELECT").
				WillReturnError(sql.ErrConnDone)

			_, err := jwt.GetKeySet("test-set")

			assert.Equal(t, sql.ErrConnDone, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
		convey.Convey("获取成功", func() {
			rows := sqlmock.NewRows([]string{"data"}).
				AddRow(`{"use":"sig","kty":"oct","kid":"key1","alg":"HS256","k":"Rmt0UGo5SmRERlg3TVNCYURRUGJ1UTdma3BkU1FocG8"}`)
			mock.ExpectQuery("SELECT").
				WithArgs("test-set").
				WillReturnRows(rows)

			result, err := jwt.GetKeySet("test-set")

			assert.Equal(t, nil, err)
			assert.Equal(t, 1, len(result.Keys))
			assert.Equal(t, "key1", result.Keys[0].KeyID)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func TestGetKey(t *testing.T) {
	db, mock, jwt := setupJWTTest(t)
	defer db.Close()

	convey.Convey("Test DBJWT GetKey()", t, func() {
		convey.Convey("数据库错误", func() {
			mock.ExpectQuery("SELECT").
				WillReturnError(sql.ErrConnDone)

			_, err := jwt.GetKey("test-key")

			assert.Equal(t, sql.ErrConnDone, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
		convey.Convey("获取成功", func() {
			rows := sqlmock.NewRows([]string{"data"}).
				AddRow(`{"use":"sig","kty":"oct","kid":"single-key","alg":"HS256","k":"Rmt0UGo5SmRERlg3TVNCYURRUGJ1UTdma3BkU1FocG8"}`)
			mock.ExpectQuery("SELECT data FROM t_jwt_keys WHERE id").
				WithArgs("single-key").
				WillReturnRows(rows)

			key, err := jwt.GetKey("single-key")

			assert.Equal(t, nil, err)
			assert.Equal(t, "single-key", key.KeyID)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func TestAddBlacklist(t *testing.T) {
	db, mock, jwt := setupJWTTest(t)
	defer db.Close()

	convey.Convey("Test DBJWT Blacklist()", t, func() {
		convey.Convey("数据库错误", func() {
			mock.ExpectExec("INSERT INTO t_jwt_blacklist").
				WithArgs("test-token").
				WillReturnError(sql.ErrConnDone)

			err := jwt.AddBlacklist("test-token")

			assert.Equal(t, sql.ErrConnDone, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
		convey.Convey("添加成功", func() {
			mock.ExpectExec("INSERT INTO t_jwt_blacklist").
				WithArgs("test-token").
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := jwt.AddBlacklist("test-token")

			assert.Equal(t, nil, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}

func TestGetBlacklist(t *testing.T) {
	db, mock, jwt := setupJWTTest(t)
	defer db.Close()

	convey.Convey("Test DBJWT GetBlacklist()", t, func() {
		convey.Convey("数据库错误", func() {
			mock.ExpectQuery("SELECT").
				WithArgs("test-token").
				WillReturnError(sql.ErrConnDone)

			exists, err := jwt.GetBlacklist("test-token")

			assert.Equal(t, sql.ErrConnDone, err)
			assert.Equal(t, false, exists)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
		convey.Convey("获取成功", func() {
			rows := sqlmock.NewRows([]string{"count"}).
				AddRow(1)
			mock.ExpectQuery("SELECT").
				WithArgs("test-token").
				WillReturnRows(rows)

			exists, err := jwt.GetBlacklist("test-token")

			assert.Equal(t, nil, err)
			assert.Equal(t, true, exists)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	})
}
