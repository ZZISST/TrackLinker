package mq

import (
	"github.com/streadway/amqp"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/common"
)

func mapDeliveryToMessage(d amqp.Delivery) common.Message {
	return common.Message{
		ContentType:     d.ContentType,
		ContentEncoding: d.ContentEncoding,
		DeliveryMode:    d.DeliveryMode,
		Priority:        d.Priority,
		CorrelationId:   d.CorrelationId,
		ReplyTo:         d.ReplyTo,
		Expiration:      d.Expiration,
		MessageId:       d.MessageId,
		Timestamp:       d.Timestamp,
		Type:            d.Type,
		UserId:          d.UserId,
		AppId:           d.AppId,
		ConsumerTag:     d.ConsumerTag,
		MessageCount:    d.MessageCount,
		DeliveryTag:     d.DeliveryTag,
		Redelivered:     d.Redelivered,
		Exchange:        d.Exchange,
		RoutingKey:      d.RoutingKey,
		Body:            d.Body,
	}
}

func mapAmqpQueueToQueueInfo(q amqp.Queue) common.QueueInfo {
	return common.QueueInfo{
		Name:      q.Name,
		Messages:  q.Messages,
		Consumers: q.Consumers,
	}
}
