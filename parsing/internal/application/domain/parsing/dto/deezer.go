package dto

type DeezerSong struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Preview string `json:"preview"`
	Artist  struct {
		Name string `json:"name"`
		Link string `json:"link"`
	} `json:"artist"`
	Album struct {
		Cover string `json:"cover_xl"`
	} `json:"album"`
}

type DeezerResponse struct {
	Data []DeezerSong `json:"data"`
}
