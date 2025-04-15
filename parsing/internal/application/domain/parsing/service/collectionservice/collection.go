/*
Package collectionservice provides services for communication with the collection microservice.

This package includes methods for processing parsing track requests and listening to message queues.

Authors:

	Vladimir Stabrovskiy @godnesty
*/
package collectionservice

import (
	"context"
	"encoding/json"
	"log"

	"github.com/mdobak/go-xerrors"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/dto"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/service"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/common"
	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/ports"
)

/*
ICollectionService defines the interface for the collection service.

This interface includes methods for processing parsing track requests and listening to message queues.

Author: Vladimir Stabrovskiy @godnesty
*/
type ICollectionService interface {
	ProcessParsingTrackRequest(ctx context.Context, d common.Message) error
	ListenQueue()
}

/*
CollectionService implements the ICollectionService interface.

This service is responsible for handling communication with the collection microservice.

Author: Vladimir Stabrovskiy @godnesty
*/
type CollectionService struct {
	musicService service.IMusicService
	collection   ports.CollectionPort
}

/*
ProcessParsingTrackRequest processes a parsing track request received from the collection microservice.

This method unmarshals the request, finds the track data using the music service, and sends a response back to the collection microservice.

Parameters:
  - ctx (context.Context): The context for the request.
  - d (common.Message): The message containing the parsing track request.

Returns:
  - error: An error if the operation fails.

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *CollectionService) ProcessParsingTrackRequest(ctx context.Context, d common.Message) error {
	var requestData dto.RequestParsing

	if err := json.Unmarshal(d.Body, &requestData); err != nil {
		xerrors.Print(err)
		return err
	}

	musicTrack, err := s.musicService.FindTrackByData(ctx, requestData.Name, requestData.Artists)
	if err != nil {
		return err
	}

	respData, err := json.Marshal(musicTrack)
	if err != nil {
		return err
	}

	err = s.collection.Publish("", d.ReplyTo, respData, "", d.CorrelationId)
	if err != nil {
		return err
	}

	return nil
}

/*
ListenQueue starts listening to the message queue for incoming requests.

This method runs a goroutine that continuously listens to the queue and processes incoming messages.

Parameters:

	None

Returns:

	None

Author: Vladimir Stabrovskiy @godnesty
*/
func (s *CollectionService) ListenQueue() {
	go func() {

		msgs, err := s.collection.Subscribe()
		if err != nil {
			log.Printf("Ошибка подписки на очередь: %v", err)
			return
		}

		for d := range msgs {
			ctx := context.Background()
			err := s.ProcessParsingTrackRequest(ctx, d)
			if err != nil {
				log.Printf("Error processing parsing request: %v", err)
			}
		}
	}()
}

/*
NewCollectionService creates and returns a new instance of CollectionService.

Parameters:
  - musicService [IMusicService]: The music service used for finding track data.
  - collection (ports.CollectionPort): The collection port for publishing messages.

Returns:
  - ICollectionService: A new instance of CollectionService.

Author: Vladimir Stabrovskiy @godnesty
*/
func NewCollectionService(musicService service.IMusicService, collection ports.CollectionPort) ICollectionService {
	return &CollectionService{
		musicService: musicService,
		collection:   collection,
	}
}
