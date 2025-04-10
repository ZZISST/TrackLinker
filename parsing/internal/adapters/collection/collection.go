package collection

import (
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/common"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type Adapter struct {
	mq ports.MQPort
}

func (a *Adapter) Subscribe() (<-chan common.Message, error) {
	msgs, err := a.mq.Subscribe("parsing.queue", "collection-consumer")
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (a *Adapter) Publish(exchange, routingKey string, body []byte, replyTo, correlationID string) error {
	return a.mq.Publish(exchange, routingKey, body, replyTo, correlationID)
}

func NewAdapter(mq ports.MQPort) *Adapter {
	return &Adapter{
		mq: mq,
	}
}
