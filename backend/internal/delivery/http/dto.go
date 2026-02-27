package http

type CreateShortenRequest struct {
	LongURL string `json:"original_url" binding:"required"`
}

type SearchByOriginalURLRequest struct {
	URL string `json:"url" binding:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status string `json:"status"`
}
