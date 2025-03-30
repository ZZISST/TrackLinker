package modelutils

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"
)

type ILinkFormater interface {
	IdentifyMusicServiceByLink(link *url.URL) (model.MusicService, error)
	IdentifyMusicTypeByLink(link *url.URL) (*model.MusicRequestType, error)
	FormatLinkForDB(mtype model.MusicRequestType, link *url.URL) (*string, error)
	FormatLinkFromDBToOutput(dbFormat string, musicSvc model.MusicService) (*string, error)
}

type LinkFormater struct {
}

// identifyAlbumOrTrackByLink implements ILinkFormater.
func (l *LinkFormater) IdentifyMusicTypeByLink(link *url.URL) (*model.MusicRequestType, error) {

	var musicType model.MusicRequestType

	log.Printf("IdentifyMusicTypeByLink: %s", link)

	musicService, err := l.IdentifyMusicServiceByLink(link)
	if err != nil {
		return nil, fmt.Errorf("unsupported music service")
	}

	linkPath := link.Path

	switch musicService {
	case model.Apple:
		// Apple Music
		// Example track https://music.apple.com/us/song/the-beautiful-people/1440747194
		// Example album https://music.apple.com/us/album/lest-we-forget-the-best-of-marilyn-manson/1440746922
		if strings.Contains(linkPath, "song") {
			musicType = model.MusicRequestTrack
		} else if strings.Contains(linkPath, "album") {
			musicType = model.MusicRequestAlbum
		} else {
			return nil, fmt.Errorf("wrong link")
		}

	case model.Yandex:
		// Yandex Music
		// Example track https://music.yandex.ru/track/33267
		// Example track in album https://music.yandex.ru/album/2754/track/33267
		// Example album https://music.yandex.ru/album/2754
		if strings.Contains(linkPath, "track") && strings.Contains(linkPath, "album") {
			musicType = model.MusicRequestTrackInAlbum
		} else if strings.Contains(linkPath, "album") {
			musicType = model.MusicRequestAlbum
		} else if strings.Contains(linkPath, "track") {
			musicType = model.MusicRequestTrack
		} else {
			return nil, fmt.Errorf("wrong link")
		}

	case model.Spotify:
		// Spotify
		// Example track https://open.spotify.com/track/2aIB1CdRRG7YLBu9hNw9nR
		// Example album https://open.spotify.com/album/7APnZJwayR3lH9aUDPDYQB
		if strings.Contains(linkPath, "track") {
			musicType = model.MusicRequestTrack
		} else if strings.Contains(linkPath, "album") {
			musicType = model.MusicRequestAlbum
		} else {
			return nil, fmt.Errorf("wrong link")
		}

	case model.Deezer:
		// Deezer
		// Example track with region https://www.deezer.com/us/track/87938845
		// Example album with region https://www.deezer.com/us/album/8880231
		// Example track 			 https://www.deezer.com/track/87938845
		// Example album 			 https://www.deezer.com/album/8880231
		if strings.Contains(linkPath, "track") {
			musicType = model.MusicRequestTrack
		} else if strings.Contains(linkPath, "album") {
			musicType = model.MusicRequestAlbum
		} else {
			return nil, fmt.Errorf("wrong link")
		}

	default:
		return nil, fmt.Errorf("unsupported music service")
	}

	return &musicType, nil
}

// FormatLinkForDB implements ILinkFormater.
func (l *LinkFormater) FormatLinkForDB(mtype model.MusicRequestType, link *url.URL) (*string, error) {
	var formattedInputLink string
	var trackID string
	var albumID string

	musicService, err := l.IdentifyMusicServiceByLink(link)
	if err != nil {
		return nil, fmt.Errorf("unsupported music service")
	}

	pathParts := strings.Split(link.Path, "/")

	switch musicService {
	case model.Apple:
		// Apple Music
		// Example track with region https://music.apple.com/us/song/the-beautiful-people/1440747194
		// Example album with region https://music.apple.com/us/album/lest-we-forget-the-best-of-marilyn-manson/1440746922
		// Example track https://music.apple.com/song/the-beautiful-people/1440747194
		// Example album https://music.apple.com/album/lest-we-forget-the-best-of-marilyn-manson/1440746922
		// Link can be with name or not, so it doesnt matter for work, just save id

		// name = pathParts[len(pathParts)-2] // the-beautiful-people

		if mtype == model.MusicRequestTrack {
			trackID = pathParts[len(pathParts)-1] // 1440747194
			formattedInputLink = "/song/" + trackID
		} else if mtype == model.MusicRequestAlbum {
			albumID = pathParts[len(pathParts)-1] // 1440746922
			formattedInputLink = "/album/" + albumID
		} else {
			return nil, fmt.Errorf("wrong link")
		}

		log.Printf(formattedInputLink)

	case model.Yandex:
		// Yandex Music
		// Example track https://music.yandex.ru/track/33267
		// Example track in album https://music.yandex.ru/album/2754/track/33267
		// Example album https://music.yandex.ru/album/2754

		if mtype == model.MusicRequestTrack {
			trackID = pathParts[len(pathParts)-1] // 33267
			formattedInputLink = "/track/" + trackID
		} else if mtype == model.MusicRequestAlbum {
			albumID = pathParts[len(pathParts)-1] // 2754
			formattedInputLink = "/album/" + albumID
		} else if mtype == model.MusicRequestTrackInAlbum {
			albumID = pathParts[len(pathParts)-3] // 2754
			trackID = pathParts[len(pathParts)-1] // 33267
			formattedInputLink = "/album/" + albumID + "/track/" + trackID
		} else {
			return nil, fmt.Errorf("wrong link")
		}

	case model.Spotify:
		// Spotify
		// Example track https://open.spotify.com/track/2aIB1CdRRG7YLBu9hNw9nR
		// Example album https://open.spotify.com/album/7APnZJwayR3lH9aUDPDYQB

		if mtype == model.MusicRequestTrack {
			trackID = pathParts[len(pathParts)-1] // 2aIB1CdRRG7YLBu9hNw9nR
			formattedInputLink = "/track/" + trackID
		} else if mtype == model.MusicRequestAlbum {
			albumID = pathParts[len(pathParts)-1] // 7APnZJwayR3lH9aUDPDYQB
			formattedInputLink = "/album/" + albumID
		} else {
			return nil, fmt.Errorf("wrong link")
		}

	case model.Deezer:
		// Deezer
		// Example track with region https://www.deezer.com/us/track/87938845
		// Example album with region https://www.deezer.com/us/album/8880231
		// Example track 			 https://www.deezer.com/track/87938845
		// Example album 			 https://www.deezer.com/album/8880231

		if mtype == model.MusicRequestTrack {
			trackID = pathParts[len(pathParts)-1] // 87938845
			formattedInputLink = "/track/" + trackID
		} else if mtype == model.MusicRequestAlbum {
			albumID = pathParts[len(pathParts)-1] // 8880231
			formattedInputLink = "/album/" + albumID
		} else {
			return nil, fmt.Errorf("wrong link")
		}

	default:
		return nil, fmt.Errorf("unsupported music service")
	}

	return &formattedInputLink, nil
}

// FormatLinkForOutput implements ILinkFormater.
func (l *LinkFormater) FormatLinkFromDBToOutput(dbFormat string, musicSvc model.MusicService) (*string, error) {
	var formattedOutputLink string

	switch musicSvc {
	case model.Apple:
		formattedOutputLink = "https://" + string(model.AppleCom) + dbFormat
	case model.Yandex:
		formattedOutputLink = "https://" + string(model.YandexCom) + dbFormat
	case model.Spotify:
		formattedOutputLink = "https://" + string(model.OpenSpotifyCom) + dbFormat
	case model.Deezer:
		formattedOutputLink = "https://" + string(model.DeezerLink) + dbFormat
	default:
		return nil, fmt.Errorf("unsupported music service")
	}

	return &formattedOutputLink, nil
}

// IdentifyMusicServiceByLink implements ILinkFormater.
func (l *LinkFormater) IdentifyMusicServiceByLink(link *url.URL) (model.MusicService, error) {
	var streamingService model.MusicService

	host := link.Host

	if host == "" {
		log.Printf("Ошибка: поле Host у URL пустое. URL: %s", link.String())
		return model.Unsupported, fmt.Errorf("URL не содержит хост")
	}

	if strings.HasPrefix(host, "www.") {
		// Remove the "www." prefix
		host = strings.TrimPrefix(link.Host, "www.")
	}

	log.Printf("IdentifyMusicServiceByLink: %s", host)

	// identify music service
	switch {
	case strings.Contains(host, string(model.AppleCom)):
		streamingService = model.Apple
	case strings.Contains(host, string(model.YandexCom)) || strings.Contains(host, string(model.YandexRu)):
		streamingService = model.Yandex
	case strings.Contains(host, string(model.SpotifyCom)) || strings.Contains(host, string(model.OpenSpotifyCom)):
		streamingService = model.Spotify
	case strings.Contains(host, string(model.DeezerLink)) || strings.Contains(host, string(model.DeezerLinkShort)) || strings.Contains(host, string(model.DeezerCom)):
		streamingService = model.Deezer
	default:
		return model.Unsupported, fmt.Errorf("Unsupported service")

	}
	log.Printf("Identified music service: %v", streamingService)

	return streamingService, nil
}

func NewLinkFormater() ILinkFormater {
	return &LinkFormater{}
}
