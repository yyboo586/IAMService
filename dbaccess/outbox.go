package dbaccess

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/logUtils"
)

var (
	oOnce     sync.Once
	oInstance interfaces.DBOutbox
)

type outbox struct {
	db     *sql.DB
	logger *logUtils.Logger
}

func NewOutbox() interfaces.DBOutbox {
	oOnce.Do(func() {
		oInstance = &outbox{
			db:     dbPool,
			logger: loggerInstance,
		}
	})

	return oInstance
}

func (o *outbox) Add(ctx context.Context, tx *sql.Tx, msg *interfaces.OutboxMessage) error {
	sqlStr := "insert into t_outbox(id, op, msg, status) values(?, ?, ?, ?)"

	_, err := tx.ExecContext(ctx, sqlStr, msg.ID, int(msg.Op), msg.Msg, msg.Status)
	if err != nil {
		return fmt.Errorf("dbaccess: failed to add outbox message: %w", err)
	}

	return nil
}

func (o *outbox) Get(ctx context.Context, status interfaces.OutboxMessageStatus) (msg *interfaces.OutboxMessage, exists bool, err error) {
	sqlStr := "select id, op, msg from t_outbox where status = ? order by created_at desc limit 1"

	msg = &interfaces.OutboxMessage{}
	if err := o.db.QueryRowContext(ctx, sqlStr, int(status)).Scan(&msg.ID, &msg.Op, &msg.Msg); err != nil {
		if err == sql.ErrNoRows {
			return msg, false, nil
		}
		return msg, false, fmt.Errorf("dbaccess: failed to get outbox message: %w", err)
	}

	return msg, true, nil
}

func (o *outbox) Update(ctx context.Context, tx *sql.Tx, id string, status interfaces.OutboxMessageStatus) error {
	sqlStr := "update t_outbox set status = ? where id = ?"

	_, err := tx.ExecContext(ctx, sqlStr, int(status), id)
	if err != nil {
		return fmt.Errorf("dbaccess: failed to update outbox message: %w", err)
	}

	return nil
}

func (o *outbox) Delete(ctx context.Context, status interfaces.OutboxMessageStatus, batchSize int) (rowsAffected int64, err error) {
	sqlStr := "delete from t_outbox where status = ? and updated_at < now() - interval 1 day limit ?"

	result, err := o.db.ExecContext(ctx, sqlStr, int(status), batchSize)
	if err != nil {
		return 0, fmt.Errorf("dbaccess: failed to delete outbox message: %w", err)
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("dbaccess: failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
