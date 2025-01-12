package logics

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/yyboo586/IAMService/dbaccess"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/logUtils"
)

var (
	oOnce          sync.Once
	outboxInstance *outbox
)

type outbox struct {
	deleteBatchSize int
	deleteInterval  time.Duration // 删除间隔, 避免数据库压力过大
	backupInterval  time.Duration // 消息推送失败，休眠间隔
	msgChan         chan struct{} // 消息通道
	dbPool          *sql.DB
	logger          *logUtils.Logger
	dbOutbox        interfaces.DBOutbox
	handler         map[interfaces.OutboxBussinessType]interfaces.OutboxHandler
}

func NewOutbox() interfaces.LogicsOutbox {
	oOnce.Do(func() {
		outboxInstance = &outbox{
			deleteBatchSize: 100,
			deleteInterval:  500 * time.Millisecond,
			backupInterval:  5 * time.Second,
			msgChan:         make(chan struct{}, 1),
			dbPool:          dbPoolInstance,
			logger:          loggerInstance,
			dbOutbox:        dbaccess.NewOutbox(),
			handler:         make(map[interfaces.OutboxBussinessType]interfaces.OutboxHandler),
		}

		// 启动推送线程
		go outboxInstance.pushWorker()
		go outboxInstance.deleteWorker()
	})

	return outboxInstance
}

func (o *outbox) RegisterHandler(op interfaces.OutboxBussinessType, handler interfaces.OutboxHandler) {
	if _, ok := o.handler[op]; ok {
		panic(fmt.Sprintf("outbox: handler already registered for op: %d", op))
	}
	o.handler[op] = handler
}

func (o *outbox) AddMessage(ctx context.Context, tx *sql.Tx, op interfaces.OutboxBussinessType, data []byte) error {
	msg := &interfaces.OutboxMessage{
		ID:     uuid.Must(uuid.NewV4()).String(),
		Op:     op,
		Msg:    data,
		Status: interfaces.OutboxMessageStatusUnhandled,
	}

	if err := o.dbOutbox.Add(ctx, tx, msg); err != nil {
		return err
	}

	// 非阻塞方式发送通知，通知推送线程有新消息
	select {
	case o.msgChan <- struct{}{}:
	default:
	}
	return nil
}

func (o *outbox) pushWorker() {
	ctx := context.Background()
	for {
		msg, exists, err := o.dbOutbox.Get(ctx, interfaces.OutboxMessageStatusUnhandled)
		if err != nil {
			o.logger.Error("outbox: failed to get outbox message", err)
			continue
		}
		if !exists {
			select {
			case <-o.msgChan:
				continue
			case <-time.After(o.backupInterval):
				o.logger.Debug("outbox: push goroutine: no message to handle")
				continue
			}
		}

		if err := o.messageReply(ctx, msg); err != nil {
			o.logger.Error("outbox: failed to reply message", err)
			time.Sleep(o.backupInterval)
		}
	}
}

func (o *outbox) messageReply(ctx context.Context, msg *interfaces.OutboxMessage) (err error) {
	return withTransaction(o.dbPool, func(tx *sql.Tx) error {
		handler, ok := o.handler[msg.Op]
		if !ok {
			return fmt.Errorf("outbox: handler not found for op: %d", msg.Op)
		}

		if err := handler(ctx, msg); err != nil {
			return fmt.Errorf("outbox: handler failed to handle message: %w", err)
		}

		if err := o.dbOutbox.Update(ctx, tx, msg.ID, interfaces.OutboxMessageStatusHandled); err != nil {
			return fmt.Errorf("outbox: failed to update outbox message: %w", err)
		}

		return nil
	})
}

func (o *outbox) deleteWorker() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// 立即执行一次
	o.doDelete()

	// 每隔一段时间执行一次
	for range ticker.C {
		o.doDelete()
	}
}

func (o *outbox) doDelete() {
	o.logger.Info("outbox: delete goroutine start")

	var totalRowsAffected int64
	ctx := context.Background()
	for {
		rowsAffected, err := o.dbOutbox.Delete(ctx, interfaces.OutboxMessageStatusHandled, o.deleteBatchSize)
		if err != nil {
			o.logger.Error("outbox: failed to delete outbox message", err)
			break
		}
		if rowsAffected == 0 {
			break
		}
		totalRowsAffected += rowsAffected

		// 删除成功后，休眠一段时间，防止频繁删除
		time.Sleep(o.deleteInterval)
	}
	o.logger.Infof("outbox: delete goroutine end, deleted %d outbox messages", totalRowsAffected)
}
