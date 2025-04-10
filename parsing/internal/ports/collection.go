package ports

import (
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/common"
)

type CollectionPort interface {
	Subscribe() (<-chan common.Message, error)
	Publish(exchange, routingKey string, body []byte, replyTo, correlationID string) error
}
