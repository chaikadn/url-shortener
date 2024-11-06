package handler

type request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}
