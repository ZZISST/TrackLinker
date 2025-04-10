package mq

import (
	"github.com/streadway/amqp"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/common"
)

type Adapter struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func (a *Adapter) Publish(exchange, routingKey string, body []byte, replyTo, correlationID string) error {
	return a.Channel.Publish(
		exchange,   // exchange (если не используете exchange, можно задать пустую строку)
		routingKey, // routing key (может быть именем очереди)
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			Body:          body,
			ReplyTo:       replyTo,
			CorrelationId: correlationID,
		},
	)
}

func (a *Adapter) Subscribe(queueName, consumerTag string) (<-chan common.Message, error) {
	// Получаем канал amqp.Delivery от AMQP клиента
	deliveryChan, err := a.Channel.Consume(
		queueName,   // очередь
		consumerTag, // consumer tag
		true,        // autoAck
		false,       // exclusive
		false,       // noLocal
		false,       // noWait
		nil,         // args
	)
	if err != nil {
		return nil, err
	}

	// Создаем выходной канал с типом Message
	out := make(chan common.Message)

	// Запускаем горутину для маппинга сообщений
	go func() {
		for d := range deliveryChan {
			msg := mapDeliveryToMessage(d)
			out <- msg
		}
		close(out)
	}()

	return out, nil
}

func NewAdapter(amqpURL string) (*Adapter, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &Adapter{
		Connection: conn,
		Channel:    ch,
	}, nil
}
