package dto

type RequestTrackParsing struct {
	Name    string   `json:"name"`
	Artists []string `json:"artists"`
}

type ResponseTrackParsing struct {
	TrackID string `json:"track_id"`
	AlbumID string `json:"album_id"`
	Status  string `json:"status"`
}
