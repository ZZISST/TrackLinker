/*
Package service provides implementations for music service interactions.

Authors:

	Vladimir Stabrovskiy @godnesty
	Alexander Subbotin @CAHEK_aka_zzisst
	Nikita Kizhvatkin @djsega1
*/
package musicservice

import "gitlab.com/etherealSent/tracklinker-backend/parsing/internal/application/domain/parsing/model"

/*
IParsingMusicService defines the basic methods required for interacting with a music service.

Methods in this interface typically include operations for retrieving track or album data.

Author: Vladimir Stabrovskiy @godnesty
*/
type IParsingMusicService interface {
	GetTrackData(trackUrl string) (title string, artist string, err error)
	GetAlbumData(trackUrl string) (title string, artist string, err error)
	GetTrackUrl(title, artist string) (trackUrl string, err error)
	GetAlbumUrl(title, artist string) (trackUrl string, err error)
	Service() model.MusicService
}

/*
IFullParsingMusicService extends IParsingMusicService by adding methods for full metadata parsing.

This interface includes additional methods for retrieving detailed track and album data,
as well as constructing URLs for tracks and albums.

Author: Vladimir Stabrovskiy @godnesty
*/
type IFullParsingMusicService interface {
	IParsingMusicService
	GetAllTrackData(title, artist string) (*model.ParsingTrack, error)
	GetAllAlbumData(title, artist string) (*model.ParsingAlbum, error)
}
