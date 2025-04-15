package dto

type DeezerTrackSearch struct {
	ID    int    `json:"id"`
	Link  string `json:"link"`
	Album struct {
		ID int `json:"id"`
	} `json:"album"`
}

type DeezerTrackResponse struct {
	Data []DeezerTrackSearch `json:"data"`
}

type DeezerAlbumnSearch struct {
	ID    int `json:"id"`
	Album struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
		Type  string `json:"type"`
	} `json:"album"`
}

type DeezerAlbumResponse struct {
	Data []DeezerAlbumnSearch `json:"data"`
}

type DeezerContributor struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	PictureXl string `json:"picture_xl"`
}

type DeezerAlbum struct {
	ID           int                 `json:"id"`
	Title        string              `json:"title"`
	Cover        string              `json:"cover_xl"`
	Genres       DeezerGenres        `json:"genres,omitempty"` // Update this field
	ReleaseDate  string              `json:"release_date"`
	RecordType   string              `json:"record_type"`
	Contributors []DeezerContributor `json:"contributors,omitempty"`
	Tracks       DeezerTrackData     `json:"tracks,omitempty"` // Update this field
}

type DeezerGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type DeezerGenres struct {
	Data []DeezerGenre `json:"data"` // Match the "data" field in the JSON
}

type DeezerTrackData struct {
	Data []DeezerTrack `json:"data"` // Match the "data" field in the JSON
}

type DeezerTrack struct {
	ID            int                 `json:"id"`
	Title         string              `json:"title"`
	Duration      int                 `json:"duration"`
	Link          string              `json:"link"`
	TrackPosition int                 `json:"track_position"`
	DiscNumber    int                 `json:"disc_number"`
	ReleaseDate   string              `json:"release_date"`
	Preview       string              `json:"preview"`
	Contributors  []DeezerContributor `json:"contributors"`
	Album         DeezerAlbum         `json:"album"`
}
