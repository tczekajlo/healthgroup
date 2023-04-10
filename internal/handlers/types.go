package handlers

// ResponseHTTP represents response body
type ResponseHTTP struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
