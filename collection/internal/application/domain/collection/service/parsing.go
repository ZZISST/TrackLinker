package service

import (
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/dto"
	"gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"
)

type IParsingService interface {
	RequestTrackParsing(requestData dto.RequestTrackParsing) (track *dto.ResponseTrackParsing, err error)
}

type ParsingService struct {
	parsing ports.ParsingPort
}

func NewParsingService(parsing ports.ParsingPort) IParsingService {
	return &ParsingService{
		parsing: parsing,
	}
}

func (s *ParsingService) RequestTrackParsing(requestData dto.RequestTrackParsing) (track *dto.ResponseTrackParsing, err error) {
	track, err = s.parsing.RequestTrackParsing(requestData)
	if err != nil {
		return nil, err
	}

	return track, nil
}
