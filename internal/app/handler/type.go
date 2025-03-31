package handler

type request struct {
	URL string `json:"url"`
}

type response struct {
	Result string `json:"result"`
}

// переименовать
type requestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// переименовать
type responseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
