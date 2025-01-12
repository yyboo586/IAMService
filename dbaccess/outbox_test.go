package dbaccess

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofrs/uuid"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/logUtils"
)

func setupOutboxTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, interfaces.DBOutbox) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	logger, _ := logUtils.NewLogger("debug")

	return db, mock, &outbox{db: db, logger: logger}
}

func TestNewOutbox(t *testing.T) {
	convey.Convey("Test DBOutbox NewOutbox()", t, func() {
		o1 := NewOutbox()
		o2 := NewOutbox()
		assert.Equal(t, o1, o2)
	})
}

func TestAdd(t *testing.T) {
	convey.Convey("Test DBOutbox Add()", t, func() {
		db, mock, dbOutbox := setupOutboxTest(t)
		defer db.Close()

		ctx := context.Background()
		msg := &interfaces.OutboxMessage{
			ID:     "test-id",
			Op:     interfaces.UserCreatedMQ,
			Msg:    []byte("test message"),
			Status: interfaces.OutboxMessageStatusUnhandled,
		}
		convey.Convey("数据库错误", func() {
			// 设置期望
			mock.ExpectBegin()
			mock.ExpectExec("insert into t_outbox").
				WithArgs(msg.ID, int(msg.Op), msg.Msg, msg.Status).
				WillReturnError(errDatabase)
			mock.ExpectRollback()

			// 开启事务并执行
			tx, _ := db.Begin()
			err := dbOutbox.Add(ctx, tx, msg)
			tx.Rollback()

			// 断言结果
			assert.Equal(t, fmt.Errorf("dbaccess: failed to add outbox message: %w", errDatabase), err)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
		convey.Convey("数据插入成功", func() {
			// 设置期望
			mock.ExpectBegin()
			mock.ExpectExec("insert into t_outbox").
				WithArgs(msg.ID, int(msg.Op), msg.Msg, msg.Status).
				WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectCommit()

			// 开启事务并执行
			tx, _ := db.Begin()
			err := dbOutbox.Add(ctx, tx, msg)
			tx.Commit()

			// 断言结果
			assert.Equal(t, nil, err)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
	})
}

func TestGet(t *testing.T) {
	convey.Convey("Test DBOutbox Get()", t, func() {
		db, mock, dbOutbox := setupOutboxTest(t)
		defer db.Close()

		ctx := context.Background()

		convey.Convey("数据不存在", func() {
			mock.ExpectQuery("select").WillReturnError(sql.ErrNoRows)

			_, exists, err := dbOutbox.Get(ctx, interfaces.OutboxMessageStatusUnhandled)

			assert.Equal(t, nil, err)
			assert.Equal(t, false, exists)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("有未满足的期望: %s", err)
			}
		})

		convey.Convey("数据库错误", func() {
			mock.ExpectQuery("select").WillReturnError(errDatabase)

			_, exists, err := dbOutbox.Get(ctx, interfaces.OutboxMessageStatusUnhandled)

			assert.Equal(t, fmt.Errorf("dbaccess: failed to get outbox message: %w", errDatabase), err)
			assert.Equal(t, false, exists)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("有未满足的期望: %s", err)
			}
		})

		convey.Convey("获取数据成功", func() {
			expectedMsg := &interfaces.OutboxMessage{
				ID:  "test-id",
				Op:  1,
				Msg: []byte("test message"),
			}

			rows := sqlmock.NewRows([]string{"id", "op", "msg"}).
				AddRow(expectedMsg.ID, expectedMsg.Op, expectedMsg.Msg)

			mock.ExpectQuery("select").WillReturnRows(rows)

			msg, exists, err := dbOutbox.Get(ctx, interfaces.OutboxMessageStatusUnhandled)

			assert.Equal(t, nil, err)
			assert.Equal(t, true, exists)
			assert.Equal(t, expectedMsg, msg)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("有未满足的期望: %s", err)
			}
		})
	})
}

func TestUpdate(t *testing.T) {
	convey.Convey("Test DBOutbox Update()", t, func() {
		db, mock, dbOutbox := setupOutboxTest(t)
		defer db.Close()

		ctx := context.Background()

		convey.Convey("更新失败", func() {
			mock.ExpectBegin()
			tx, _ := db.Begin()

			mock.ExpectExec("update t_outbox").
				WithArgs(2, "test-id").
				WillReturnError(errDatabase)

			err := dbOutbox.Update(ctx, tx, "test-id", 2)

			assert.Equal(t, fmt.Errorf("dbaccess: failed to update outbox message: %w", errDatabase), err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("有未满足的期望: %s", err)
			}
		})

		convey.Convey("更新成功", func() {
			mock.ExpectBegin()
			tx, _ := db.Begin()

			mock.ExpectExec("update t_outbox").
				WithArgs(2, "test-id").
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := dbOutbox.Update(ctx, tx, "test-id", 2)

			assert.Equal(t, nil, err)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("有未满足的期望: %s", err)
			}
		})
	})
}

func TestDelete(t *testing.T) {
	convey.Convey("Test DBOutbox Delete()", t, func() {
		db, mock, dbOutbox := setupOutboxTest(t)
		defer db.Close()

		ctx := context.Background()

		convey.Convey("删除失败-数据库错误", func() {
			mock.ExpectExec("delete from t_outbox").
				WillReturnError(errDatabase)

			_, err := dbOutbox.Delete(ctx, interfaces.OutboxMessageStatusHandled, 100)

			assert.Equal(t, fmt.Errorf("dbaccess: failed to delete outbox message: %w", errDatabase), err)
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
		convey.Convey("删除成功-当有多条不同状态的记录时", func() {
			// 准备测试数据
			now := time.Now()
			twoDaysAgo := now.Add(-48 * time.Hour)
			halfDayAgo := now.Add(-12 * time.Hour)

			// 模拟插入测试数据
			mock.ExpectExec("insert into t_outbox").
				WithArgs(
					sqlmock.AnyArg(), // id 1
					sqlmock.AnyArg(),
					"msg1",
					int(interfaces.OutboxMessageStatusHandled),
					twoDaysAgo,
					twoDaysAgo,
				).WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectExec("insert into t_outbox").
				WithArgs(
					sqlmock.AnyArg(), // id 2
					sqlmock.AnyArg(),
					"msg2",
					int(interfaces.OutboxMessageStatusHandled),
					twoDaysAgo,
					twoDaysAgo,
				).WillReturnResult(sqlmock.NewResult(2, 1))

			mock.ExpectExec("insert into t_outbox").
				WithArgs(
					sqlmock.AnyArg(), // id 3
					sqlmock.AnyArg(),
					"msg3",
					int(interfaces.OutboxMessageStatusHandled),
					halfDayAgo,
					halfDayAgo,
				).WillReturnResult(sqlmock.NewResult(3, 1))

			// 插入测试数据
			for _, msg := range []struct {
				content   string
				createdAt time.Time
				updatedAt time.Time
			}{
				{"msg1", twoDaysAgo, twoDaysAgo},
				{"msg2", twoDaysAgo, twoDaysAgo},
				{"msg3", halfDayAgo, halfDayAgo},
			} {
				_, err := db.ExecContext(ctx,
					"insert into t_outbox (id, op, msg, status, created_at, updated_at) values (?, ?, ?, ?, ?, ?)",
					uuid.Must(uuid.NewV4()),
					int(interfaces.UserCreatedMQ),
					msg.content,
					int(interfaces.OutboxMessageStatusHandled),
					msg.createdAt,
					msg.updatedAt,
				)
				assert.Equal(t, nil, err)
			}

			// 模拟删除操作，预期只删除状态为Handled且更新时间超过1天的记录
			mock.ExpectExec("delete from t_outbox").
				WithArgs(int(interfaces.OutboxMessageStatusHandled), 100).
				WillReturnResult(sqlmock.NewResult(0, 2)) // 应该删除2条记录（msg1和msg2）

			// 验证剩余记录
			mock.ExpectQuery("select count\\(\\*\\) from t_outbox where status = ?").
				WithArgs(int(interfaces.OutboxMessageStatusHandled)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1)) // 应该剩下1条记录（msg3）

			// 执行删除操作
			rowsAffected, err := dbOutbox.Delete(ctx, interfaces.OutboxMessageStatusHandled, 100)

			// 验证删除结果
			assert.Equal(t, nil, err)
			assert.Equal(t, int64(2), rowsAffected)

			// 验证剩余记录数
			var count int
			err = db.QueryRowContext(ctx, "select count(*) from t_outbox where status = ?",
				int(interfaces.OutboxMessageStatusHandled)).Scan(&count)
			assert.Equal(t, nil, err)
			assert.Equal(t, 1, count) // 验证还剩1条记录（更新时间小于1天的msg3）
			assert.Equal(t, nil, mock.ExpectationsWereMet())
		})
	})
}
