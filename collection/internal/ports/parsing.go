package ports

import "gitlab.com/etherealSent/tracklinker-backend/collection/internal/application/domain/collection/dto"

type ParsingPort interface {
	RequestTrackParsing(requestData dto.RequestTrackParsing) (track *dto.ResponseTrackParsing, err error)
}
