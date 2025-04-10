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
		return nil
	}

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

	return nil
}

func (s *CollectionService) ListenQueue() {
	go func() error {

		msgs, err := s.collection.Subscribe()
		if err != nil {
			return err
		}

		go func() {
			for d := range msgs {
				ctx := context.Background()
				// Вызов ProcessParsingTrackRequest для обработки запроса
				err := s.ProcessParsingTrackRequest(ctx, d)
				if err != nil {
					log.Printf("Error processing parsing request: %v", err)
				}
			}
		}()

		return nil
	}()
}

func NewCollectionService(musicService IMusicService, collection ports.CollectionPort) ICollectionService {
	return &CollectionService{
		musicService: musicService,
		collection:   collection,
	}
}
