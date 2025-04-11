package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/dto"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/model"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/repository"
	"gorm.io/gorm"
)

type IListeningService interface {
	MakeListening(ctx context.Context, userID, name, artists string, timeStart, timeEnd time.Time) (status *string, err error)
}

type ListeningService struct {
	listeningRepo    repository.IListeningRepository
	addedContentRepo repository.IAddedContentRepository
	collectionRepo   repository.ICollectionRepository
	trackRepo        repository.ITrackRepository
	artistRepo       repository.IArtistRepository
	parsingSvc       IParsingService
}

func NewListeningService(
	listeningRepo repository.IListeningRepository,
	addedContentRepo repository.IAddedContentRepository,
	collectionRepo repository.ICollectionRepository,
	trackRepo repository.ITrackRepository,
	artistRepo repository.IArtistRepository,
	parsingSvc IParsingService,
) IListeningService {
	return &ListeningService{
		listeningRepo:    listeningRepo,
		addedContentRepo: addedContentRepo,
		collectionRepo:   collectionRepo,
		trackRepo:        trackRepo,
		artistRepo:       artistRepo,
		parsingSvc:       parsingSvc,
	}
}

type ParsedTrack struct {
	TrackID string
	AlbumID string
	Status  string
}

func (s *ListeningService) MakeListening(ctx context.Context, userID, name, artists string, timeStart, timeEnd time.Time) (status *string, err error) {

	// track на который указывает listening
	var parsedTrack ParsedTrack

	// Define the transaction handler
	handler := func() error {

		// Форматирование Artists
		formattedArtists := processArtists(artists)
		// Получение Artists
		var artistsID []string
		for _, artistName := range formattedArtists {

			artist, err := s.artistRepo.GetArtistByName(ctx, artistName)
			log.Printf("Artist: %v", artist)

			if err != nil {
				log.Printf("Error: %v", err)

				if errors.Is(err, gorm.ErrRecordNotFound) {
					// запрос в parsing (Если нет артиста -> значит нет трэка)
					// после выполнения запроса все артисты будут созданы
					requestTrack, err := s.parsingSvc.RequestTrackParsing(dto.RequestTrackParsing{
						Name:    name,
						Artists: formattedArtists,
					})
					if err != nil {
						log.Printf("Error: %v", err)
						return err
					} else {
						parsedTrack = ParsedTrack{
							TrackID: requestTrack.TrackID,
							AlbumID: requestTrack.AlbumID,
							Status:  requestTrack.Status,
						}
					}
				}
			}
			artistsID = append(artistsID, artist.ID)
		}

		// Получение Track
		track, err := s.trackRepo.GetTrackByNameAndArtistsID(ctx, name, artistsID)
		log.Printf("Track: %v", track)
		if err != nil {
			log.Printf("Error: %v", err)
			// Если трек не найден, то создаем его
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// запрос в parsing ()
				// FindMusic запрос (не будет создавать артистов заново есть проверка)
				requestTrack, err := s.parsingSvc.RequestTrackParsing(dto.RequestTrackParsing{
					Name:    name,
					Artists: formattedArtists,
				})
				if err != nil {
					log.Printf("Error: %v", err)
					return err
				} else {
					parsedTrack = ParsedTrack{
						TrackID: requestTrack.TrackID,
						AlbumID: requestTrack.AlbumID,
						Status:  requestTrack.Status,
					}
				}
			}
		} else {
			parsedTrack = ParsedTrack{
				TrackID: track.ID,
				AlbumID: track.AlbumID,
				Status:  "nice",
			}
		}

		log.Printf("ParsedTrack: %+v", parsedTrack)

		var collection *model.Collection

		// Получение Collection
		collection, err = s.collectionRepo.GetCollectionByUserID(ctx, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				collection = model.NewCollection(userID)
				s.collectionRepo.CreateCollection(ctx, collection)
			} else {
				return err
			}
		}

		log.Printf("Collection: %v", collection)

		// Получение AddedContent
		addedContent, err := s.addedContentRepo.GetByCollectionIDAndTrackID(ctx, collection.ID, parsedTrack.TrackID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				addedContent = nil
			} else {
				return err
			}
		}

		log.Printf("AddedContent: %v", addedContent)

		// Проверка на существование AddedContent в коллекции
		// Если AddedContent не существует, то создаем его
		if addedContent == nil {

			// Создание AddedContent
			newAddedContent := model.NewAddedContent(collection.ID, parsedTrack.TrackID, "track")
			s.addedContentRepo.Create(ctx, &newAddedContent)

			log.Printf("AddedContent: %v", newAddedContent)

			// Создание Listening
			newListening := model.NewListening(timeStart, timeEnd, newAddedContent.ID)
			s.listeningRepo.Create(ctx, &newListening)
			
			log.Printf("Listening: %v", newListening)
		} else {
			// Если AddedContent существует, то создаем Listening
			newListening := model.NewListening(timeStart, timeEnd, addedContent.ID)
			s.listeningRepo.Create(ctx, &newListening)
			log.Printf("Listening: %v", newListening)
		}

		return nil
	}

	// Start the transaction
	err = s.listeningRepo.WithTransaction(handler)
	if err != nil {
		return nil, err
	}

	return &parsedTrack.Status, nil
}

func processArtists(artists string) []string {
	// Split the string by comma
	artistList := strings.Split(artists, ",")

	// Trim spaces around each artist name
	for i := range artistList {
		artistList[i] = strings.TrimSpace(artistList[i])
	}

	return artistList
}
