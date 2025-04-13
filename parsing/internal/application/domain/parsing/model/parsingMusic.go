package model

import "time"

type ParsingTrack struct {
	ID            int             `json:"id" gorm:"-"`
	Title         string          `json:"title" gorm:"-"`
	DurationSec   int             `json:"duration_sec" gorm:"-"`
	Link          string          `json:"link" gorm:"-"`
	TrackPosition int             `json:"track_position" gorm:"-"`
	DiscNumber    int             `json:"disc_number" gorm:"-"`
	ReleaseDate   time.Time       `json:"release_date" gorm:"-"`
	Artists       []ParsingArtist `json:"contributors" gorm:"-"`
	Album         ParsingAlbum    `json:"album" gorm:"-"`
	Preview       string          `json:"preview" gorm:"-"`
}

type ParsingData struct {
	Title     string
	Artist    string
	DataError error
}

type ParsingArtist struct {
	ID    int    `json:"id" gorm:"-"`
	Name  string `json:"name" gorm:"-"`
	Cover string `json:"cover_xl" gorm:"-"`
}

type ParsingAlbum struct {
	ID          int             `json:"id" gorm:"-"`
	Title       string          `json:"title" gorm:"-"`
	Cover       string          `json:"cover_xl" gorm:"-"`
	Genres      []ParsingGenre  `json:"genres" gorm:"-"`
	ReleaseDate time.Time       `json:"release_date" gorm:"-"`
	RecordType  string          `json:"record_type" gorm:"-"`
	Artists     []ParsingArtist `json:"artists" gorm:"-"`
	Tracks      []ParsingTrack  `json:"tracks" gorm:"-"`
}

type ParsingGenre struct {
	ID   int    `json:"id" gorm:"-"`
	Name string `json:"name" gorm:"-"`
}

func NewParsingArtist(id int, name string) ParsingArtist {
	return ParsingArtist{
		ID:   id,
		Name: name,
	}
}

func NewParsingAlbum(id int, title string, cover string) ParsingAlbum {
	return ParsingAlbum{
		ID:    id,
		Title: title,
		Cover: cover,
	}
}

func NewParsingGenre(id int, name string) ParsingGenre {
	return ParsingGenre{
		ID:   id,
		Name: name,
	}
}

func NewParsingTrack(id int, title string, durationSec int, link string, trackPostion int, discNumber int, releaseDate time.Time, artists []ParsingArtist, album ParsingAlbum) ParsingTrack {
	return ParsingTrack{
		ID:            id,
		Title:         title,
		DurationSec:   durationSec,
		Link:          link,
		TrackPosition: trackPostion,
		DiscNumber:    discNumber,
		ReleaseDate:   releaseDate,
		Artists:       artists,
		Album:         album,
	}
}

func NewParsingData(title string, artist string, err error) ParsingData {
	return ParsingData{
		Title:     title,
		Artist:    artist,
		DataError: err,
	}
}
