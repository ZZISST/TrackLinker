package ports

import "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/common"

type MQPort interface {
	Publish(exchange, routingKey string, body []byte, replyTo, correlationID string) error
	Subscribe(queueName, consumerTag string) (<-chan common.Message, error)
	DeclareQueue(queueName string) (common.QueueInfo, error)
}
