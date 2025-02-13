package model

type URLEntry struct {
	ID          int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
