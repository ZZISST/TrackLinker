package model

type MusicRequestType string

const (
	MusicRequestAlbum        MusicRequestType = "album"
	MusicRequestTrack        MusicRequestType = "track"
	MusicRequestTrackInAlbum MusicRequestType = "trackinalbum"
)
