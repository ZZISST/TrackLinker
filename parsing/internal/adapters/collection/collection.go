package collection

import (
	"log"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/common"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type Adapter struct {
	mq ports.MQPort
}

func (a *Adapter) Subscribe() (<-chan common.Message, error) {
	// Объявляем очередь, чтобы убедиться, что она существует
	queue, err := a.mq.DeclareQueue("parsing.queue")
	if err != nil {
	 log.Printf("Ошибка при объявлении очереди parsing.queue: %v", err)
	 return nil, err
	}
	log.Printf("Очередь объявлена: %s", queue.Name)

	msgs, err := a.mq.Subscribe("parsing.queue", "collection-consumer")
	if err != nil {
		log.Printf("Ошибка при подписке: %v", err)
		return nil, err
	}
	log.Println("Подписка на очередь parsing.queue успешно установлена")
	return msgs, nil
}

func (a *Adapter) Publish(exchange, routingKey string, body []byte, replyTo, correlationID string) error {
	log.Printf("Публикация сообщения в очередь %s, replyTo=%s, correlationID=%s", routingKey, replyTo, correlationID)
	return a.mq.Publish(exchange, routingKey, body, replyTo, correlationID)
}

func NewAdapter(mq ports.MQPort) *Adapter {
	return &Adapter{
		mq: mq,
	}
}
