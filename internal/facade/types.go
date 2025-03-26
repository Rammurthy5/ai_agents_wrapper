package facade

type ApiResponse struct {
	Source  string `json:"source"`
	Message string `json:"message"`
	Error   error  `json:"error",omitempty`
}

type MergedApiResponse struct {
	Results []ApiResponse `json:"results"`
}
