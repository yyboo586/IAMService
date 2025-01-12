package drivenadapters

import (
	"context"
	"sync"

	"github.com/nsqio/go-nsq"
	"github.com/yyboo586/IAMService/interfaces"
	"github.com/yyboo586/common/logUtils"
)

var (
	mqOnce     sync.Once
	mqInstance *mq
)

type mq struct {
	producer *nsq.Producer
	logger   *logUtils.Logger
}

func NewMQ() interfaces.DrivenMQ {
	mqOnce.Do(func() {
		mqInstance = &mq{
			producer: mqProducerInstance,
			logger:   loggerInstance,
		}
	})

	return mqInstance
}

func (m *mq) Publish(_ context.Context, topic string, msg []byte) error {
	if err := m.producer.Publish(topic, msg); err != nil {
		return err
	}

	return nil
}
