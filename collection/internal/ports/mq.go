package ports
<<<<<<< HEAD

import "gitlab.com/etherealSent/tracklinker-backend/collection/internal/common"

type MQPort interface {
	Publish(exchange, routingKey string, body []byte, replyTo, correlationID string) error
	Subscribe(queueName, consumerTag string) (<-chan common.Message, error)
	// Unsubscribe(topic string) error
	DeclareQueue(queueName string) (common.QueueInfo, error)
}
=======
>>>>>>> 2105d7e486ad49dd4b6965826fe675d1acbea723
