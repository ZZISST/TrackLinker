package model

type LinkResult struct {
	URL     string       `json:"url"`
	Service MusicService `json:"service"`
	Error   error        `json:"-"`
}

type LinkData struct {
	Links map[string]LinkResult `json:"links"`
}
