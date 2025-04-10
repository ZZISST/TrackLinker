package parsing

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/dto"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"
)

type Adapter struct {
	mq ports.MQPort
}

func NewAdapter(mq ports.MQPort) *Adapter {
	return &Adapter{
		mq: mq,
	}
}

func (a *Adapter) RequestTrackParsing(requestData dto.RequestTrackParsing) (track *dto.ResponseTrackParsing, err error) {
	// Генерация уникального идентификатора для корреляции
	correlationId := uuid.New().String()

	// Объявляем временную (callback) очередь для получения ответа
	replyQueue, err := a.mq.DeclareQueue("")
	if err != nil {
		return nil, err
	}

	// Подписываемся на callback очередь
	msgs, err := a.mq.Subscribe(replyQueue.Name, "collection-consumer")
	if err != nil {
		return nil, err
	}

	// Преобразуем requestData в JSON
	body, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	// Публикуем сообщение в очередь Parsing сервиса, устанавливая replyTo и correlationId
	err = a.mq.Publish(
		"",              // exchange (используем дефолтный)
		"parsing.queue", // routing key – имя очереди, на которую подписан Parsing сервис
		body,
		replyQueue.Name,
		correlationId,
	)
	if err != nil {
		return nil, err
	}

	// Ждем ответа с совпадающим correlationId
	timeout := time.After(10 * time.Second)
	for {
		select {
		case d := <-msgs:
			if d.CorrelationId == correlationId {
				var track *dto.ResponseTrackParsing
				if err := json.Unmarshal(d.Body, &track); err != nil {
					return nil, err
				}
				return track, nil
			}
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for parsing response")
		}
	}
}
