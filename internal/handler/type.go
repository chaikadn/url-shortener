package handler

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserURLsResponse struct {
	UserID int       `json:"user_id"`
	URLs   []URLPair `json:"urls"`
}

type URLPair struct {
	OriginalURL string `json:"original"`
	ShortURL    string `json:"short"`
}
