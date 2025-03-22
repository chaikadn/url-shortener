package handler

// убрать
type request struct {
	URL string `json:"url"`
}

// убрать
type response struct {
	Result string `json:"result"`
}

type requestBatch struct {
	ID          string `json:"correlation_id"`
	OriginalURL string `json:"original_url"`
}

type responseBatch struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}
