package dto

type DeezerSearch struct {
	ID    int `json:"id"`
	Album struct {
		ID int `json:"id"`
	} `json:"album"`
}

type DeezerResponse struct {
	Data []DeezerSearch `json:"data"`
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
	Genres       []DeezerGenre       `json:"genres,omitempty"`
	ReleaseDate  string              `json:"release_date"`
	RecordType   string              `json:"record_type"`
	Contributors []DeezerContributor `json:"contributors,omitempty"`
	Tracks       []DeezerTrack       `json:"tracks,omitempty"`
}

type DeezerGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
