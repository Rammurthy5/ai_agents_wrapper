package facade

type ApiResponse struct {
	Source  string `json:"source"`
	Message string `json:"message"`
	Error   string `json:"error",omitempty`
}

type MergedApiResponse struct {
	Results []ApiResponse `json:"results"`
}
