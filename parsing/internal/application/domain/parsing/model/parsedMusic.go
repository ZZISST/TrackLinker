package model

type ParsedMusic struct {
	ID      int    `json:"id" gorm:"-"`
	Title   string `json:"title" gorm:"-"`
	Link    string `json:"link" gorm:"-"`
	Preview string `json:"preview" gorm:"-"`
	Artist  struct {
		Name string `json:"name" gorm:"-"`
		Link string `json:"link" gorm:"-"`
	} `json:"artist" gorm:"-"`
	Album struct {
		Cover string `json:"cover_xl" gorm:"-"`
	} `json:"album" gorm:"-"`
}

type ParsingData struct {
	Title     string
	Artist    string
	DataError error
}

func NewParsedMusic(id int, title string, link string, preview string, artistName string, artistLink string, albumCover string) ParsedMusic {
	return ParsedMusic{
		ID:      id,
		Title:   title,
		Link:    link,
		Preview: preview,
		Artist: struct {
			Name string `json:"name" gorm:"-"`
			Link string `json:"link" gorm:"-"`
		}{
			Name: artistName,
			Link: artistLink,
		},
		Album: struct {
			Cover string `json:"cover_xl" gorm:"-"`
		}{
			Cover: albumCover,
		},
	}
}

func NewParsingData(title string, artist string, err error) ParsingData {
	return ParsingData{
		Title:     title,
		Artist:    artist,
		DataError: err,
	}
}
