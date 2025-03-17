package handler

type request struct {
	URL string `json:"url"`
}

type response struct {
	Result string `json:"result"`
}
