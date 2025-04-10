package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/common"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

type ICollectionService interface {
	ProcessParsingTrackRequest(ctx context.Context, d common.Message) error
	ListenQueue()
}

type CollectionService struct {
	musicService IMusicService
	collection   ports.CollectionPort
}

func (s *CollectionService) ProcessParsingTrackRequest(ctx context.Context, d common.Message) error {
	var requestData dto.RequestParsing

	if err := json.Unmarshal(d.Body, &requestData); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return nil
	}

	log.Printf("Обработка запроса: Name=%s, Artists=%s", requestData.Name, requestData.Artists)

	musicTrack, err := s.musicService.findTrackByData(ctx, requestData.Name, requestData.Artists)
	if err != nil {
		return err
	}

	responseBody, err := json.Marshal(musicTrack)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	err = s.collection.Publish("", d.ReplyTo, responseBody, "", d.CorrelationId)
	if err != nil {
		return fmt.Errorf("error sending parsing response: %w", err)
	}

	log.Printf("Ответ успешно отправлен для CorrelationId=%s", d.CorrelationId)
	return nil
}

func (s *CollectionService) ListenQueue() {
	go func() {

		msgs, err := s.collection.Subscribe()
		if err != nil {
			log.Printf("Ошибка подписки на очередь: %v", err)
			return
		}

		log.Println("Начало прослушивания очереди parsing.queue")
		for d := range msgs {
			log.Printf("Получено сообщение: CorrelationId=%s, ReplyTo=%s", d.CorrelationId, d.ReplyTo)
			ctx := context.Background()
			// Вызов ProcessParsingTrackRequest для обработки запроса
			err := s.ProcessParsingTrackRequest(ctx, d)
			if err != nil {
				log.Printf("Error processing parsing request: %v", err)
			}
		}
	}()
}

func NewCollectionService(musicService IMusicService, collection ports.CollectionPort) ICollectionService {
	return &CollectionService{
		musicService: musicService,
		collection:   collection,
	}
}
